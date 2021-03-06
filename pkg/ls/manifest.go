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

// SchemaManifest contains the minimal information to define a schema variant with an optional bundle
type SchemaManifest struct {
	ID         string
	Type       string
	TargetType string
	Bundle     string
	Schema     string
	Overlays   []string
}

// GetID returns the schema manifest ID
func (m *SchemaManifest) GetID() string { return m.ID }
