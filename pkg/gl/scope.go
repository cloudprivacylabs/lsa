package gl

import ()

// Scope keeps the symbols visible at any point during the execution
// of a script. A nil scope is a valid scope
type Scope struct {
	symbols map[string]Value
	parent  *Scope
}

// NewScope returns a new empty scope
func NewScope() *Scope {
	return &Scope{
		symbols: make(map[string]Value),
	}
}

// NewScope returns a nested scope
func (scope *Scope) NewScope() *Scope {
	ret := NewScope()
	ret.parent = scope
	return ret
}

// Get looks up the symbol in the enclosing scopes
func (scope *Scope) Get(symbol string) Value {
	if scope == nil {
		return nil
	}
	v, ok := scope.symbols[symbol]
	if ok {
		return v
	}
	if scope.parent != nil {
		return scope.parent.Get(symbol)
	}
	return nil
}

// Set the symbol value in the active scope
func (scope *Scope) Set(key string, value interface{}) *Scope {
	scope.symbols[key] = ValueOf(value)
	return scope
}

func (scope *Scope) SetValue(key string, value interface{}) error {
	if _, ok := scope.symbols[key]; ok {
		scope.symbols[key] = ValueOf(value)
		return nil
	}
	if scope.parent != nil {
		return scope.parent.SetValue(key, value)
	}
	return ErrUnknownIdentifier(key)
}
