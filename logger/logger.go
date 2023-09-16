package logger

import (
	"os"

	"github.com/op/go-logging"
)

var logger *logging.Logger

func init() {
	InitLogger(logging.INFO)
}

func InitLogger(level logging.Level) {
	var backend logging.Backend
	var format logging.Formatter

	newLogger := logging.MustGetLogger("raha-xray")
	ppid := os.Getppid()

	backend = logging.NewLogBackend(os.Stderr, "", 0)
	if ppid == 0 {
		format = logging.MustStringFormatter(`%{level} - %{message}`)
	} else {
		format = logging.MustStringFormatter(`%{time:2006/01/02 15:04:05} %{level} - %{message}`)
	}

	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(level, "raha-xray")
	newLogger.SetBackend(backendLeveled)

	logger = newLogger
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warning(args ...interface{}) {
	logger.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}
