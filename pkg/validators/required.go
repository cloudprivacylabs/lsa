package validators

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// RequiredTerm validates if a required properties exist
var RequiredTerm = ls.NewTerm(ls.LS, "validation/required", false, false, ls.OverrideComposition, struct {
	RequiredValidator
}{
	RequiredValidator{},
})

// RequiredValidator validates if a required value exists
type RequiredValidator struct{}

// Validate checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) Validate(docNode, schemaNode graph.Node) error {
	if docNode == nil {
		return nil
	}
	required := ls.AsPropertyValue(schemaNode.GetProperty(RequiredTerm)).AsStringSlice()
	if len(required) > 0 {
		names := make(map[string]struct{})
		for _, node := range graph.TargetNodes(docNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.HasTerm)) {
			name := ls.AsPropertyValue(node.GetProperty(ls.AttributeNameTerm))
			if name.IsString() {
				names[name.AsString()] = struct{}{}
			}
		}
		for _, str := range required {
			if _, ok := names[str]; !ok {
				return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: %s" + str}
			}
		}
	}
	return nil
}

// CompileTerm compiles the required properties array
func (validator RequiredValidator) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if !value.IsStringSlice() {
		return ls.ErrValidatorCompile{Validator: RequiredTerm, Msg: "Array of required attributes expected"}
	}
	return nil
}
