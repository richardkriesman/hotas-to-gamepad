package config

import (
	"errors"
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/holoplot/go-evdev"
	"strings"
)

type Control struct {
	Type evdev.EvType
	Code evdev.EvCode
}

func NewControl(s string) (Control, error) {
	normalized := strings.ToUpper(s)
	possibleNames := []string{
		normalized,
		fmt.Sprintf("ABS_%s", normalized),
		fmt.Sprintf("BTN_%s", normalized),
	}
	codeMaps := map[evdev.EvType]map[string]evdev.EvCode{
		evdev.EV_ABS: evdev.ABSFromString,
		evdev.EV_KEY: evdev.KEYFromString,
	}

	// check code maps for each supported type for the full name or short name
	for _, name := range possibleNames {
		for evType, codeMap := range codeMaps {
			if _, ok := codeMap[name]; ok {
				return Control{
					Type: evType,
					Code: codeMap[name],
				}, nil
			}
		}
	}

	return Control{}, errors.New(fmt.Sprintf("invalid control: %s", s))
}

func (c Control) MarshalYAML() ([]byte, error) {
	codeName := evdev.CodeName(c.Type, c.Code)
	if codeName == "UNKNOWN" {
		return nil, errors.New(fmt.Sprintf("invalid control. cannot marshal type %d and code %d", c.Type, c.Code))
	}
	return []byte(codeName), nil
}

func (c *Control) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}
	control, err := NewControl(s)
	if err != nil {
		return err
	}
	*c = control
	return nil
}
