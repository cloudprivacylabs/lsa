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
