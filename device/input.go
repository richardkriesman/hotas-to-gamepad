package device

import (
	"github.com/holoplot/go-evdev"
)

type InputDevice struct {
	absInfos        map[evdev.EvCode]evdev.AbsInfo
	inDev           *evdev.InputDevice
	shouldTerminate bool
	persistentID    PersistentID
}

func Open(path string) (*InputDevice, error) {
	inDev, err := evdev.Open(path)
	if err != nil {
		return nil, err
	}
	if err := inDev.Grab(); err != nil {
		return nil, err
	}
	absInfos, err := inDev.AbsInfos()
	if err != nil {
		return nil, err
	}

	return &InputDevice{
		absInfos:        absInfos,
		inDev:           inDev,
		shouldTerminate: false,
		persistentID:    createPersistentID(inDev),
	}, nil
}

func (d *InputDevice) Close() error {
	d.shouldTerminate = true
	err := d.inDev.Ungrab()
	if err != nil {
		return err
	}
	return d.inDev.Close()
}

func (d *InputDevice) AbsInfos() map[evdev.EvCode]evdev.AbsInfo {
	return d.absInfos
}

func (d *InputDevice) Name() string {
	name, err := d.inDev.Name()
	if err != nil {
		name = "unknown"
	}
	return name
}

func (d *InputDevice) Listen() *Channels {
	channels := Channels{
		Errors: make(chan error),
		Events: make(chan Event),
	}
	go d.listen(&channels)
	return &channels
}

func (d *InputDevice) PersistentID() PersistentID {
	return d.persistentID
}

func (d *InputDevice) Raw() *evdev.InputDevice {
	return d.inDev
}

func (d *InputDevice) listen(channels *Channels) {
	for !d.shouldTerminate {
		// wait for inDev events and build frame
		frame := Frame{}
		seq := uint(0)
		for {
			ev, err := d.inDev.ReadOne()
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
