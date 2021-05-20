package layers

import (
	"github.com/bserdar/digraph"
	"github.com/piprate/json-gold/ld"
)

// GetKeyValue returns the value of the key in the node. The node must
// be a map
func GetKeyValue(key string, node interface{}) (interface{}, bool) {
	var m map[string]interface{}
	arr, ok := node.([]interface{})
	if ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	} else {
		m, _ = node.(map[string]interface{})
	}
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	return v, ok
}

// GetStringValue returns a string value from the node with the
// key. The node must be a map
func GetStringValue(key string, node interface{}) string {
	v, _ := GetKeyValue(key, node)
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetNodeID returns the node @id. The argument must be a map
func GetNodeID(node interface{}) string {
	return GetStringValue("@id", node)
}

// GetNodeTypes returns the node @type. The argument must be a map
func GetNodeTypes(node interface{}) []string {
	m, ok := node.(map[string]interface{})
	if !ok {
		return nil
	}
	arr, ok := m["@type"].([]interface{})
	if ok {
		ret := make([]string, 0, len(arr))
		for _, x := range arr {
			s, _ := x.(string)
			if len(s) > 0 {
				ret = append(ret, s)
			}
		}
		return ret
	}
	return nil
}

// GetNodeValue returns the node @value. The argument must be a map
func GetNodeValue(node interface{}) interface{} {
	v, _ := GetKeyValue("@value", node)
	return v
}

// GetListElements returns the elements of a @list node. The input can
// be a [{"@list":elements}] or {@list:elements}. If the input cannot
// be interpreted as a list, returns nil
func GetListElements(node interface{}) []interface{} {
	var m map[string]interface{}
	if arr, ok := node.([]interface{}); ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	}
	if m == nil {
		m, _ = node.(map[string]interface{})
	}
	if len(m) == 0 {
		return []interface{}{}
	}
	lst, ok := m["@list"]
	if !ok {
		return nil
	}
	elements, ok := lst.([]interface{})
	if !ok {
		return nil
	}
	return elements
}

// UnmarshalExpanded unmarshals an attribute from expanded
// JSON-LS. The input is a map[string]interface{}. The attribute may
// or may not have an ID.  It must have a type
func UnmarshalExpandedAttribute(target *digraph.Graph, in interface{}) (*digraph.Node, error) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return nil, MakeErrInvalidInput("Invalid attribute")
	}
	id := GetNodeID(in)
	if target.AllNodesWithLabel(id).HasNext() {
		return nil, ErrDuplicateAttributeID(id)
	}
	attribute := Attribute{Properties: make(map[string]interface{})}
	node := target.NewNode(id, nil)
	types := GetNodeTypes(in)
	attribute.SetTypes(types...)
	if len(getAttributeTypes(types)) > 1 {
		return nil, ErrMultipleTypes(id)
	}
	var payload func() interface{}
	// Process the nested attribute nodes
	for k, val := range m {
		switch k {
		case "@id", "@type":
		case TypeTerms.Attributes, TypeTerms.AttributeList:
			attribute.AddTypes(AttributeTypes.Object)
			// m must be an array of attributes
			attrArray, ok := val.([]interface{})
			if !ok {
				return nil, MakeErrInvalidInput(id, "Array of attributes expected here")
			}
			for _, attr := range attrArray {
				attrNode, err := UnmarshalExpandedAttribute(target, attr)
				if err != nil {
					return nil, err
				}
				target.NewEdge(node, attrNode, k, nil)
			}
			if payload != nil {
				return nil, ErrMultipleTypes(id)
			}
			payload = func() interface{} { return &ObjectAttribute{Attribute: attribute} }

		case TypeTerms.Reference:
			attribute.AddTypes(AttributeTypes.Reference)
			// There can be at most one reference
			oid := GetNodeID(val)
			if len(oid) == 0 {
				return nil, MakeErrInvalidInput(id)
			}
			if payload != nil {
				return nil, ErrMultipleTypes(id)
			}
			payload = func() interface{} {
				ret := &ReferenceAttribute{Attribute: attribute}
				ret.SetReference(oid)
				return ret
			}

		case TypeTerms.ArrayItems:
			attribute.AddTypes(AttributeTypes.Array)
			// m must be an array of 1
			itemsArr, ok := val.([]interface{})
			if !ok {
				return nil, MakeErrInvalidInput(id, "Invalid array items")
			}
			if len(itemsArr) > 1 {
				return nil, MakeErrInvalidInput(id, "Multiple array items")
			}
			if len(itemsArr) == 1 {
				itemsNode, err := UnmarshalExpandedAttribute(target, itemsArr[0])
				if err != nil {
					return nil, err
				}
				target.NewEdge(node, itemsNode, k, nil)
			}
			if payload != nil {
				return nil, ErrMultipleTypes(id)
			}
			payload = func() interface{} { return &ArrayAttribute{Attribute: attribute} }

		case TypeTerms.AllOf, TypeTerms.OneOf:
			attribute.AddTypes(AttributeTypes.Array)
			// m must be a list
			elements := GetListElements(val)
			if elements == nil {
				return nil, MakeErrInvalidInput(id)
			}
			for _, element := range elements {
				elemNode, err := UnmarshalExpandedAttribute(target, element)
				if err != nil {
					return nil, err
				}
				target.NewEdge(node, elemNode, k, nil)
			}
			if payload != nil {
				return nil, ErrMultipleTypes(id)
			}
			if k == TypeTerms.AllOf {
				payload = func() interface{} { return &CompositeAttribute{Attribute: attribute} }
			} else {
				payload = func() interface{} { return &PolymorphicAttribute{Attribute: attribute} }
			}

		default:
			// Use RDF
			n, err := JsonLdToGraph(target, val)
			if err != nil {
				return nil, err
			}
			if n != nil {
				target.NewEdge(node, n, k, nil)
			}
		}
	}
	if payload == nil {
		t := getAttributeTypes(types)
		if len(t) == 0 {
			payload = func() interface{} { return &ValueAttribute{Attribute: attribute} }
		} else {
			switch t[0] {
			case AttributeTypes.Value:
				payload = func() interface{} { return &ValueAttribute{Attribute: attribute} }
			case AttributeTypes.Object:
				payload = func() interface{} { return &ObjectAttribute{Attribute: attribute} }
			case AttributeTypes.Array:
				payload = func() interface{} { return &ArrayAttribute{Attribute: attribute} }
			case AttributeTypes.Reference:
				payload = func() interface{} { return &ReferenceAttribute{Attribute: attribute} }
			case AttributeTypes.Composite:
				payload = func() interface{} { return &CompositeAttribute{Attribute: attribute} }
			case AttributeTypes.Polymorphic:
				payload = func() interface{} { return &PolymorphicAttribute{Attribute: attribute} }
			}
		}
	}
	node.Payload = payload()
	return node, nil
}

