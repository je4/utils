package zLogger

import "github.com/rs/zerolog"

type StringArray []string

func (sa StringArray) MarshalZerologArray(arr *zerolog.Array) {
	for _, s := range sa {
		arr.Str(s)
	}
}

var _ = zerolog.LogArrayMarshaler(StringArray(nil))
