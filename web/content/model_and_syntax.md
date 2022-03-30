---
title: Model and Syntax
menu: 
  main:
    weight: 5
    parent: docs
---

# Schema Model

A layered schema is a labeled property graph that defines a data
structure. The nodes of the graph represent data elements, and the
edges represent relationships between those data elements. A layered
schema can be represented as a JSON-LD file or as a JSON graph
file. Layered schemas can also be imported from CSV file
specifications, or JSON schemas.

{{<figure src="../schemamodel.png" class="text-center my-3" width="100%;">}} 

The JSON-LD representation for this schema is below:

{{< highlight javascript >}}
{
  "@context": "https://lschema.org/ls.json",
  "@type": "Schema",
  "@id": "https://schema_id",
  "valueType": "Obj",
  "entityIdFields": "https://test.org/Obj/idField",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Obj",
    "valueType": "Obj",
    "attributes": {
      "https://test.org/Obj/idField": {
        "@type": "Value",
        "attributeName": "id"
      },
      "https://test.org/Obj/adr": {
        "@type": "Object",
        "attributeName": "adr",
        "attributeList": [
          {
            "@id": "https://test.org/Obj/adr/street",
            "@type": "Value",
            "attributeName": "street"
          },
          {
            "@id": "https://test.org/Obj/adr/state",
            "@type": "Value",
            "attributeName": "state"
          },
          {
            "@id": "https://test.org/Obj/adr/postalCode",
            "@type": "Value",
            "attributeName": "postalCode"
          }
        ]
      }
    }
  }
}
{{< /highlight >}}


# JSON-LD Syntax
