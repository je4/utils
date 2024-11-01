package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type EnvString string

var envRegexp = regexp.MustCompile(`%%([^%]+)%%`)

func (es *EnvString) UnmarshalText(text []byte) error {
	var str = string(text)
	matches := envRegexp.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		data := os.Getenv(match[1])
		str = strings.ReplaceAll(str, match[0], data)
	}
	*es = (EnvString)(str)
	return nil
}

func (es *EnvString) String() string {
	return string(*es)
}

func (es *EnvString) MarshalText() ([]byte, error) {
	return []byte(*es), nil
}

func (es *EnvString) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return es.UnmarshalText([]byte(text))
}

func (es *EnvString) MarshalYAML() (interface{}, error) {
	return string(*es), nil
}

var _ fmt.Stringer = (*EnvString)(nil)
var _ yaml.Unmarshaler = (*EnvString)(nil)
var _ yaml.Marshaler = (*EnvString)(nil)

// var _ json.Marshaler = (*EnvString)(nil)
// var _ toml.Marshaler = (*EnvString)(nil)
