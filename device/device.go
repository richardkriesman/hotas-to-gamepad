package device

import (
	"github.com/holoplot/go-evdev"
)

type Device struct {
	absInfos        map[evdev.EvCode]evdev.AbsInfo
	dev             *evdev.InputDevice
	shouldTerminate bool
	persistentID    PersistentID
}

type Channels struct {
	Errors chan error
	Events chan Event
}

type Frame struct {
	Events []*Event
}

type Event struct {
	evdev.InputEvent
	Device   *Device
	Frame    *Frame
	Sequence uint
}

func Create() (*Device, error) {
	dev, err := evdev.CreateDevice("hotas-to-gamepad virtual controller", evdev.InputID{
		BusType: evdev.BUS_VIRTUAL,
		Vendor:  0,
		Product: 0,
		Version: 1,
	}, map[evdev.EvType][]evdev.EvCode{
		evdev.EV_ABS: {
			// left stick
			evdev.ABS_X,
			evdev.ABS_Y,
			// right stick
			evdev.ABS_RX,
			evdev.ABS_RY,
			// triggers
			evdev.ABS_HAT2Y, // lower left
			evdev.ABS_HAT2X, // lower right
			// d-pad (analog form)
			evdev.ABS_HAT0X,
			evdev.ABS_HAT0Y,
		},
		evdev.EV_KEY: {
			// "xbox" or another center button
			evdev.BTN_MODE,
			// action buttons (a, b, x, y)
			evdev.BTN_NORTH,
			evdev.BTN_SOUTH, // also BTN_GAMEPAD - used for gamepad detection
			evdev.BTN_EAST,
			evdev.BTN_WEST,
			// thumbstick center click
			evdev.BTN_THUMBL,
			evdev.BTN_THUMBR,
			// start/select buttons
			evdev.BTN_START,
			evdev.BTN_SELECT,
			// shoulder buttons
			evdev.BTN_TL, // upper left
			evdev.BTN_TR, // upper right
			// d-pad (digital button form)
			//evdev.BTN_DPAD_UP,
			//evdev.BTN_DPAD_DOWN,
			//evdev.BTN_DPAD_LEFT,
			//evdev.BTN_DPAD_RIGHT,
		},
	})
	if err != nil {
		return nil, err
	}
	// FIXME: doesn't work on created devices - may need to introduce custom ranges
	//absInfo, err := dev.AbsInfos()
	//if err != nil {
	//	return nil, err
	//}
	return &Device{
		//absInfos:     absInfo,
		dev:          dev,
		persistentID: createPersistentID(dev),
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
		dev:          dev,
		persistentID: createPersistentID(dev),
	}
	return &inputDevice, nil
}

func (d *Device) AbsInfos() map[evdev.EvCode]evdev.AbsInfo {
	return d.absInfos
}

func (d *Device) Close() error {
	d.shouldTerminate = true
	err := d.dev.Ungrab()
	if err != nil {
		return err
	}
	return d.dev.Close()
}

func (d *Device) Name() string {
	name, err := d.dev.Name()
	if err != nil {
		name = "unknown"
	}
	return name
}

func (d *Device) Listen() *Channels {
	channels := Channels{
		Errors: make(chan error),
		Events: make(chan Event),
	}
	go d.listen(&channels)
	return &channels
}

func (d *Device) PersistentID() PersistentID {
	return d.persistentID
}

func (d *Device) Raw() *evdev.InputDevice {
	return d.dev
}

func (d *Device) Send(event *evdev.InputEvent) error {
	return d.dev.WriteOne(event)
}

func (d *Device) listen(channels *Channels) {
	for !d.shouldTerminate {
		// wait for input events and build frame
		frame := Frame{}
		seq := uint(0)
		for {
			ev, err := d.dev.ReadOne()
			if err != nil {
				channels.Errors <- err
				continue
			}
			frame.Events = append(frame.Events, &Event{
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
