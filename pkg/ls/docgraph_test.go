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
	"os"
	"testing"
)

func TestDup(t *testing.T) {
	d, err := os.ReadFile("testdata/dup.json")
	if err != nil {
		t.Fatal(err)
		return
	}
	m := JSONMarshaler{}
	g := NewDocumentGraph()
	if err := m.Unmarshal(d, g); err != nil {
		t.Fatal(err)
		return
	}
	ei := GetEntityInfo(g)
	if len(FindDuplicatedEntities(ei)) == 0 {
		t.Errorf("No dups")
	}
	RemoveDuplicateEntities(ei)
	ei = GetEntityInfo(g)
	dups := FindDuplicatedEntities(ei)
	if len(dups) != 0 {
		t.Errorf("There are still duplicates: %v", dups)
	}

}
