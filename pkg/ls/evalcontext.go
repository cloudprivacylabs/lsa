package ls

import (
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher"
)

func PropertyValueFromNative(key string, value any) any {
	return NewPropertyValue(key, value)
}

func NewEvalContext(g *lpg.Graph) *opencypher.EvalContext {
	ctx := opencypher.NewEvalContext(g)
	ctx.PropertyValueFromNativeFilter = PropertyValueFromNative
	return ctx
}
