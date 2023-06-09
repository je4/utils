package config

import "time"

type Time time.Time

func (t *Time) UnmarshalText(text []byte) error {
	time, err := time.Parse("2006-01-02T15:04:05Z", string(text))
	if err == nil {
		*t = (Time)(time)
	}
	return err
}
