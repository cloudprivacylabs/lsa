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

package gl

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func TestExpr1(t *testing.T) {
	var v interface{}
	err := json.Unmarshal([]byte(`	{
  "@graph": [
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#product",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "5",
      "https://lschema.org/csv/columnName": "Vaccine_Medicinal_Product"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#certificateId",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "14",
      "https://lschema.org/csv/columnName": "A_Unique_Certificate_Identifier"
    },
    {
      "@id": "Date_Of_Birth",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "10/3/1970",
      "https://lschema.org/csv/columnIndex": "2",
      "https://lschema.org/csv/columnName": "Date_Of_Birth",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#dob"
      }
    },
    {
      "@id": "Date_Of_Vaccination",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "8/1/2021",
      "https://lschema.org/csv/columnIndex": "9",
      "https://lschema.org/csv/columnName": "Date_Of_Vaccination",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#vaccinationDate"
      }
    },
    {
      "@id": "Country_Of_Vaccination",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "United States",
      "https://lschema.org/csv/columnIndex": "11",
      "https://lschema.org/csv/columnName": "Country_Of_Vaccination",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#country"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#vaccinationDate",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "9",
      "https://lschema.org/csv/columnName": "Date_Of_Vaccination"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#healthpassTypeObservation",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "15",
      "https://lschema.org/csv/columnName": "VC_Healthpass_Type_Observation"
    },
    {
      "@id": "A_Unique_Certificate_Identifier",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "12120012001",
      "https://lschema.org/csv/columnIndex": "14",
      "https://lschema.org/csv/columnName": "A_Unique_Certificate_Identifier",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#certificateId"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#manufacturer",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "7",
      "https://lschema.org/csv/columnName": "Vaccine_Marketing_Authorization_Holder_Or_Manufacturer"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#country",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "11",
      "https://lschema.org/csv/columnName": "Country_Of_Vaccination"
    },
    {
      "@id": "last_name",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "Butt",
      "https://lschema.org/csv/columnIndex": "1",
      "https://lschema.org/csv/columnName": "last_name",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#lastName"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#firstName",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "0",
      "https://lschema.org/csv/columnName": "first_name"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#lastName",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "1",
      "https://lschema.org/csv/columnName": "last_name"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#dob",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "2",
      "https://lschema.org/csv/columnName": "Date_Of_Birth"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#doseNumber",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "8",
      "https://lschema.org/csv/columnName": "Number_In_A_Series_Of_Vaccinations_of_Doses"
    },
    {
      "@id": "http://example.org/vaccvc/0",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/data/has": [
        {
          "@id": "first_name"
        },
        {
          "@id": "last_name"
        },
        {
          "@id": "Date_Of_Birth"
        },
        {
          "@id": "Disease_or_Agent_Targeted "
        },
        {
          "@id": "Vaccine_or_Prophylaxis"
        },
        {
          "@id": "Vaccine_Medicinal_Product"
        },
        {
          "@id": "Cvx_Code"
        },
        {
          "@id": "Vaccine_Marketing_Authorization_Holder_Or_Manufacturer"
        },
        {
          "@id": "Number_In_A_Series_Of_Vaccinations_of_Doses"
        },
        {
          "@id": "Date_Of_Vaccination"
        },
        {
          "@id": "State_or_Province_Of_Vaccination"
        },
        {
          "@id": "Country_Of_Vaccination"
        },
        {
          "@id": "Certificate_Issuer_(Name)"
        },
        {
          "@id": "Certificate_Issuer_(Public_Key_Did)"
        },
        {
          "@id": "A_Unique_Certificate_Identifier"
        },
        {
          "@id": "VC_Healthpass_Type_Observation"
        }
      ]
    },
    {
      "@id": "Vaccine_Medicinal_Product",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "Pfizer COVID Vaccine 02",
      "https://lschema.org/csv/columnIndex": "5",
      "https://lschema.org/csv/columnName": "Vaccine_Medicinal_Product",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#product"
      }
    },
    {
      "@id": "Certificate_Issuer_(Name)",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "OSCO Drugs",
      "https://lschema.org/csv/columnIndex": "12",
      "https://lschema.org/csv/columnName": "Certificate_Issuer_(Name)",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#certificateIssuer"
      }
    },
    {
      "@id": "Certificate_Issuer_(Public_Key_Did)",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "2021-0515",
      "https://lschema.org/csv/columnIndex": "13",
      "https://lschema.org/csv/columnName": "Certificate_Issuer_(Public_Key_Did)",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#certificateIssuerPubkey"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#certificateIssuer",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "12",
      "https://lschema.org/csv/columnName": "Certificate_Issuer_(Name)"
    },
    {
      "@id": "Disease_or_Agent_Targeted ",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "COVID-19",
      "https://lschema.org/csv/columnIndex": "3",
      "https://lschema.org/csv/columnName": "Disease_or_Agent_Targeted "
    },
    {
      "@id": "Vaccine_or_Prophylaxis",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "COVID Vaccine",
      "https://lschema.org/csv/columnIndex": "4",
      "https://lschema.org/csv/columnName": "Vaccine_or_Prophylaxis",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#vacinationOrProphylaxis"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#vacinationOrProphylaxis",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "4",
      "https://lschema.org/csv/columnName": "Vaccine_or_Prophylaxis"
    },
    {
      "@id": "Cvx_Code",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "2021-415",
      "https://lschema.org/csv/columnIndex": "6",
      "https://lschema.org/csv/columnName": "Cvx_Code",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#cvx"
      }
    },
    {
      "@id": "Vaccine_Marketing_Authorization_Holder_Or_Manufacturer",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "Pfizer",
      "https://lschema.org/csv/columnIndex": "7",
      "https://lschema.org/csv/columnName": "Vaccine_Marketing_Authorization_Holder_Or_Manufacturer",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#manufacturer"
      }
    },
    {
      "@id": "Number_In_A_Series_Of_Vaccinations_of_Doses",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "2",
      "https://lschema.org/csv/columnIndex": "8",
      "https://lschema.org/csv/columnName": "Number_In_A_Series_Of_Vaccinations_of_Doses",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#doseNumber"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#certificateIssuerPubkey",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "13",
      "https://lschema.org/csv/columnName": "Certificate_Issuer_(Public_Key_Did)"
    },
    {
      "@id": "first_name",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "James",
      "https://lschema.org/csv/columnIndex": "0",
      "https://lschema.org/csv/columnName": "first_name",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#firstName"
      }
    },
    {
      "@id": "State_or_Province_Of_Vaccination",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "New Orleans",
      "https://lschema.org/csv/columnIndex": "10",
      "https://lschema.org/csv/columnName": "State_or_Province_Of_Vaccination",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#state"
      }
    },
    {
      "@id": "VC_Healthpass_Type_Observation",
      "@type": "https://lschema.org/DocumentNode",
      "https://lschema.org/attributeValue": "CDC-Permament",
      "https://lschema.org/csv/columnIndex": "15",
      "https://lschema.org/csv/columnName": "VC_Healthpass_Type_Observation",
      "https://lschema.org/data/instanceOf": {
        "@id": "https://pulseconnect.net/vaccination/test/input#healthpassTypeObservation"
      }
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#cvx",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "6",
      "https://lschema.org/csv/columnName": "Cvx_Code"
    },
    {
      "@id": "https://pulseconnect.net/vaccination/test/input#state",
      "@type": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "https://lschema.org/attributeIndex": "10",
      "https://lschema.org/csv/columnName": "State_or_Province_Of_Vaccination"
    }
  ]}`), &v)
	if err != nil {
		panic(err)
	}
	g := graph.NewOCGraph()
	err = ls.UnmarshalJSONLDGraph(v, g, nil)
	if err != nil {
		panic(err)
	}
	scope := NewScope()
	scope.Set("source", g)
	expr, err := Parse("source.first(node->(node.walk('https://lschema.org/data/instanceOf','https://pulseconnect.net/vaccination/test/input#firstName'))).value")
	if err != nil {
		t.Error(err)
		return
	}
	value, err := expr.Evaluate(scope)
	if err != nil {
		t.Error(err)
		return
	}
	nv := value.(StringValue)
	t.Log(nv)
	if nv != "James" {
		t.Errorf("James expected, got %v", nv)
	}
}