func JsonLdToGraph(target *digraph.Graph, input interface{}) (*digraph.Node, error) {
	if input == nil {
		return nil, nil
	}
	arr, ok := input.([]interface{})
	if !ok {
		return jsonLdToGraph(target, input)
	}
	switch len(arr) {
	case 0:
		return nil, nil
	case 1:
		return jsonLdToGraph(target, arr[0])
	default:
		newNode := target.NewNode(nil, nil)
		for _, x := range arr {
			n, err := jsonLdToGraph(target, x)
			if err != nil {
				return nil, err
			}
			target.NewEdge(newNode, n, nil, nil)
		}
		return newNode, nil
	}
	return nil, nil
}

func jsonLdToGraph(target *digraph.Graph, input interface{}) (*digraph.Node, error) {
	// The input must be a map or a nonempty array
	// A value or ID produces a single node
	if ld.IsValue(input) {
		return target.NewNode(GetNodeValue(input), nil), nil
	}

	if m, ok := input.(map[string]interface{}); ok {
		if len(m) == 1 {
			if id, ok := m["@id"]; ok {
				return target.NewNode(id, nil), nil
			}
		}
	}

	// A graph containing multiple nodes, use RDF
	rdf, err := ld.NewJsonLdProcessor().ToRDF(input, nil)
	if err != nil {
		return nil, err
	}
	quads := rdf.(*ld.RDFDataset).GetQuads("@default")
	nodes := make(map[string]*digraph.Node)
	for _, quad := range quads {
		// There must be a node for the subject
		value := quad.Subject.GetValue()
		subjectNode, exists := nodes[value]
		if !exists {
			subjectNode = target.NewNode(value, nil)
			nodes[value] = subjectNode
		}

		// If the object is a Literal, create a new node for it
		var objectNode *digraph.Node
		lit, ok := quad.Object.(*ld.Literal)
		if ok {
			objectNode = target.NewNode(lit.GetValue(), nil)
		} else {
			value := quad.Object.GetValue()
			objectNode, exists = nodes[value]
			if !exists {
				objectNode = target.NewNode(value, nil)
				nodes[value] = objectNode
			}
		}

		target.NewEdge(subjectNode, objectNode, quad.Predicate.GetValue(), nil)
	}
	// The graph must have a root, that is, a node with no incoming edges
	var root *digraph.Node
	for _, node := range nodes {
		if !node.AllIncomingEdges().HasNext() {
			if root != nil {
				return nil, ErrInvalidJsonLdGraph
			}
			root = node
		}
	}
	if root == nil {
		return nil, ErrInvalidJsonLdGraph
	}
	return root, nil
}
