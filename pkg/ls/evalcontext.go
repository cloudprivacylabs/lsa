package ls

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher"
)

func PropertyValueFromNative(key string, value interface{}) interface{} {
	if s, ok := value.(string); ok {
		return StringPropertyValue(key, s)
	}
	if arr, ok := value.([]string); ok {
		return StringSlicePropertyValue(key, arr)
	}
	return StringPropertyValue(key, fmt.Sprint(value))
}

func NewEvalContext(g *lpg.Graph) *opencypher.EvalContext {
	ctx := opencypher.NewEvalContext(g)
	ctx.PropertyValueFromNativeFilter = PropertyValueFromNative
	return ctx
}
