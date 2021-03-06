package validators

import (
	"fmt"
	"regexp"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// PatternTerm validates agains a regex
var PatternTerm = ls.NewTerm(ls.LS+"validation#pattern", false, false, ls.OverrideComposition, struct {
	PatternValidator
}{
	PatternValidator{},
})

// PatternValidator validates a string value agains a regex
type PatternValidator struct{}

// Validate validates the node value if it is non-nil
func (validator PatternValidator) Validate(docNode, schemaNode ls.Node) error {
	if docNode == nil {
		return nil
	}
	value := docNode.GetValue()
	if value == nil {
		return nil
	}
	pattern := schemaNode.GetCompiledDataMap()[PatternTerm].(*regexp.Regexp)
	if pattern.MatchString(fmt.Sprint(value)) {
		return nil
	}
	return ls.ErrValidation{Validator: PatternTerm, Msg: "Value does not match pattern " + pattern.String()}
}

// Compile the pattern
func (validator PatternValidator) CompileTerm(_ string, value *ls.PropertyValue) (interface{}, error) {
	if !value.IsString() {
		return nil, ls.ErrValidatorCompile{Validator: PatternTerm, Msg: "Pattern is not a string value"}
	}
	pattern, err := regexp.Compile(value.AsString())
	if err != nil {
		return nil, ls.ErrValidatorCompile{Validator: PatternTerm, Msg: "Invalid pattern", Err: err}
	}
	return pattern, nil
}
