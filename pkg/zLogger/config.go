package zLogger

import "github.com/je4/trustutil/v2/pkg/config"

type StashConfig struct {
	LogstashHost       string
	LogstashPort       int
	LogstashTraceLevel string
	Namespace          string
	TLS                *config.TLSConfig
	Dataset            string
}

type Config struct {
	Level string
	File  string
	Stash StashConfig
}
