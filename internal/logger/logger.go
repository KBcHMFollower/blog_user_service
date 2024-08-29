package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	ReqIdKey        = "req-id"
	EmailKey        = "email"
	EventIdKey      = "event-id"
	ReqUserKey      = "req-user"
	ActionUserIdKey = "action-user-id"
	ActionEmailKey  = "action-email"
	ErrKey          = "err"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

const loggerCtxKey = keyType(0)

type keyType int

type HandlerMiddleware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) *HandlerMiddleware {
	return &HandlerMiddleware{next: next}
}

func (h *HandlerMiddleware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

func (h *HandlerMiddleware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(loggerCtxKey).(map[string]interface{}); ok {
		for k, v := range c {
			rec.Add(k, v)
		}
	}
	return h.next.Handle(ctx, rec)
}

func (h *HandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddleware{next: h.next.WithAttrs(attrs)}
}

func (h *HandlerMiddleware) WithGroup(name string) slog.Handler {
	return &HandlerMiddleware{next: h.next.WithGroup(name)}
}

func UpdateLoggerCtx(ctx context.Context, key string, v any) context.Context {
	c, ok := ctx.Value(loggerCtxKey).(map[string]interface{})
	if !ok {
		c = make(map[string]interface{})
	}
	c[key] = v
	return context.WithValue(ctx, loggerCtxKey, c)
}

func SetupLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envLocal, envDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true})
	}

	log := slog.New(NewHandlerMiddleware(handler))

	return log
}
