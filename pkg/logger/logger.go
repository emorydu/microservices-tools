package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type Logger interface {
	getLevel() log.Level
	Debug(args ...any)
	Debugf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Panic(args ...any)
	Panicf(format string, args ...any)
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Trace(args ...any)
	Tracef(format string, args ...any)
}

type Config struct {
	LogLevel string `mapstructure:"level"`
}

// appLogger Application logger.
type appLogger struct {
	level  string
	logger *log.Logger
}

// loggerLevelMap For mapping config logger to email-service logger levels.
var loggerLevelMap = map[string]log.Level{
	"debug": log.DebugLevel,
	"info":  log.InfoLevel,
	"warn":  log.WarnLevel,
	"error": log.ErrorLevel,
	"panic": log.PanicLevel,
	"fatal": log.FatalLevel,
	"trace": log.TraceLevel,
}

func (l *appLogger) getLevel() log.Level {
	level, ok := loggerLevelMap[l.level]
	if !ok {
		return log.DebugLevel
	}

	return level
}

func InitLogger(cfg *Config) Logger {
	l := &appLogger{level: cfg.LogLevel}

	l.logger = log.StandardLogger()

	logLevel := l.getLevel()

	env := os.Getenv("APP_ENV")

	if env == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to this.
		log.SetFormatter(&log.TextFormatter{DisableColors: false,
			ForceColors: true, FullTimestamp: true,
		})
	}

	log.SetLevel(logLevel)

	return l
}

func (l *appLogger) Debug(args ...any) {
	l.logger.Debug(args...)
}

func (l *appLogger) Debugf(format string, args ...any) {
	l.logger.Debugf(format, args...)
}

func (l *appLogger) Info(args ...any) {
	l.logger.Info(args...)
}

func (l *appLogger) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

func (l *appLogger) Trace(args ...any) {
	l.logger.Trace(args...)
}

func (l *appLogger) Tracef(format string, args ...any) {
	l.logger.Tracef(format, args...)
}

func (l *appLogger) Error(args ...any) {
	l.logger.Error(args...)
}

func (l *appLogger) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

func (l *appLogger) Warn(args ...any) {
	l.logger.Warn(args...)
}

func (l *appLogger) Warnf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

func (l *appLogger) Panic(args ...any) {
	l.logger.Panic(args...)
}

func (l *appLogger) Panicf(format string, args ...any) {
	l.logger.Panicf(format, args...)
}

func (l *appLogger) Fatal(args ...any) {
	l.logger.Fatal(args...)
}

func (l *appLogger) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}
