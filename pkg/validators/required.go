package validators

import (
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// RequiredTerm validates if a required properties exist.
//
//	{
//	  @id: attrId
//	  validation/required: true
//	}
var RequiredTerm = ls.RegisterBooleanTerm(ls.NewTerm(ls.LS, "validation/required").SetComposition(ls.OverrideComposition).SetTags(ls.ValidationTag, ls.SchemaElementTag).SetMetadata(struct {
	RequiredValidator
}{
	RequiredValidator{},
}))

// RequiredValidator validates if a required value exists
type RequiredValidator struct{}

// ValidateValue checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateValue(value *string, schemaNode *lpg.Node) error {
	if RequiredTerm.PropertyValue(schemaNode) && value == nil {
		return ls.ErrValidation{Validator: RequiredTerm.Name, Msg: "Missing required attribute: " + ls.GetNodeID(schemaNode)}
	}
	return nil
}

// ValidateNode checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) ValidateNode(docNode, schemaNode *lpg.Node) error {
	if RequiredTerm.PropertyValue(schemaNode) {
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
