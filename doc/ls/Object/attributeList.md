---
title: "Object/attributeList"
type: "term"
see: 
  - label: Object
    url: /Object
  - label: Object/atttibutes
    url: /Object/attributes
---

# Object/attributeList

{{% termheader %}}
Term: https://lschema.org/Object/attributeList
Type: Edge Label 
Use: Schema edges
{{% /termheader %}}

In a schema, `Object/attributeList` edges link an `Object` node to an
ordered list of attribute nodes. The resulting structure means that
the ordered attributes are linked to the `Object` node.

![Object/attributeList model](../attributelist_model.png)

## JSON-LD Schema Representation

The following JSON-LD schema fragment shows an `Object` that has an ordered set of attributes:

```
{
  "@type": "Object",
  "@id": "myObjectId",
  "attributeName": "objectName",
  "attributeList": [
     {
        "@id": "attr1",
        "@type": "Value"
     },
     {
        "@id": "attr2",
        "@type": "Value
     }
  ]
}
```
