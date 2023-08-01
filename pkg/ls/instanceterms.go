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

import ()

// InstanceOfTerm is an edge term that is used to connect values with
// their schema specifications
var InstanceOfTerm = RegisterStringTerm(NewTerm(LS, "instanceOf").SetComposition(ErrorComposition))

// SchemaNodeIDTerm denotes the schema node ID for ingested nodes
var SchemaNodeIDTerm = RegisterStringTerm(NewTerm(LS, "schemaNodeId").SetComposition(ErrorComposition).SetTags(SchemaElementTag))
