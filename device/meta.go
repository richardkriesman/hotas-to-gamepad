package device

import (
	"github.com/holoplot/go-evdev"
)

type Meta struct {
	evdev.InputPath
	PersistentID PersistentID
}

func List() ([]Meta, error) {
	var list []Meta
	paths, err := evdev.ListDevicePaths()
	if err != nil {
		return list, err
	}
	for _, path := range paths {
		dev, err := evdev.Open(path.Path)
		if err != nil {
			return list, err
		}
		list = append(list, Meta{
			InputPath:    path,
			PersistentID: createPersistentID(dev),
		})
		err = dev.Close()
		if err != nil {
			return list, err
		}
	}
	return list, nil
}

func (d *Meta) Open() (*Device, error) {
	return Open(d.Path)
}
