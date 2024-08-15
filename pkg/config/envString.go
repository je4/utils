package config

import (
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

func (es *EnvString) UnmarshalYAML(value *yaml.Node) error {
	var text string
	value.Decode(&text)
	return es.UnmarshalText([]byte(text))
}
