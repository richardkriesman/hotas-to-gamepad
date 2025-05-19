package mapping

import (
	"github.com/holoplot/go-evdev"
	"hotas-to-gamepad/device"
)

type Table map[device.PersistentID]map[evdev.EvType]map[evdev.EvCode]TableRecord

type TableRecord struct {
	Type evdev.EvType
	Code evdev.EvCode
	Mode ModeFunction
}

type ModeFunction func(inputValue int32, inputInfo ControlInfo, outputInfo ControlInfo) int32

type ControlInfo struct {
	Type    evdev.EvType
	Code    evdev.EvCode
	Maximum int32
	Minimum int32
}

func (t Table) Remap(event *device.Event) (*evdev.InputEvent, bool) {
	// get remapping data
	record, ok := t[event.Device.PersistentID()][event.Type][event.Code]
	if !ok {
		return nil, false
	}

	// build input control info
	inputInfo := ControlInfo{
		Type:    event.Type,
		Code:    event.Code,
		Maximum: 1,
		Minimum: 0,
	}
	if event.Type == evdev.EV_ABS {
		inputInfo.Maximum = event.Device.AbsInfos()[event.Code].Maximum
		inputInfo.Minimum = event.Device.AbsInfos()[event.Code].Minimum
	}

	// build output control info
	outputInfo := ControlInfo{
		Type:    event.Type,
		Code:    event.Code,
		Maximum: 1,
		Minimum: 0,
	}
	if event.Type == evdev.EV_ABS {
		outputInfo.Maximum = 65535
	}

	// construct a new input event with the remapped value
	return &evdev.InputEvent{
		Time:  event.Time,
		Type:  record.Type,
		Code:  record.Code,
		Value: record.Mode(event.Value, inputInfo, outputInfo),
	}, true
}
