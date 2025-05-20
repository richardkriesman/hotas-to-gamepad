package mapping

import (
	"github.com/holoplot/go-evdev"
	"github.com/richardkriesman/hotas-to-gamepad/device"
)

type Table map[device.PersistentID]map[evdev.EvType]map[evdev.EvCode]TableRecord

func (t Table) Remap(event *device.InputEvent, outputConfig device.Config) (*evdev.InputEvent, bool) {
	// get remapping data
	record, ok := t[event.Device.PersistentID()][event.Type][event.Code]
	if !ok {
		return nil, false
	}

	// build input control info
	inputInfo := ControlInfo{
		Type:    event.Type,
		Code:    event.Code,
		Maximum: 1, // default value for buttons
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
		Maximum: 1, // default for buttons
		Minimum: 0,
	}
	if event.Type == evdev.EV_ABS {
		outputInfo.Minimum = outputConfig.Axes[record.Code].Minimum
		outputInfo.Maximum = outputConfig.Axes[record.Code].Maximum
	}

	// construct a new input event with the remapped value
	return &evdev.InputEvent{
		Time:  event.Time,
		Type:  record.Type,
		Code:  record.Code,
		Value: record.Mode(event.Value, inputInfo, outputInfo),
	}, true
}
