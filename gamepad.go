package main

import "github.com/holoplot/go-evdev"

type Gamepad struct {
	dev *evdev.InputDevice
}

func CreateGamepad() (*Gamepad, error) {
	dev, err := evdev.CreateDevice("hotas-to-gamepad virtual controller", evdev.InputID{
		BusType: evdev.BUS_USB,
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
		},
		evdev.EV_KEY: {
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
			// d-pad
			evdev.BTN_DPAD_UP,
			evdev.BTN_DPAD_DOWN,
			evdev.BTN_DPAD_LEFT,
			evdev.BTN_DPAD_RIGHT,
		},
	})
	if err != nil {
		return nil, err
	}
	return &Gamepad{
		dev: dev,
	}, nil
}

func (g *Gamepad) Send(event *evdev.InputEvent) error {
	return g.dev.WriteOne(event)
}

func (g *Gamepad) Close() error {
	return g.dev.Close()
}
