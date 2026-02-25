package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"p2p_wallet/internal/shared/config"

	"github.com/rs/zerolog"
)

var (
	FormatText = "console"
	FormatJSON = "json"
)

type logger struct {
	log *zerolog.Logger
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
	zerolog.TimeFieldFormat = time.RFC3339Nano

	level, err := parseLevel(cfg.Level)
	if err != nil {
		return &logger{}, err
	}

	zerolog.SetGlobalLevel(level)

	format, err := parseFormat(cfg.Format)
	if err != nil {
		return &logger{}, err
	}

	var output io.Writer = os.Stdout
	if format == FormatText {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339Nano,
		}
	}

	log := zerolog.New(output).With().Timestamp().Logger()

	return &logger{log: &log}, nil
}

func (l *logger) Debug(msg string, keysAndValues ...any) {
	event := l.log.Debug()
	appendFields(event, keysAndValues...)
	event.Msg(msg)
}

func (l *logger) Info(msg string, keysAndValues ...any) {
	event := l.log.Info()
	appendFields(event, keysAndValues...)
	event.Msg(msg)
}

func (l *logger) Warn(msg string, keysAndValues ...any) {
	event := l.log.Warn()
	appendFields(event, keysAndValues...)
	event.Msg(msg)
}

func (l *logger) Error(msg string, keysAndValues ...any) {
	event := l.log.Error()
	appendFields(event, keysAndValues...)
	event.Msg(msg)
}

func (l *logger) Sync() error {
	return nil
}

func (l *logger) With(keysAndValues ...any) *logger {
	ctx := l.log.With()

	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		ctx = ctx.Interface(key, keysAndValues[i+1])
	}

	if len(keysAndValues)%2 != 0 {
		last := keysAndValues[len(keysAndValues)-1]
		ctx = ctx.Interface("extra", last)
	}

	child := ctx.Logger()
	return &logger{log: &child}
}

func appendFields(event *zerolog.Event, keysAndValues ...any) {
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		key := fmt.Sprint(keysAndValues[i])
		value := fmt.Sprint(keysAndValues[i+1])
		event.Str(key, value)
	}

	if len(keysAndValues)%2 != 0 {
		last := keysAndValues[len(keysAndValues)-1]
		event.Str("extra", fmt.Sprint(last))
	}
}

func parseLevel(raw string) (zerolog.Level, error) {

	level, err := zerolog.ParseLevel(raw)
	if err != nil {
		return zerolog.NoLevel, fmt.Errorf("invalid log level %s: %w", raw, err)
	}

	return level, nil
}

func parseFormat(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "text", "console":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unsupported log format %q, allowed: text/json", raw)
	}
}
