package ls

import (
	"context"
)

type Context struct {
	context context.Context
	logger  Logger
}

func (ctx *Context) GetLogger() Logger {
	if ctx.logger == nil {
		return &nopLogger{}
	}
	return ctx.logger
}

func (ctx *Context) SetLogger(log Logger) *Context {
	ctx.logger = log
	return ctx
}

func DefaultContext() *Context {
	return &Context{context: context.Background(), logger: DefaultLogger{}}
}

func NewContext(ctx context.Context) *Context {
	return &Context{context: ctx, logger: DefaultLogger{}}
}
