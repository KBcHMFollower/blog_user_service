package logger

import "context"

const (
	ReqIdKey        = "req-id"
	EmailKey        = "email"
	EventIdKey      = "event-id"
	ReqUserKey      = "req-user"
	ActionUserIdKey = "action-user-id"
	ActionEmailKey  = "action-email"
	ErrKey          = "err"
)

const LoggerCtxKey = KeyType(0)

type KeyType int

type Logger interface {
	DebugContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	Info(msg string, args ...any)
}

func UpdateLoggerCtx(ctx context.Context, key string, v any) context.Context {
	c, ok := ctx.Value(LoggerCtxKey).(map[string]interface{})
	if !ok {
		c = make(map[string]interface{})
	}

	c[key] = v

	return context.WithValue(ctx, LoggerCtxKey, c)
}
