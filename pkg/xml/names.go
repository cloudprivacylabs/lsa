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

package xml

import (
	"encoding/xml"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// MatchName will test if name matches requiredName.
//
// If required name has both namespace and local name, the name must match exactly.
//
// If required name has only local name, the local name of name must match
func MatchName(name, requiredName xml.Name) bool {
	if len(requiredName.Space) > 0 {
		return name == requiredName
	}
	return name.Local == requiredName.Local
}

// GetXMLName gets the XML name from the node's namespace and
// localname properties
func GetXMLName(node graph.Node) xml.Name {
	return xml.Name{
		Space: ls.AsPropertyValue(node.GetProperty(NamespaceTerm)).AsString(),
		Local: ls.AsPropertyValue(node.GetProperty(ls.AttributeNameTerm)).AsString(),
	}
}
