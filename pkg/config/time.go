package config

import (
	"emperror.dev/errors"
	"gopkg.in/yaml.v3"
	"time"
)

type Time time.Time

func (t *Time) UnmarshalText(text []byte) error {
	var es EnvString
	if err := es.UnmarshalText(text); err != nil {
		return errors.WithStack(err)
	}
	time, err := time.Parse("2006-01-02T15:04:05Z", string(es))
	if err == nil {
		*t = (Time)(time)
	}
	return err
}

func (t *Time) MarshalText() ([]byte, error) {
	return []byte(time.Time(*t).Format(time.RFC3339)), nil
}

func (t *Time) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return t.UnmarshalText([]byte(text))
}

func (t *Time) MarshalYAML() (interface{}, error) {
	return time.Time(*t).Format(time.RFC3339), nil
}
