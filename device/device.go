package device

import (
	"crypto/rand"
	"fmt"
	"github.com/holoplot/go-evdev"
)

type Device struct {
	absInfos        map[evdev.EvCode]evdev.AbsInfo
	inDev           *evdev.InputDevice
	outDev          *evdev.InputDevice
	shouldTerminate bool
	persistentID    PersistentID
}

func Create(config Config) (*Device, error) {
	outId := rand.Text()[:6] // needed to identify the virtual device in evdev
	outName := fmt.Sprintf("hotas-to-gamepad virtual controller %s", outId)

	// build capabilities and uinput_user_device params
	capabilities := make(map[evdev.EvType][]evdev.EvCode)
	absParams := evdev.UserDeviceAbsParams{}
	for code, params := range config.Axes {
		capabilities[evdev.EV_ABS] = append(capabilities[evdev.EV_ABS], code)
		absParams.Absmin[code] = params.Minimum
		absParams.Absmax[code] = params.Maximum
		absParams.Absfuzz[code] = params.Fuzz
		absParams.Absflat[code] = params.Flat
	}
	for _, code := range config.Buttons {
		capabilities[evdev.EV_KEY] = append(capabilities[evdev.EV_KEY], code)
	}

	// create a virtual device
	outDev, err := evdev.CreateDeviceWithAbsParams(outName, evdev.InputID{
		BusType: evdev.BUS_VIRTUAL,
		Vendor:  0,
		Product: 0,
		Version: 1,
	}, capabilities, absParams)
	if err != nil {
		return nil, err
	}

	// find and open the created device in evdev
	var outPath string
	paths, err := evdev.ListDevicePaths()
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		if path.Name == outName {
			outPath = path.Path
			break
		}
	}
	inDev, err := evdev.Open(outPath)
	if err != nil {
		return nil, err
	}

	// get abs info
	absInfo, err := inDev.AbsInfos()
	if err != nil {
		return nil, err
	}

	return &Device{
		absInfos:     absInfo,
		inDev:        inDev,
		outDev:       outDev,
		persistentID: createPersistentID(inDev),
	}, nil
}

func Open(path string) (*Device, error) {
	dev, err := evdev.Open(path)
	if err != nil {
		return nil, err
	}
	err = dev.Grab()
	if err != nil {
		return nil, err
	}
	absInfos, err := dev.AbsInfos()
	if err != nil {
		return nil, err
	}
	inputDevice := Device{
		absInfos:     absInfos,
		inDev:        dev,
		persistentID: createPersistentID(dev),
	}
	return &inputDevice, nil
}

func (d *Device) AbsInfos() map[evdev.EvCode]evdev.AbsInfo {
	return d.absInfos
}

func (d *Device) Close() error {
	d.shouldTerminate = true
	err := d.inDev.Ungrab()
	if err != nil {
		return err
	}
	return d.inDev.Close()
}

func (d *Device) Name() string {
	name, err := d.inDev.Name()
	if err != nil {
		name = "unknown"
	}
	return name
}

func (d *Device) Listen() *Channels {
	channels := Channels{
		Errors: make(chan error),
		Events: make(chan InputEvent),
	}
	go d.listen(&channels)
	return &channels
}

func (d *Device) PersistentID() PersistentID {
	return d.persistentID
}

func (d *Device) Raw() *evdev.InputDevice {
	return d.inDev
}

func (d *Device) Send(event *evdev.InputEvent) error {
	return d.outDev.WriteOne(event)
}

func (d *Device) listen(channels *Channels) {
	for !d.shouldTerminate {
		// wait for input events and build frame
		frame := Frame{}
		seq := uint(0)
		for {
			ev, err := d.inDev.ReadOne()
			if err != nil {
				channels.Errors <- err
				continue
			}
			frame.Events = append(frame.Events, &InputEvent{
				Device:     d,
				InputEvent: *ev,
				Frame:      &frame,
				Sequence:   seq,
			})
			if ev.Type == evdev.EV_SYN {
				break
			}
			seq++
		}

		// unwind the frame, emitting each event individually
		for _, event := range frame.Events {
			channels.Events <- *event
		}
	}
}
