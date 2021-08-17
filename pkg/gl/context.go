package gl

import ()

type Context struct {
	symbols map[string]Value
	parent  *Context
}

func NewContext() *Context {
	return &Context{
		symbols: make(map[string]Value),
	}
}

func (context *Context) NewNestedContext() *Context {
	ret := NewContext()
	ret.parent = context
	return ret
}

func (context *Context) Get(symbol string) Value {
	v, ok := context.symbols[symbol]
	if ok {
		return v
	}
	if context.parent != nil {
		return context.parent.Get(symbol)
	}
	return nil
}

func (context *Context) Set(key string, value interface{}) *Context {
	context.symbols[key] = ValueOf(value)
	return context
}
