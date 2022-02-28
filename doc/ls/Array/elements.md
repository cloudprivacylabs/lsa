---
title: "Array/elements"
type: "term"
---

# Array/elements

{{% termheader %}}
Term: https://lschema.org/Array
Type: Edge label
Use: Schema edges
{{% /termheader %}}

In a scheama `Array/elements` edges link an `Array` node to the
definition of array elements. The resulting structure means that the
`Array` has elements whose strucure are described by the node arrived
by following the `Array/elements` edge.

![Array/elements model](../array_node_model.png)


## JSON-LD Schema Representation


```
{
  "@type": "Array",
  "@id": "myArrayId",
  "attributeName": "arrayName",
  "arrayElements": {
     "@id": "arrElements",
     "@type": "Value"
   }
}

```
