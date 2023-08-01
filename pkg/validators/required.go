package validators

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// RequiredTerm validates if a required properties exist.
//
//	{
//	  @id: attrId
//	  validation/required: true
//	}
var RequiredTerm = ls.NewTerm(ls.LS, "validation/required").SetComposition(ls.OverrideComposition).SetTags(ls.ValidationTag, ls.SchemaElementTag).SetMetadata(struct {
	RequiredValidator
}{
	RequiredValidator{},
}).Register()

// RequiredValidator validates if a required value exists
type RequiredValidator struct{}

func requiredTermBool(node *lpg.Node) bool {
	pv, ok := ls.GetPropertyValue(node, RequiredTerm.Name)
	if !ok {
		return false
	}
	bl := ls.BooleanType{}
	s, _ := bl.Coerce(pv.Value())
	v, _ := s.(bool)
	return v
}

// ValidateValue checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateValue(value *string, schemaNode *lpg.Node) error {
	if requiredTermBool(schemaNode) && value == nil {
		return ls.ErrValidation{Validator: RequiredTerm.Name, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
	}
	return nil
}

// ValidateNode checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateNode(docNode, schemaNode *lpg.Node) error {
	if requiredTermBool(schemaNode) {
		if docNode == nil {
			return ls.ErrValidation{Validator: RequiredTerm.Name, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
		}
		_, ok := ls.GetRawNodeValue(docNode)
		if !ok {
			return ls.ErrValidation{Validator: RequiredTerm.Name, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
		}
	}
	return nil
}

// Validate checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) Validate(docNode, schemaNode *lpg.Node) error {
	if docNode == nil {
		return nil
	}
	sl := ls.StringSliceType{}
	pv, ok := ls.GetPropertyValue(schemaNode, RequiredTerm.Name)
	if !ok {
		return nil
	}
	slice, err := sl.Coerce(pv.Value())
	if err != nil {
		return nil
	}
	required := slice.([]string)
	if len(required) > 0 {
		req := make(map[string]struct{})
		for _, x := range required {
			req[x] = struct{}{}
		}
		for edges := docNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			to := edges.Edge().GetTo()
			if !to.GetLabels().Has(ls.DocumentNodeTerm.Name) {
				continue
			}
			id := ls.SchemaNodeIDTerm.PropertyValue(to)
			if len(id) > 0 {
				delete(req, id)
			}
		}
		if len(req) > 0 {
			return ls.ErrValidation{Validator: RequiredTerm.Name, Msg: "Missing required attribute: " + fmt.Sprint(req)}
		}
	}
	return nil
}

// CompileTerm compiles the required properties array
func (validator RequiredValidator) CompileTerm(_ *ls.CompileContext, target ls.CompilablePropertyContainer, term string, value ls.PropertyValue) error {
	if _, ok := value.Value().(string); ok {
		return nil
	}
	if _, ok := value.Value().(bool); ok {
		return nil
	}
	sl := ls.StringSliceType{}
	if _, err := sl.Coerce(value.Value()); err != nil {
		return ls.ErrValidatorCompile{Validator: RequiredTerm.Name, Object: target, Msg: "Array of required attributes expected"}
	}
	return nil
}
