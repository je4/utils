package config

import (
	"gopkg.in/yaml.v3"
	"time"
)

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err == nil {
		*d = (Duration)(duration)
	}
	return err
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return d.UnmarshalText([]byte(text))
}
