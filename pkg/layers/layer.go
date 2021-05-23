package layers

import (
	"github.com/bserdar/digraph"
)

const LS = "https://layeredschemas.org/"

const SchemaTerm = LS + "Schema"
const OverlayTerm = LS + "Overlay"

const TargetType = LS + "targetType"

type Layer struct {
	Graph    *digraph.Graph
	RootNode *digraph.Node
}

func (l *Layer) GetID() string {
	return l.RootNode.Label().(string)
}

func (l *Layer) GetLayerType() string {
	schNode := l.RootNode.Payload.(*SchemaNode)
	if schNode.HasType(SchemaTerm) {
		return SchemaTerm
	}
	if schNode.HasType(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

func (l *Layer) GetTargetTypes() []string {
	schNode := l.RootNode.Payload.(*SchemaNode)
	v := schNode.Properties[TargetType]
	if arr, ok := v.([]interface{}); ok {
		ret := make([]string, len(arr))
		for _, x := range arr {
			ret = append(ret, x.(string))
		}
		return ret
	}
	if str, ok := v.(string); ok {
		return []string{str}
	}
	return nil
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(*digraph.Node) bool) {
	var forEachAttribute func(*digraph.Node, func(*digraph.Node) bool) bool
	forEachAttribute = func(root *digraph.Node, f func(*digraph.Node) bool) bool {
		payload, _ := root.Payload.(*SchemaNode)
		if payload.HasType(AttributeTypes.Attribute) {
			if !f(root) {
				return false
			}
		}
		for outgoing := root.AllOutgoingEdges(); outgoing.HasNext(); {
			edge := outgoing.Next()
			if !IsAttributeTreeEdge(edge) {
				continue
			}
			next := edge.To()
			np, _ := next.Payload.(*SchemaNode)
			if np.HasType(AttributeTypes.Attribute) {
				if !forEachAttribute(next, f) {
					return false
				}
			}
		}
		return true
	}

	forEachAttribute(l.RootNode, f)
}
