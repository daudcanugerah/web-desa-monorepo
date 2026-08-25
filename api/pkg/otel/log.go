package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogdedup "github.com/veqryn/slog-dedup"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
)

type Log struct {
	l *slog.Logger
}

func NewSLogProvider(name string, level string) slog.Handler {
	var logLevel slog.Level = slog.LevelInfo

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "info":
		logLevel = slog.LevelInfo
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	defaultLog := slogdedup.NewOverwriteHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}), nil)

	slog.SetLogLoggerLevel(logLevel)
	logProvider := global.GetLoggerProvider()
	_, ok := logProvider.(noop.LoggerProvider)
	if ok {
		return defaultLog
	}

	return slogmulti.Fanout(
		otelslog.NewHandler(name, otelslog.WithLoggerProvider(logProvider)),
		defaultLog,
	)
}

func NewLogger(name string, level string) Logger {
	ll := NewSLogProvider(name, level)
	k := &Log{slog.New(ll)}
	k.WithField("app_name", name)
	return k
}

func (l *Log) Debug(ctx context.Context, msg string, keyvalues ...any) {
	l.l.DebugContext(ctx, msg, keyvalues...)
}

func (l *Log) Info(ctx context.Context, msg string, keyvalues ...any) {
	l.l.InfoContext(ctx, msg, keyvalues...)
}

func (l *Log) Warn(ctx context.Context, msg string, keyvalues ...any) {
	l.l.WarnContext(ctx, msg, keyvalues...)
}

func (l *Log) Error(ctx context.Context, msg string, keyvalues ...any) {
	l.l.ErrorContext(ctx, msg, keyvalues...)
}

func (l *Log) WithFields(m map[string]any) Logger {
	z := make([]interface{}, 0, len(m))
	for v, k := range m {
		z = append(z, v, k)
	}

	return &Log{l.l.With(z...)}
}

func (l *Log) WithField(k string, v any) Logger {
	return &Log{l.l.With(k, v)}
}

func (l *Log) Debugf(ctx context.Context, msg string, values ...any) {
	l.l.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf(msg, values...))
}

func (l *Log) Infof(ctx context.Context, msg string, values ...any) {
	l.l.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf(msg, values...))
}

func (l *Log) Warnf(ctx context.Context, msg string, values ...any) {
	l.l.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf(msg, values...))
}

func (l *Log) Errorf(ctx context.Context, msg string, values ...any) {
	l.l.LogAttrs(ctx, slog.LevelError, fmt.Sprintf(msg, values...))
}
