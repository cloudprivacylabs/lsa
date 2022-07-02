---
title: "attributeIndex"
type: "term"
---

{{< termheader term="https://lschema.org/attributeIndex" key="attributeIndex" >}}

Specifies the sequence of the attribute in the object containing
it. 

{{</termheader>}}

When used in a schema attribute, specifies the order of the attribute
in the object containing it. When data elements are ingested, the
`attributeIndex` is set by the ingestion algorithm to the index of the
ingested attribute.

The `attributeIndex` is optional, it does not have to be unique, or
sequential. It should be primarily used for sorting the elements of an
object.

Schema example:

```
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/Person/schemaBase",
    "@type": "Schema",
    "valueType": "https://example.org/Person",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/Person",
        "attributes": [
            {
                "@id": "http://example.org/Person/firstName",
                "@type": "Value",
                "attributeName":"firstName",
                "attributeIndex": 0
            },
            {
                "@id": "http://example.org/Person/lastName",
                "@type": "Value",
                "attributeName": "lastName",
                "attributeIndex": 1
            }
        ]
    }
}
```



