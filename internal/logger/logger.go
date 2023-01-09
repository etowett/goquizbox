package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type loggerOptions struct {
	initialFields map[string]interface{}
}

type LoggerOpt func(opts *loggerOptions)

func InitialFields(m map[string]interface{}) LoggerOpt {
	return func(opts *loggerOptions) {
		opts.initialFields = m
	}
}

type Logger interface {
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Flush() error
}

type Entry struct {
	l      Logger
	fields []Field
}

func NewEntry(l Logger, fields ...Field) *Entry {
	return &Entry{
		l:      l,
		fields: fields,
	}
}

func (e *Entry) With(fields ...Field) *Entry {
	return &Entry{
		l:      e.l,
		fields: append(e.fields, fields...),
	}
}

func (e *Entry) WithContext(ctx context.Context) *Entry {
	return &Entry{
		l:      e.l,
		fields: append(e.fields, unwrapContext(ctx)...),
	}
}

func (e *Entry) Panic(msg string) {
	e.l.Panic(msg, e.fields...)
}

func (e *Entry) Panicf(format string, args ...interface{}) {
	e.Panic(fmt.Sprintf(format, args...))
}

func (e *Entry) Fatal(msg string) {
	e.l.Fatal(msg, e.fields...)
}

func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.Fatal(fmt.Sprintf(format, args...))
}

func (e *Entry) Error(msg string) {
	e.l.Error(msg, e.fields...)
}

func (e *Entry) Errorf(format string, args ...interface{}) {
	e.Error(fmt.Sprintf(format, args...))
}

func (e *Entry) Warn(msg string) {
	e.l.Warn(msg, e.fields...)
}

func (e *Entry) Warnf(format string, args ...interface{}) {
	e.Warn(fmt.Sprintf(format, args...))
}

func (e *Entry) Info(msg string) {
	e.l.Info(msg, e.fields...)
}

func (e *Entry) Infof(format string, args ...interface{}) {
	e.Info(fmt.Sprintf(format, args...))
}

func (e *Entry) Debug(msg string) {
	e.l.Debug(msg, e.fields...)
}

func (e *Entry) Debugf(format string, args ...interface{}) {
	e.Debug(fmt.Sprintf(format, args...))
}

func newLogger(cfg zap.Config, envKey string) (*zap.Logger, error) {
	filePath := os.Getenv(envKey)
	if filePath != "" {
		cfg.OutputPaths = []string{
			filePath,
		}
	}

	return cfg.Build()
}

func GetLogger() Logger {
	return logger
}
