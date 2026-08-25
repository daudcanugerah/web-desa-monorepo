package otel

import (
	"context"
)

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

type Level int

// Logger implementation is responsible for providing structured and levled
// logging functions.
type Logger interface {
	WithFields(m map[string]any) Logger
	WithField(key string, v any) Logger

	Debug(ctx context.Context, msg string, keyvalues ...any)
	Info(ctx context.Context, msg string, keyvalues ...any)
	Warn(ctx context.Context, msg string, keyvalues ...any)
	Error(ctx context.Context, msg string, keyvalues ...any)
	Debugf(ctx context.Context, msg string, values ...any)
	Infof(ctx context.Context, msg string, values ...any)
	Warnf(ctx context.Context, msg string, values ...any)
	Errorf(ctx context.Context, msg string, values ...any)
}
