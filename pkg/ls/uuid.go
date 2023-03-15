// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ls

import (
	"github.com/google/uuid"

	"github.com/cloudprivacylabs/lpg"
)

var UUIDTerm = NewTerm(LS, "uuid").SetComposition(OverrideComposition).SetTags(SchemaElementTag).SetMetadata(uuidSemantics{}).Term

type uuidSemantics struct{}

// ProcessNodePostDocIngest will generate a uuid
func (uuidSemantics) ProcessNodePostDocIngest(schemaRootNode, schemaNode *lpg.Node, term *PropertyValue, docNode *lpg.Node) error {
	value := uuid.New().String()
	SetRawNodeValue(docNode, value)
	SetEntityIDVectorElementFromNode(docNode, value)
	return nil
}
