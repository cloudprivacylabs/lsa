package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// PropertiesValue is a map[string]*PropertyValue on the stack
type PropertiesValue struct {
	BasicValue
	Properties map[string]*ls.PropertyValue
}

var propertiesSelectors = map[string]func(PropertiesValue) (Value, error){
	"length": func(v PropertiesValue) (Value, error) {
		return ValueOf(len(v.Properties)), nil
	},
}

func (value PropertiesValue) Selector(sel string) (Value, error) {
	selected := propertiesSelectors[sel]
	if selected != nil {
		return selected(value)
	}
	return nil, ErrUnknownSelector{Selector: sel}
}

func (value PropertiesValue) Index(index Value) (Value, error) {
	str, err := index.AsString()
	if err != nil {
		return nil, err
	}
	if value.Properties == nil {
		return ValueOf(nil), nil
	}
	v, ok := value.Properties[str]
	if !ok {
		return ValueOf(nil), nil
	}
	if v.IsString() {
		return ValueOf(v.AsString()), nil
	}
	if v.IsStringSlice() {
		return ValueOf(v.AsStringSlice()), nil
	}
	return ValueOf(nil), nil
}

func (value PropertiesValue) AsBool() (bool, error) { return len(value.Properties) > 0, nil }
func (PropertiesValue) AsInt() (int, error)         { return 0, ErrNotANumber }

func (value PropertiesValue) Eq(v Value) (bool, error) {
	p, ok := v.(PropertiesValue)
	if !ok {
		return false, ErrIncomparable
	}
	for k, v := range p.Properties {
		val, ok := value.Properties[k]
		if !ok {
			return false, nil
		}
		if !v.Equal(val) {
			return false, nil
		}
	}
	return true, nil
}
