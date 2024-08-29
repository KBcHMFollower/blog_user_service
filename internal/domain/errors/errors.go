package ctxerrors

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound            = errors.New("resource not found")
	ErrBadRequest          = errors.New("bad request")
	ErrForbidden           = errors.New("forbidden")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConflict            = errors.New("conflict")
	ErrInternalServerError = errors.New("internal server error")
)

type ContextError struct {
	ctx context.Context
	msg string
	err error
}

func NewErr(s context.Context, e error) error {
	return &ContextError{ctx: s, err: e}
}

func (e *ContextError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.err.Error())
}

func (e *ContextError) Unwrap() error {
	if x, ok := e.err.(interface{ Unwrap() error }); ok != false {
		return x.Unwrap()
	}
	return e.err
}

func WrapCtx(ctx context.Context, err error) error {
	errCtx := getCtxOrErrCtx(err, ctx)

	return &ContextError{ctx: errCtx, err: err}
}

func Wrap(message string, err error) error {
	errCtx := getCtxOrErrCtx(err, context.Background())

	return &ContextError{ctx: errCtx, msg: message, err: err}
}

func getCtxOrErrCtx(err error, ctx context.Context) context.Context {
	errCtx := ctx

	var ctxErr ContextError
	if errors.As(err, &ctxErr) {
		errCtx = ctxErr.ctx
	}

	return errCtx
}
