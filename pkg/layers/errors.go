package layers

import (
	"errors"
	"fmt"
)

type ErrInvalidInput struct {
	ID  string
	Msg string
}

func (e ErrInvalidInput) Error() string {
	if len(e.Msg) > 0 {
		return fmt.Sprintf("Invalid input: %s - %s", e.ID, e.Msg)
	}
	return fmt.Sprintf("Invalid input: %s", e.ID)
}

func MakeErrInvalidInput(id ...string) error {
	ret := ErrInvalidInput{}
	if len(id) > 0 {
		ret.ID = id[0]
	}
	if len(id) > 1 {
		ret.Msg = id[1]
	}
	return ret
}

type ErrDuplicateAttributeID string

func (e ErrDuplicateAttributeID) Error() string {
	return fmt.Sprintf("Duplicate attribute id: %s", string(e))
}

type ErrMultipleTypes string

func (e ErrMultipleTypes) Error() string {
	return fmt.Sprintf("Multiple types declared for attribute: %s", string(e))
}

var ErrInvalidJsonLdGraph = errors.New("Invalid JsonLd graph")
var ErrAttributeWithoutID = errors.New("Attribute without id")
var ErrNotALayer = errors.New("Not a layer")
var ErrCompositionSourceNotOverlay = errors.New("Composition source is not an overlay")
var ErrIncompatibleComposition = errors.New("Incompatible composition of layers")
