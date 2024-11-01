package config

import (
	"emperror.dev/errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"time"
)

type Duration time.Duration

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func (d *Duration) UnmarshalText(text []byte) error {
	var es EnvString
	if err := es.UnmarshalText(text); err != nil {
		return errors.WithStack(err)
	}
	duration, err := time.ParseDuration(string(es))
	if err == nil {
		*d = (Duration)(duration)
	}
	return err
}

func (d *Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(*d).String()), nil
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return d.UnmarshalText([]byte(text))
}

func (d *Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(*d).String(), nil
}

var _ yaml.Unmarshaler = (*Duration)(nil)
var _ yaml.Marshaler = (*Duration)(nil)
var _ fmt.Stringer = (*Duration)(nil)
