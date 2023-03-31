package validators

import (
	"regexp"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// PatternTerm validates against a regex
var PatternTerm = ls.NewTerm(ls.LS, "validation/pattern").SetComposition(ls.OverrideComposition).SetTags(ls.ValidationTag, ls.SchemaElementTag).SetMetadata(struct {
	PatternValidator
}{
	PatternValidator{},
}).Register()

// PatternValidator validates a string value against a regex
type PatternValidator struct{}

const compiledPatternTerm = "$compiledPattern"

// ValidateValue validates the  value
func (validator PatternValidator) ValidateValue(value *string, schemaNode *lpg.Node) error {
	if value == nil {
		return nil
	}
	ipattern, _ := schemaNode.GetProperty(compiledPatternTerm)
	pattern := ipattern.(*regexp.Regexp)
	if pattern.MatchString(*value) {
		return nil
	}
	return ls.ErrValidation{Validator: PatternTerm, Msg: "Value does not match pattern " + pattern.String()}
}

// Validate validates the node value if it is non-nil
func (validator PatternValidator) ValidateNode(docNode, schemaNode *lpg.Node) error {
	if docNode == nil {
		return nil
	}
	value, ok := ls.GetRawNodeValue(docNode)
	if !ok {
		return nil
	}
	return validator.ValidateValue(&value, schemaNode)
}

// Compile the pattern
func (validator PatternValidator) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if !value.IsString() {
		return ls.ErrValidatorCompile{Validator: PatternTerm, Msg: "Pattern is not a string value"}
	}
	pattern, err := regexp.Compile(value.AsString())
	if err != nil {
		return ls.ErrValidatorCompile{Validator: PatternTerm, Msg: "Invalid pattern", Err: err}
	}
	target.SetProperty(compiledPatternTerm, pattern)
	return nil
}
