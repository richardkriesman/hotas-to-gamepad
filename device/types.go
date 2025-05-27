package device

import "github.com/holoplot/go-evdev"

type AxisParams struct {
	Minimum int32
	Maximum int32
	Fuzz    int32
	Flat    int32
}

type Config struct {
	Axes    map[evdev.EvCode]AxisParams
	Buttons []evdev.EvCode
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
	Device   *InputDevice
	Frame    *Frame
	Sequence uint
}
