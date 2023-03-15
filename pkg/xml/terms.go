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
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

const XML = "https://lschema.org/xml/"

// NamespaceTerm captures the element/attribute namespace in the
// ingested data graph. It also determines the namespace for the
// schema element/attribute.
var NamespaceTerm = ls.NewTerm(XML, "ns").SetComposition(ls.OverrideComposition).SetTags(ls.SchemaElementTag).Term

// ValueAttributeTerm gives the name of the attribute containing the value of the node
var ValueAttributeTerm = ls.NewTerm(XML, "valueAttr").SetComposition(ls.OverrideComposition).SetTags(ls.SchemaElementTag).Term

// AttributeTerm marks the attribute as an XML attribute of an element
var AttributeTerm = ls.NewTerm(XML, "attribute").SetComposition(ls.OverrideComposition).SetTags(ls.SchemaElementTag).Term
