package ls

import (
	"context"
)

type Context struct {
	context.Context
	logger     Logger
	interner   Interner
	properties map[any]any
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

func (ctx *Context) Get(key any) any {
	return ctx.properties[key]
}

func (ctx *Context) Set(key, value any) {
	ctx.properties[key] = value
}

func DefaultContext() *Context {
	return &Context{
		Context:    context.Background(),
		logger:     NewDefaultLogger(),
		interner:   NewInterner(),
		properties: make(map[any]any),
	}
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		Context:    ctx,
		logger:     NewDefaultLogger(),
		interner:   NewInterner(),
		properties: make(map[any]any),
	}
}
