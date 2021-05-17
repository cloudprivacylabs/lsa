package layers

type Attribute interface {
	GetID() string

	GetTypes() []string
	HasType(string) bool
	AddTypes(...string)
	RemoveTypes(...string)

	Clone() Attribute
}

type BaseAttribute struct {
	ID string

	types    []string
	typesMap map[string]struct{}
}

func (a *BaseAttribute) GetTypes() []string {
	return a.types
}

func (a *BaseAttribute) HasType(t string) bool {
	_, ok := a.typesMap[t]
	return ok
}

func (a *BaseAttribute) AddTypes(t ...string) {
	for _, x := range t {
		if _, ok := a.typesMap[x]; !ok {
			a.typesMap[x] = struct{}{}
			a.types = append(a.types, x)
		}
	}
}

func (a *BaseAttribute) RemoveTypes(t ...string) {
	for _, x := range t {
		delete(a.typesMap, x)
	}
	a.types = make([]string, 0, len(a.typesMap))
	for x := range x.typesMap {
		a.types = append(a.types, x)
	}
}

type ValueAttribute struct {
	BaseAttribute
}

func (v *ValueAttribute) Clone() Attribute {
	return &ValueAttribute{BaseAttribute: *v.BaseAttribute.Clone()}
}

type ReferenceAttribute struct {
	BaseAttribute
	Reference string
}

func (v *ReferenceAttribute) Clone() Attribute {
	return &ReferenceAttribute{BaseAttribute: *v.BaseAttribute.Clone(),
		Reference: v.Reference,
	}
}

type ArrayAttribute struct {
	BaseAttribute
}

func (v *ArrayAttribute) Clone() Attribute {
	return &ArrayAttribute{BaseAttribute: *v.BaseAttribute.Clone()}
}

type PolymorphicAttribute struct {
	BaseAttribute
}

func (v *PolymorphicAttribute) Clone() Attribute {
	return &PolymorphicAttribute{BaseAttribute: *v.BaseAttribute.Clone()}
}

type CompositeAttribute struct {
	BaseAttribute
	Options []Attribute
}

type ObjectAttribute struct {
	BaseAttribute

	ordered      bool
	attributes   []Attribute
	attributeMap map[string]Attribute
}
