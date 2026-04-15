package config

import (
	"encoding"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalTOML(a any) error {
	if text, ok := a.(string); ok {
		return d.UnmarshalText([]byte(text))
	}
	return errors.Errorf("expected string for duration, got %T", a)
}

func (d Duration) MarshalTOML() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", d.String())), nil
}

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

func (d Duration) MarshalYAML() (any, error) {
	return time.Duration(d).String(), nil
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var text string
	if err := value.Decode(&text); err != nil {
		return err
	}
	return d.UnmarshalText([]byte(text))
}

var _ yaml.Unmarshaler = (*Duration)(nil)
var _ yaml.Marshaler = (*Duration)(nil)
var _ fmt.Stringer = (*Duration)(nil)
var _ encoding.TextMarshaler = (*Duration)(nil)
var _ toml.Marshaler = (*Duration)(nil)
var _ toml.Unmarshaler = (*Duration)(nil)
