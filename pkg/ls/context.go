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
		return &NopLogger{}
	}
	return ctx.logger
}

func (ctx *Context) SetLogger(log Logger) {
	ctx.logger = log
}

func DefaultContext() *Context {
	return &Context{context: context.Background(), logger: DefaultLogger{}}
}

func NewContext(ctx context.Context) *Context {
	return &Context{context: ctx, logger: DefaultLogger{}}
}
