package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	l    *zap.Logger
	opts *loggerOptions
}

func NewZapLogger(outputFilePath string, level Level, options ...LoggerOpt) *ZapLogger {
	zl := &ZapLogger{
		opts: &loggerOptions{},
	}
	for _, opt := range options {
		opt(zl.opts)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(getLogLevel(level))
	cfg.Sampling = nil
	cfg.EncoderConfig = encoderConfig()
	cfg.InitialFields = zl.initialFields()

	if outputFilePath != "" {
		cfg.OutputPaths = []string{
			outputFilePath,
		}
	}

	if l, err := cfg.Build(); err != nil {
		panic(fmt.Errorf("while initializing logger: %v", err))
	} else {
		zl.l = l
	}

	return zl
}

func (zl *ZapLogger) initialFields() map[string]interface{} {
	defaultFields := map[string]interface{}{}
	if zl.opts.initialFields == nil || len(zl.opts.initialFields) == 0 {
		return defaultFields
	}

	for k, v := range zl.opts.initialFields {
		defaultFields[k] = v
	}
	return defaultFields
}

func (zl *ZapLogger) Panic(msg string, fields ...Field) {
	zl.l.Panic(msg, fields...)
}

func (zl *ZapLogger) Fatal(msg string, fields ...Field) {
	zl.l.Fatal(msg, fields...)
}

func (zl *ZapLogger) Error(msg string, fields ...Field) {
	zl.l.Error(msg, fields...)
}

func (zl *ZapLogger) Warn(msg string, fields ...Field) {
	zl.l.Warn(msg, fields...)
}

func (zl *ZapLogger) Info(msg string, fields ...Field) {
	zl.l.Info(msg, fields...)
}

func (zl *ZapLogger) Debug(msg string, fields ...Field) {
	zl.l.Debug(msg, fields...)
}

func (zl *ZapLogger) Flush() error {
	return zl.l.Sync()
}

func getLogLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	}

	if os.Getenv("ENV") == "prod" {
		return zapcore.InfoLevel
	} else {
		return zapcore.DebugLevel
	}
}

func encodeTime(d time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(d.Format(time.RFC3339Nano))
}

func encoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()

	return zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		TimeKey:      "time",
		NameKey:      cfg.NameKey,
		LineEnding:   cfg.LineEnding,
		EncodeLevel:  cfg.EncodeLevel,
		EncodeTime:   encodeTime,
		EncodeCaller: cfg.EncodeCaller,
		EncodeName:   cfg.EncodeName,
	}
}
