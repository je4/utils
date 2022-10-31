package logger

import (
	"testing"
)

func TestMain(m *testing.M) {
	logger, outfile := CreateLogger("test", "", nil, "DEBUG", `%{time:2006-01-02T15:04:05.000} %{shortpkg}::%{longfunc} [%{shortfile}] > %{level:.5s} - %{message}`)
	defer outfile.Close()
	logger.Errorf("xxx")

}
