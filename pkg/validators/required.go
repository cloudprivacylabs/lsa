package validators

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg"
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

// ValidateValue checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateValue(value *string, schemaNode *lpg.Node) error {
	if ls.AsPropertyValue(schemaNode.GetProperty(RequiredTerm)).AsString() == "true" && value == nil {
		return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
	}
	return nil
}

// ValidateNode checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateNode(docNode, schemaNode *lpg.Node) error {
	if ls.AsPropertyValue(schemaNode.GetProperty(RequiredTerm)).AsString() == "true" {
		if docNode == nil {
			return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
		}
		_, ok := ls.GetRawNodeValue(docNode)
		if !ok {
			return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
		}
	}
	return nil
}

// Validate checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) Validate(docNode, schemaNode *lpg.Node) error {
	if docNode == nil {
		return nil
	}
	required := ls.AsPropertyValue(schemaNode.GetProperty(RequiredTerm)).MustStringSlice()
	if len(required) > 0 {
		req := make(map[string]struct{})
		for _, x := range required {
			req[x] = struct{}{}
		}
		for edges := docNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			to := edges.Edge().GetTo()
			if !to.GetLabels().Has(ls.DocumentNodeTerm) {
				continue
			}
			id := ls.AsPropertyValue(to.GetProperty(ls.SchemaNodeIDTerm)).AsString()
			if len(id) > 0 {
				delete(req, id)
			}
		}
		if len(req) > 0 {
			return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: " + fmt.Sprint(req)}
		}
	}
	return nil
}

// CompileTerm compiles the required properties array
func (validator RequiredValidator) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if !value.IsString() && !value.IsStringSlice() {
		return ls.ErrValidatorCompile{Validator: RequiredTerm, Object: target, Msg: "Array of required attributes expected"}
	}
	return nil
}
