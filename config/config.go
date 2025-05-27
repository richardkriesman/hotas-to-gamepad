package config

import (
	"github.com/goccy/go-yaml"
	"github.com/richardkriesman/hotas-to-gamepad/device"
	"github.com/richardkriesman/hotas-to-gamepad/mapping"
	"os"
)

type Config struct {
	Inputs map[device.PersistentID]map[Control]Control `yaml:"inputs"`
}

func Load(filePath string) (*Config, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) ToMappingTable() *mapping.Table {
	table := make(mapping.Table)
	for id, controls := range c.Inputs {
		for inputControl, outputControl := range controls {
			table.Add(
				id,
				inputControl.Type,
				inputControl.Code,
				outputControl.Type,
				outputControl.Code,
				mapping.ModeLinear, // TODO: read type and params from json
			)
		}
	}
	return &table
}
