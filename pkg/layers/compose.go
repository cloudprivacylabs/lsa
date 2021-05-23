package layers

import (
	"fmt"

	"github.com/bserdar/digraph"
)

type ComposeOptions struct {
	// While composing an object with layer1 and layer2, if layer2 has
	// attributes missing in layer1, then add those attributes to the
	// result. By default, the result will only have attributes of
	// layer1
	Union bool
}

// Compose schema layers. Directly modifies the target. The source
// must be an overlay.
func (layer *Layer) Compose(options ComposeOptions, source *Layer) error {
	if source.GetLayerType() != OverlayTerm {
		return ErrCompositionSourceNotOverlay
	}
	layerTypes := layer.GetTargetTypes()
	sourceTypes := source.GetTargetTypes()
	if len(layerTypes) > 0 && len(sourceTypes) > 0 {
		compatible := false
		for _, t := range sourceTypes {
			found := false
			for _, x := range layerTypes {
				if x == t {
					found = true
					break
				}
			}
			if found {
				compatible = true
				break
			}
		}
		if !compatible {
			return ErrIncompatibleComposition
		}
	}

	merged := make(map[*digraph.Node]struct{})
	source.ForEachAttribute(func(sourceNode *digraph.Node) bool {
		snode, _ := sourceNode.Payload.(*SchemaNode)
		// If node exists in target, merge
		targetNodes := layer.Graph.AllNodesWithLabel(node.Label()).All()
		switch len(targetNodes) {
		case 1:
			// Target node exists. Merge if paths match
			if pathsMatch(targetNodes[0], node, node) {
				if err := mergeNodes(targetNodes[0], node, merged); err != nil {
					return err
				}
			}
		case 0:
			// Target node does not exist. Add if options==Union
			if options.Union {

			}
		default:
			return ErrDuplicateAttributeID(fmt.Sprint(node.Label()))
		}
	})
	return nil
}

// Merge source into target.
func mergeNodes(target, source *digraph.Node, merged map[*digraph.Node]struct{}) error {

}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(target, source, initialSource *digraph.Node) bool {
	if source.Label() != target.Label() {
		return false
	}
	sourceParents := GetParentAttributes(source)
	// If sourceParents reached the top level, stop
	if len(sourceParents) != 1 {
		return true
	}
	payload := sourceParents[0].Payload.(*SchemaNode)
	if payload.HasType(SchemaTerm) || payload.HasType(OverlayTerm) {
		return true
	}
	targetParents := GetParentAttributes(target)
	if len(targetParents) != 1 {
		return false
	}
	payload = targetParents[0].Payload.(*SchemaNode)
	if payload.HasType(SchemaTerm) || payload.HasType(OverlayTerm) {
		return false
	}
	// Loop?
	if sourceParents[0] == initialSource {
		return false
	}
	return pathsMatch(targetParents[0], sourceParents[0], initialSource)
}
