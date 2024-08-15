package config

import (
	"emperror.dev/errors"
	"gopkg.in/yaml.v3"
	"time"
)

type Duration time.Duration

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

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return d.UnmarshalText([]byte(text))
}
