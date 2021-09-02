package gl

// A Closure is a function that captures the current scope, with an argument
type Closure struct {
	basicValue
	Symbol string
	F      Evaluatable
}

// Evaluate the closure with a new scope containing the argument set to arg
func (c Closure) Evaluate(arg Value, scope *Scope) (Value, error) {
	scope = scope.NewScope()
	if len(c.Symbol) > 0 {
		scope.Set(c.Symbol, arg)
	}
	return c.F.Evaluate(scope)
}
