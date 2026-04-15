package config

import (
	"encoding"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type Time time.Time

func (t *Time) String() string {
	return time.Time(*t).String()
}

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

func (t Time) MarshalYAML() (any, error) {
	return time.Time(t).Format(time.RFC3339), nil
}

func (t *Time) UnmarshalYAML(value *yaml.Node) error {
	var text string
	if err := value.Decode(&text); err != nil {
		return err
	}
	return t.UnmarshalText([]byte(text))
}

func (t Time) MarshalTOML() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", time.Time(t).Format(time.RFC3339))), nil
}

func (t *Time) UnmarshalTOML(a any) error {
	if text, ok := a.(string); ok {
		return t.UnmarshalText([]byte(text))
	}
	return errors.Errorf("expected string for time, got %T", a)
}

var _ yaml.Unmarshaler = (*Time)(nil)
var _ yaml.Marshaler = (*Time)(nil)
var _ toml.Unmarshaler = (*Time)(nil)
var _ toml.Marshaler = (Time)(time.Time{})
var _ fmt.Stringer = (*Time)(nil)
var _ encoding.TextMarshaler = (*Time)(nil)
var _ encoding.TextUnmarshaler = (*Time)(nil)
