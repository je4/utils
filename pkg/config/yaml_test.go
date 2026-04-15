package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type yamlConfigStruct struct {
	Timeout Duration  `yaml:"timeout"`
	Host    EnvString `yaml:"host"`
	Start   Time      `yaml:"start"`
}

func TestDuration_YAML(t *testing.T) {
	os.Setenv("TEST_DURATION_YAML", "2h")
	defer os.Unsetenv("TEST_DURATION_YAML")

	t.Run("Unmarshal simple duration", func(t *testing.T) {
		yamlData := `timeout: "1h30m"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, Duration(1*time.Hour+30*time.Minute), cfg.Timeout)
	})

	t.Run("Unmarshal duration with env variable", func(t *testing.T) {
		yamlData := `timeout: "%%TEST_DURATION_YAML%%"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, Duration(2*time.Hour), cfg.Timeout)
	})

	t.Run("Marshal duration", func(t *testing.T) {
		cfg := yamlConfigStruct{
			Timeout: Duration(5 * time.Minute),
		}
		data, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `timeout: 5m0s`)
	})

	t.Run("Roundtrip duration", func(t *testing.T) {
		orig := yamlConfigStruct{
			Timeout: Duration(45 * time.Minute),
		}
		data, err := yaml.Marshal(orig)
		require.NoError(t, err)

		var res yamlConfigStruct
		err = yaml.Unmarshal(data, &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Timeout, res.Timeout)
	})
}

func TestEnvString_YAML(t *testing.T) {
	os.Setenv("TEST_HOST_YAML", "localhost")
	defer os.Unsetenv("TEST_HOST_YAML")

	t.Run("Unmarshal simple string", func(t *testing.T) {
		yamlData := `host: "google.com"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("google.com"), cfg.Host)
	})

	t.Run("Unmarshal string with env variable", func(t *testing.T) {
		yamlData := `host: "%%TEST_HOST_YAML%%"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("localhost"), cfg.Host)
	})

	t.Run("Unmarshal string with mixed content", func(t *testing.T) {
		yamlData := `host: "http://%%TEST_HOST_YAML%%:8080"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("http://localhost:8080"), cfg.Host)
	})

	t.Run("Marshal EnvString", func(t *testing.T) {
		cfg := yamlConfigStruct{
			Host: EnvString("myhost"),
		}
		data, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `host: myhost`)
	})

	t.Run("Roundtrip EnvString", func(t *testing.T) {
		orig := yamlConfigStruct{
			Host: EnvString("roundtrip.example.com"),
		}
		data, err := yaml.Marshal(orig)
		require.NoError(t, err)

		var res yamlConfigStruct
		err = yaml.Unmarshal(data, &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Host, res.Host)
	})
}

func TestTime_YAML(t *testing.T) {
	os.Setenv("TEST_TIME_YAML", "2024-01-01T12:00:00Z")
	defer os.Unsetenv("TEST_TIME_YAML")

	t.Run("Unmarshal simple time", func(t *testing.T) {
		yamlData := `start: "2023-05-15T10:30:00Z"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, "2023-05-15T10:30:00Z")
		assert.Equal(t, Time(expected), cfg.Start)
	})

	t.Run("Unmarshal time with env variable", func(t *testing.T) {
		yamlData := `start: "%%TEST_TIME_YAML%%"`
		var cfg yamlConfigStruct
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
		assert.Equal(t, Time(expected), cfg.Start)
	})

	t.Run("Marshal Time", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2023-12-24T20:00:00Z")
		cfg := yamlConfigStruct{
			Start: Time(start),
		}
		data, err := yaml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `start: "2023-12-24T20:00:00Z"`)
	})

	t.Run("Roundtrip Time", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2024-04-15T15:00:00Z")
		orig := yamlConfigStruct{
			Start: Time(start),
		}
		data, err := yaml.Marshal(orig)
		require.NoError(t, err)

		var res yamlConfigStruct
		err = yaml.Unmarshal(data, &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Start, res.Start)
	})
}
