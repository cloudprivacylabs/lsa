package layers

import ()

const LS = "https://layeredschemas.org/"

var AttributeTypes = struct {
	Value       string
	Object      string
	Array       string
	Reference   string
	Composite   string
	Polymorphic string
}{
	Value:       LS + "Value",
	Object:      LS + "Object",
	Array:       LS + "Array",
	Reference:   LS + "Reference",
	Composite:   LS + "Composite",
	Polymorphic: LS + "Polymorphic",
}

var TypeTerms = struct {
	Attributes    string
	AttributeList string
	Reference     string
	ArrayItems    string
	AllOf         string
	OneOf         string
}{
	Attributes:    LS + "Object#attributes",
	AttributeList: LS + "Object#attributeList",
	Reference:     LS + "Reference#reference",
	ArrayItems:    LS + "Array#items",
	AllOf:         LS + "Composite#allOf",
	OneOf:         LS + "Polymorphic#oneOf",
}

type Attribute struct {
	types    []string
	typesMap map[string]struct{}

	Properties map[string]interface{}
}

// Return all recognized attribute types. This is mainly used for
// validation, to ensure there is only one attribute type
func getAttributeTypes(types []string) []string {
	ret := make([]string, 0)
	for _, x := range types {
		if x == AttributeTypes.Value ||
			x == AttributeTypes.Object ||
			x == AttributeTypes.Array ||
			x == AttributeTypes.Reference ||
			x == AttributeTypes.Composite ||
			x == AttributeTypes.Polymorphic {
			ret = append(ret, x)
		}
	}
	return ret
}

func (a *Attribute) GetTypes() []string { return a.types }

func (a *Attribute) AddTypes(t ...string) {
	if a.typesMap == nil {
		a.typesMap = make(map[string]struct{})
	}
	for _, x := range t {
		if _, exists := a.typesMap[x]; !exists {
			a.types = append(a.types, x)
			a.typesMap[x] = struct{}{}
		}
	}
}

func (a *Attribute) RemoveTypes(t ...string) {
	if a.typesMap == nil {
		return
	}
	for _, x := range t {
		delete(a.typesMap, x)
	}
	if len(a.typesMap) != len(a.types) {
		a.types = make([]string, 0, len(a.typesMap))
		for x := range a.typesMap {
			a.types = append(a.types, x)
		}
	}
}

func (a *Attribute) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.typesMap = make(map[string]struct{})
	a.AddTypes(t...)
}

func (a *Attribute) HasType(t string) bool {
	if a.typesMap == nil {
		return false
	}
	_, exists := a.typesMap[t]
	return exists
}

type ValueAttribute struct {
	Attribute
}

type ReferenceAttribute struct {
	Attribute
}

func (r *ReferenceAttribute) SetReference(ref string) {
	r.Properties[TypeTerms.Reference] = ref
}

func (r *ReferenceAttribute) GetReference() string {
	x, ok := r.Properties[TypeTerms.Reference]
	if !ok {
		return ""
	}
	return x.(string)
}

type ArrayAttribute struct {
	Attribute
}

type PolymorphicAttribute struct {
	Attribute
}

type CompositeAttribute struct {
	Attribute
}

type ObjectAttribute struct {
	Attribute
}
