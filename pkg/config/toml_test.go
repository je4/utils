package config

import (
	"os"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configStruct struct {
	Timeout Duration  `toml:"timeout"`
	Host    EnvString `toml:"host"`
	Start   Time      `toml:"start"`
}

func TestDuration_TOML(t *testing.T) {
	os.Setenv("TEST_DURATION", "2h")
	defer os.Unsetenv("TEST_DURATION")

	t.Run("Unmarshal simple duration", func(t *testing.T) {
		tomlData := `timeout = "1h30m"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		assert.Equal(t, Duration(1*time.Hour+30*time.Minute), cfg.Timeout)
	})

	t.Run("Unmarshal duration with env variable", func(t *testing.T) {
		tomlData := `timeout = "%%TEST_DURATION%%"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		assert.Equal(t, Duration(2*time.Hour), cfg.Timeout)
	})

	t.Run("Marshal duration", func(t *testing.T) {
		cfg := configStruct{
			Timeout: Duration(5 * time.Minute),
		}
		data, err := toml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `timeout = "5m0s"`)
	})

	t.Run("Roundtrip duration", func(t *testing.T) {
		orig := configStruct{
			Timeout: Duration(45 * time.Minute),
		}
		data, err := toml.Marshal(orig)
		require.NoError(t, err)

		var res configStruct
		_, err = toml.Decode(string(data), &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Timeout, res.Timeout)
	})
}

func TestEnvString_TOML(t *testing.T) {
	os.Setenv("TEST_HOST", "localhost")
	defer os.Unsetenv("TEST_HOST")

	t.Run("Unmarshal simple string", func(t *testing.T) {
		tomlData := `host = "google.com"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("google.com"), cfg.Host)
	})

	t.Run("Unmarshal string with env variable", func(t *testing.T) {
		tomlData := `host = "%%TEST_HOST%%"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("localhost"), cfg.Host)
	})

	t.Run("Unmarshal string with mixed content", func(t *testing.T) {
		tomlData := `host = "http://%%TEST_HOST%%:8080"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		assert.Equal(t, EnvString("http://localhost:8080"), cfg.Host)
	})

	t.Run("Marshal EnvString", func(t *testing.T) {
		cfg := configStruct{
			Host: EnvString("myhost"),
		}
		data, err := toml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `host = "myhost"`)
	})

	t.Run("Roundtrip EnvString", func(t *testing.T) {
		orig := configStruct{
			Host: EnvString("roundtrip.example.com"),
		}
		data, err := toml.Marshal(orig)
		require.NoError(t, err)

		var res configStruct
		_, err = toml.Decode(string(data), &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Host, res.Host)
	})
}

func TestTime_TOML(t *testing.T) {
	os.Setenv("TEST_TIME", "2024-01-01T12:00:00Z")
	defer os.Unsetenv("TEST_TIME")

	t.Run("Unmarshal simple time", func(t *testing.T) {
		tomlData := `start = "2023-05-15T10:30:00Z"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, "2023-05-15T10:30:00Z")
		assert.Equal(t, Time(expected), cfg.Start)
	})

	t.Run("Unmarshal time with env variable", func(t *testing.T) {
		tomlData := `start = "%%TEST_TIME%%"`
		var cfg configStruct
		_, err := toml.Decode(tomlData, &cfg)
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
		assert.Equal(t, Time(expected), cfg.Start)
	})

	t.Run("Marshal Time", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2023-12-24T20:00:00Z")
		cfg := configStruct{
			Start: Time(start),
		}
		data, err := toml.Marshal(cfg)
		require.NoError(t, err)
		assert.Contains(t, string(data), `start = "2023-12-24T20:00:00Z"`)
	})

	t.Run("Roundtrip Time", func(t *testing.T) {
		start, _ := time.Parse(time.RFC3339, "2024-04-15T15:00:00Z")
		orig := configStruct{
			Start: Time(start),
		}
		data, err := toml.Marshal(orig)
		require.NoError(t, err)

		var res configStruct
		_, err = toml.Decode(string(data), &res)
		require.NoError(t, err)
		assert.Equal(t, orig.Start, res.Start)
	})
}
