package main

import (
	"github.com/holoplot/go-evdev"
)

type InputDevice struct {
	dev             *evdev.InputDevice
	shouldTerminate bool
	Errors          chan error
	Events          chan InputEvent
}

type InputFrame struct {
	Events []*InputEvent
}

type InputEvent struct {
	evdev.InputEvent
	Frame    *InputFrame
	Sequence uint
}

func OpenInputDevice(path string) (*InputDevice, error) {
	dev, err := evdev.Open(path)
	if err != nil {
		return nil, err
	}
	err = dev.Grab()
	if err != nil {
		return nil, err
	}
	inputDevice := InputDevice{
		dev:    dev,
		Errors: make(chan error),
		Events: make(chan InputEvent),
	}
	go inputDevice.listen()
	return &inputDevice, nil
}

func (d *InputDevice) Close() error {
	d.shouldTerminate = true
	err := d.dev.Ungrab()
	if err != nil {
		return err
	}
	return d.dev.Close()
}

func (d *InputDevice) Name() string {
	name, err := d.dev.Name()
	if err != nil {
		name = "unknown"
	}
	return name
}

func (d *InputDevice) RawDevice() *evdev.InputDevice {
	return d.dev
}

func (d *InputDevice) listen() {
	for !d.shouldTerminate {
		// wait for input events and build frame
		frame := InputFrame{}
		seq := uint(0)
		for {
			ev, err := d.dev.ReadOne()
			if err != nil {
				d.Errors <- err
				continue
			}
			if ev.Type == evdev.EV_SYN {
				break
			}
			frame.Events = append(frame.Events, &InputEvent{
				InputEvent: *ev,
				Frame:      &frame,
				Sequence:   seq,
			})
			seq++
		}

		// unwind the frame, emitting each event individually
		for _, event := range frame.Events {
			d.Events <- *event
		}
	}
}
