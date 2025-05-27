package mapping

import "github.com/holoplot/go-evdev"

type ModeFunction func(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32

type ControlInfo struct {
	Type    evdev.EvType
	Code    evdev.EvCode
	Maximum int32
	Minimum int32
}

type tableRecord struct {
	Type evdev.EvType
	Code evdev.EvCode
	Mode ModeFunction
}
