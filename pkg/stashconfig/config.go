package stashconfig

import (
	"github.com/je4/certloader/v2/pkg/loader"
)

type StashConfig struct {
	LogstashHost       string
	LogstashPort       int
	LogstashTraceLevel string
	Namespace          string
	TLS                *loader.Config
	Dataset            string
}

type Config struct {
	Level string
	File  string
	Stash StashConfig
}
