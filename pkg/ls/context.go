package ls

import (
	"context"
)

type Context struct {
	context  context.Context
	logger   Logger
	interner Interner
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

func (ctx *Context) GetInterner() Interner {
	return ctx.interner
}

func DefaultContext() *Context {
	return &Context{
		context:  context.Background(),
		logger:   NewDefaultLogger(),
		interner: NewInterner(),
	}
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		context:  ctx,
		logger:   NewDefaultLogger(),
		interner: NewInterner()}
}
