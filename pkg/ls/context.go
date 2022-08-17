package ls

import (
	"context"
	"log"
)

type Context struct {
	context.Context
	logger     Logger
	interner   Interner
	properties map[interface{}]interface{}
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

func (ctx *Context) AdaptToStandardLog(lg Logger) log.Logger {
	stdlog := log.Logger{}
	stdlog.SetOutput(ctx)
	return stdlog
}

func (ctx *Context) Write(p []byte) (n int, err error) {
	ctx.logger.Info(map[string]interface{}{"": string(p)})
	return len(p), nil
}

func (ctx *Context) GetInterner() Interner {
	return ctx.interner
}

func (ctx *Context) Get(key interface{}) interface{} {
	return ctx.properties[key]
}

func (ctx *Context) Set(key, value interface{}) {
	ctx.properties[key] = value
}

func DefaultContext() *Context {
	return &Context{
		Context:    context.Background(),
		logger:     NewDefaultLogger(),
		interner:   NewInterner(),
		properties: make(map[interface{}]interface{}),
	}
}

func NewContext(ctx context.Context) *Context {
	return &Context{
		Context:    ctx,
		logger:     NewDefaultLogger(),
		interner:   NewInterner(),
		properties: make(map[interface{}]interface{}),
	}
}
