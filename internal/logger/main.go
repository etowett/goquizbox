package logger

import (
	"context"
	"os"
	"sync"
)

var (
	logger        Logger
	loggerSync    = &sync.Once{}
	msgLoggerSync = &sync.Once{}
)

func MustInit(opts ...LoggerOpt) {
	logger = SetupDefault(os.Getenv("LOG_FILE"), opts...)
}

func SetupDefault(filePath string, opts ...LoggerOpt) Logger {
	return setupLogger(logger, filePath, loggerSync, opts...)
}

func Default(opts ...LoggerOpt) Logger {
	return setupLogger(logger, "", loggerSync, opts...)
}

func setupLogger(
	theLogger Logger,
	filePath string,
	theLoggerSync *sync.Once,
	opts ...LoggerOpt,
) Logger {
	theLoggerSync.Do(func() {
		if theLogger != nil {
			return
		}

		level := InfoLevel
		if os.Getenv("ENV") == "dev" {
			level = DebugLevel
		}

		theLogger = NewZapLogger(filePath, level, opts...)
	})

	return theLogger
}

func Panic(msg string) {
	Default().Panic(msg)
}

func Panicf(format string, args ...interface{}) {
	NewEntry(Default()).Panicf(format, args...)
}

func Fatal(msg string) {
	Default().Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	NewEntry(Default()).Fatalf(format, args...)
}

func Error(msg string) {
	Default().Error(msg)
}

func Errorf(format string, args ...interface{}) {
	NewEntry(Default()).Errorf(format, args...)
}

func Warn(msg string) {
	Default().Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	NewEntry(Default()).Warnf(format, args...)
}

func Info(msg string) {
	Default().Info(msg)
}

func Infof(format string, args ...interface{}) {
	NewEntry(Default()).Infof(format, args...)
}

func Debug(msg string) {
	Default().Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	NewEntry(Default()).Debugf(format, args...)
}

func With(fields ...Field) *Entry {
	return NewEntry(Default(), fields...)
}

func WithContext(ctx context.Context) *Entry {
	return NewEntry(Default()).WithContext(ctx)
}

func Flush() {
	if logger != nil {
		logger.Flush()
	}
}
