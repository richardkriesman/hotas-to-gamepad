package main

import (
	"github.com/holoplot/go-evdev"
)

type RemapTable map[evdev.EvType]map[evdev.EvCode]RemapTableRecord

type RemapTableRecord struct {
	Type evdev.EvType
	Code evdev.EvCode
	Mode RemapFunction
}

type RemapFunction func(value int32, minimum int32, maximum int32) int32

func (t RemapTable) Remap(event *evdev.InputEvent, absInfos map[evdev.EvCode]evdev.AbsInfo) (*evdev.InputEvent, bool) {
	// get remapping data
	record, ok := t[event.Type][event.Code]
	if !ok {
		return nil, false
	}

	// determine minimum and maximum values
	var minimum int32
	var maximum int32
	absInfo, ok := absInfos[event.Code]
	if ok {
		// analog controls
		minimum = absInfo.Minimum
		maximum = absInfo.Maximum
	} else {
		// button controls
		minimum = 0
		maximum = 1
	}

	// construct a new input event with the remapped value
	return &evdev.InputEvent{
		Time:  event.Time,
		Type:  record.Type,
		Code:  record.Code,
		Value: record.Mode(event.Value, minimum, maximum),
	}, true
}

func RemapExact(value int32, _ int32, _ int32) int32 {
	return value
}

func RemapShiftFromPositiveOnly(value int32, _ int32, maximum int32) int32 {
	return value - (maximum / 2)
}
