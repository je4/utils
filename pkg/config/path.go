package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func fullpath(path string) (string, error) {
	path = filepath.ToSlash(filepath.Clean(path))
	// if empty use current folder
	if path == "" {
		currdir, err := os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "cannot get current directory")
		}
		return filepath.ToSlash(currdir), nil
	}
	// replace starting ~ with user home
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "cannot get home directory")
		}
		path = filepath.ToSlash(filepath.Join(home, path[1:]))
	}
	// if it is an absolute path, all fine
	if filepath.IsAbs(path) || path[0] == '/' {
		return path, nil
	}
	currdir, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "cannot get current directory")
	}
	return filepath.ToSlash(filepath.Join(currdir, path)), nil
}

type Path string

func (es *Path) UnmarshalText(text []byte) error {
	str := string(text)
	matches := envRegexp.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		data := os.Getenv(match[1])
		str = strings.ReplaceAll(str, match[0], data)
	}
	str, err := fullpath(str)
	if err != nil {
		return errors.Wrapf(err, "cannot get full path for %s", text)
	}

	*es = (Path)(str)
	return nil
}

func (es *Path) String() string {
	return string(*es)
}

func (es *Path) MarshalText() ([]byte, error) {
	return []byte(*es), nil
}

func (es Path) MarshalYAML() (any, error) {
	return string(es), nil
}

func (es *Path) UnmarshalYAML(value *yaml.Node) error {
	var text string
	if err := value.Decode(&text); err != nil {
		return err
	}
	return es.UnmarshalText([]byte(text))
}

func (es Path) MarshalTOML() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", string(es))), nil
}

func (es *Path) UnmarshalTOML(a any) error {
	if text, ok := a.(string); ok {
		return es.UnmarshalText([]byte(text))
	}
	return fmt.Errorf("expected string for Path, got %T", a)
}

var _ fmt.Stringer = (*Path)(nil)
var _ yaml.Unmarshaler = (*Path)(nil)
var _ yaml.Marshaler = (*Path)(nil)
var _ toml.Marshaler = (Path)("")
var _ toml.Unmarshaler = (*Path)(nil)
