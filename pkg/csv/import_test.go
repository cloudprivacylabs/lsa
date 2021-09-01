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

package csv

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/bserdar/digraph"
)

func TestCSVImport(t *testing.T) {
	f, err := os.Open("testdata/GHP_data_capture_global.csv")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Error(err)
		return
	}

	idSpec := AttributeSpec{TermCol: 1}
	colSpecs := []TermSpec{
		{
			Term:    "https://lschema.org/validation#format",
			TermCol: 5,
		},
	}
	layer, err := Import(idSpec, colSpecs, 3, 0, records)
	if err != nil {
		t.Error(err)
		return
	}

	for _, s := range []string{"photo", "givenName", "middleName", "familyName", "birthDate", "linkedVaccineCertificate", "recipient", "disease", "vaccineDescription", "vaccineType", "medicinalProductName", "cvxCode", "marketingAuthorizationHolder", "doseNumber", "dosesPerCycle", "dateOfVaccination", "stateOfVaccination", "countryOfVaccination", "certificateIssuer", "certificateNumber"} {
		nodes := layer.GetAllNodes().Select(digraph.NodesByLabelPredicate(s)).All()
		if len(nodes) != 1 {
			t.Errorf("Cannot find %s", s)
		}
	}

}
