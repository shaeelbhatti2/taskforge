package logging

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey struct{}

func Setup(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}

func WithCorrelation(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, correlationID)
}

func CorrelationID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

func Info(ctx context.Context, msg string, args ...any) {
	cid := CorrelationID(ctx)
	all := append([]any{"correlation_id", cid}, args...)
	slog.InfoContext(ctx, msg, all...)
}

func Error(ctx context.Context, msg string, args ...any) {
	cid := CorrelationID(ctx)
	all := append([]any{"correlation_id", cid}, args...)
	slog.ErrorContext(ctx, msg, all...)
}
