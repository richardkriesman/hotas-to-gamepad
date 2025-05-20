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
	Events chan InputEvent
}

type Frame struct {
	Events []*InputEvent
}

type InputEvent struct {
	evdev.InputEvent
	Device   *Device
	Frame    *Frame
	Sequence uint
}
