package config

import (
	"gopkg.in/yaml.v3"
	"time"
)

type Time time.Time

func (t *Time) UnmarshalText(text []byte) error {
	time, err := time.Parse("2006-01-02T15:04:05Z", string(text))
	if err == nil {
		*t = (Time)(time)
	}
	return err
}

func (t *Time) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return t.UnmarshalText([]byte(text))
}
