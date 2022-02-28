---
title: "Object/attributes"
type: "term"
see: 
  - label: Object
    url: /Object
  - label: Object/atttibuteList
    url: /Object/attributeList
---

# Object/attributes

{{% termheader %}}
Term: https://lschema.org/Object/attributes
Type: Edge Label 
Use: Schema edges
{{% /termheader %}}

In a schema, `Object/attributes` edges link an `Object` node to a set
of attribute nodes. The resulting structure means that the unordered
attributes are linked to the `Object` node.

![Object/attributes model](../attributes_model.png)

## JSON-LD Schema Representation

The following JSON-LD schema fragment shows an `Object` that has an unordered set of attributes:

```
{
  "@type": "Object",
  "@id": "myObjectId",
  "attributeName": "objectName",
  "attributes": {
     "attr1": {
        "@type": "Value"
    },
    "attr2": {
        "@type": "Value
    }
  }
}
```
