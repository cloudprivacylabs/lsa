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

// NSMapTerm specifies a namespace map for an overlay. A Namespace
// map includes one or more expressions of the form:
//
//	from -> to
//
// where from and to are attribute id prefixes. All the prefixes of
// attributes that match from are converted to to.
//
// This is necessary when a different variants of a schema is used in
// a complex schema. Each variant gets its own namespace.
var NSMapTerm = StringSliceTerm{
	Term: NewTerm(LS, "nsMap").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register(),
}