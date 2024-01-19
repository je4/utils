package logger

import (
	logging "github.com/op/go-logging"
	"io"
	"os"
)

func CreateLogger(module string, logfile string, w io.Writer, loglevel string, logformat string) (*logging.Logger, io.Closer) {
	log := logging.MustGetLogger(module)
	var w2 io.Writer
	var closer io.Closer = io.NopCloser(nil)
	if logfile != "" {
		lf, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Cannot open logfile %v: %v", logfile, err)
		}
		closer = lf
		//defer lf.Close()
		if w != nil {
			w2 = io.MultiWriter(w, lf)
		}
	} else {
		if w != nil {
			w2 = w
		} else {
			w2 = os.Stderr
		}
	}
	backend := logging.NewLogBackend(w2, "", 0)
	backendLeveled := logging.AddModuleLevel(backend)
	level := logging.GetLevel(loglevel)
	switch loglevel {
	case "DEBUG":
		level = logging.DEBUG
	case "ERROR":
		level = logging.ERROR
	case "WARNING":
		level = logging.WARNING
	case "INFO":
		level = logging.INFO
	case "CRITICAL":
		level = logging.CRITICAL
	}
	backendLeveled.SetLevel(level, "")

	logging.SetFormatter(logging.MustStringFormatter(logformat))
	logging.SetBackend(backendLeveled)

	return log, closer
}
