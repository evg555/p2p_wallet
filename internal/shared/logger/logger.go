package logger

import (
	"fmt"
	"strings"

	"p2p_wallet/internal/shared/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	FormatText = "console"
	FormatJSON = "json"
)

type logger struct {
	log *zap.SugaredLogger
}

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Sync() error
	With(keysAndValues ...any) *logger
}

func New(cfg config.LoggerConfig) (*logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return &logger{}, err
	}

	format, err := parseFormat(cfg.Format)
	if err != nil {
		return &logger{}, err
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.Encoding = format

	base, err := zapCfg.Build()
	if err != nil {
		return &logger{}, fmt.Errorf("build zap logger: %w", err)
	}

	return &logger{log: base.Sugar()}, nil
}

func (l *logger) Debug(msg string, keysAndValues ...any) {
	l.log.Debugw(msg, keysAndValues...)
}

func (l *logger) Info(msg string, keysAndValues ...any) {
	l.log.Infow(msg, keysAndValues...)
}

func (l *logger) Warn(msg string, keysAndValues ...any) {
	l.log.Warnw(msg, keysAndValues...)
}

func (l *logger) Error(msg string, keysAndValues ...any) {
	l.log.Errorw(msg, keysAndValues...)
}

func (l *logger) Sync() error {
	return l.log.Sync()
}

func (l *logger) With(keysAndValues ...any) *logger {
	return &logger{log: l.log.With(keysAndValues...)}
}

func parseLevel(raw string) (zapcore.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "info":
		return zap.InfoLevel, nil
	case "debug":
		return zap.DebugLevel, nil
	case "warn":
		return zap.WarnLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q, allowed: debug/info/warn/error", raw)
	}
}

func parseFormat(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported log format %q, allowed: text/json", raw)
	}
}
