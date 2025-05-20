package device

import (
	"github.com/holoplot/go-evdev"
)

var Gamepad = Config{
	Axes: map[evdev.EvCode]AxisParams{
		// left stick
		evdev.ABS_X: {
			Minimum: -32767,
			Maximum: 32767,
		},
		evdev.ABS_Y: {
			Minimum: -32767,
			Maximum: 32767,
		},
		// right stick
		evdev.ABS_RX: {
			Minimum: -32767,
			Maximum: 32767,
		},
		evdev.ABS_RY: {
			Minimum: -32767,
			Maximum: 32767,
		},
		// left trigger
		evdev.ABS_Z: {
			Minimum: 0,
			Maximum: 1023,
		},
		// right trigger
		evdev.ABS_RZ: {
			Minimum: 0,
			Maximum: 1023,
		},
		// d-pad (analog form)
		evdev.ABS_HAT0X: {
			Minimum: -1,
			Maximum: 1,
		},
		evdev.ABS_HAT0Y: {
			Minimum: -1,
			Maximum: 1,
		},
	},
	Buttons: []evdev.EvCode{
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
}
