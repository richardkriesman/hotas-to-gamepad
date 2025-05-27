package device

import (
	"crypto/rand"
	"fmt"
	"github.com/holoplot/go-evdev"
)

type OutputDevice struct {
	input  InputDevice
	outDev *evdev.InputDevice
}

func Create(config Config) (*OutputDevice, error) {
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
	absInfos, err := inDev.AbsInfos()
	if err != nil {
		return nil, err
	}

	return &OutputDevice{
		input: InputDevice{
			absInfos:        absInfos,
			inDev:           inDev,
			shouldTerminate: false,
			persistentID:    createPersistentID(inDev),
		},
		outDev: outDev,
	}, nil
}

func (d *OutputDevice) Close() error {
	if err := d.input.Close(); err != nil {
		return err
	}
	return d.outDev.Close()
}

func (d *OutputDevice) Send(event *evdev.InputEvent) error {
	return d.outDev.WriteOne(event)
}
