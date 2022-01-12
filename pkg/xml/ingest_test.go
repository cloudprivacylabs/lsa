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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func xmlIngestAndCheck(xmlname, graphname string) error {
	f, err := os.Open("testdata/" + xmlname + ".xml")
	if err != nil {
		return fmt.Errorf("%s: %w", xmlname, err)
	}
	defer f.Close()

	ingester := &Ingester{}
	root, err := IngestStream(ingester, "a", f)
	if err != nil {
		return fmt.Errorf("%s: %w", xmlname, err)
	}

	d, err := ioutil.ReadFile("testdata/" + graphname + ".json")
	if err != nil {
		return fmt.Errorf("%s: %w", graphname, err)
	}
	tgt := digraph.New()
	if err := ls.UnmarshalGraphJSON(d, tgt, nil); err != nil {
		return fmt.Errorf("%s: %s", graphname, err)
	}

	src := digraph.New()
	src.AddNode(root)
	if !digraph.CheckIsomorphism(tgt.GetIndex(), src.GetIndex(), func(n1, n2 digraph.Node) bool {
		if n1.(ls.Node).GetValue() != n2.(ls.Node).GetValue() {
			fmt.Printf("Not equal: '%v' '%v'\n", n1.(ls.Node).GetValue(), n2.(ls.Node).GetValue())
		}
		return n1.(ls.Node).GetTypes().IsEqual(*n2.(ls.Node).GetTypes()) &&
			n1.(ls.Node).GetValue() == n2.(ls.Node).GetValue() &&
			ls.IsPropertiesEqual(n1.(ls.Node).GetProperties(), n2.(ls.Node).GetProperties())
	}, func(e1, e2 digraph.Edge) bool {
		return ls.IsPropertiesEqual(e1.(ls.Edge).GetProperties(), e2.(ls.Edge).GetProperties())
	}) {
		return fmt.Errorf("%s: Not isomorphic", xmlname)
	}
	return nil
}

func TestBasicIngest(t *testing.T) {
	if err := xmlIngestAndCheck("basicIngest", "basicIngest"); err != nil {
		t.Error(err)
	}
}

func TestBasicIngestWS(t *testing.T) {
	if err := xmlIngestAndCheck("basicIngest_ws", "basicIngest"); err != nil {
		t.Error(err)
	}
}
