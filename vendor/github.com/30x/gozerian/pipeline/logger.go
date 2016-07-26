package pipeline

import (
	"github.com/Sirupsen/logrus"
	"os"
)

// Logger represents the logging interface
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})

	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
}

var logger Logger

// todo: kinda lame, fix it

func getLogger() Logger {
	if logger == nil {
		getConfig().GetString(ConfigLogLevel)
		level, err := logrus.ParseLevel(getConfig().GetString(ConfigLogLevel))
		if err != nil {
			level = logrus.InfoLevel
		}
		logger = &logrus.Logger{
			Out:       os.Stderr,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     level,
		}
		if err != nil {
			logger.Errorln(err.Error())
		}
	}
	return logger
}
