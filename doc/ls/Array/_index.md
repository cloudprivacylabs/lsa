---
title: "Array"
type: "term"
see:
  - label: ingestAs
    url: /ingestAs
  - label: attributeName
    url: /attributeName
  - label: attributeIndex
    url: /attributeIndex
---

# Array

{{% termheader %}}
Term: https://lschema.org/Array
Type: Node label
Use: Schema nodes, ingested data nodes
{{% /termheader %}}

In a schema, an `Array` node groups an ordered sequence of other
attributes of a fixed type. When data elements are ingested, an
`Array` groups a sequence of data nodes. A JSON array, or a sequence
of repeating XML elements can be represented as an `Array` node. The
structure of the elements of the array is specified by an attribute
node reached through an `Array/elements` edge.

## Schema Model

If a schema node is declared with `ls:Array` label, `ls:Attribute` label is
automatically added.

The `Array` node includes the definition for the root node of the
array. The `Array/elements` edge connects the `Array` node to the
definition of its elements.

![Array node model](array_node_model.png)

## JSON-LD Schema Representation

The following JSON-LD schema fragment shows an `Array` that value attributes:

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

## Ingested Data Model

Data ingestion behavior is controlled by the value of `ls:ingestAs`
property specified in the schema node.

### `ingestAs = node` (default)

Given the schema: ![Array as node schema ](array_as_node_schema.png)

If in the schema node, the ingest as property  is specified as `ls:ingestAs=node` then the data ingestion will 
behave as ingesting the `array` as a `node`.

Ingesting an `array` as a `node` outputs a 
graph similar to: ![Array as node](array_as_node.png)
where each parent node contains the edge label "has" to each of its child nodes. 

### `ingestAs = edge`

Given the schema: ![Array as edge schema](array_as_edge_schema.png)

If in the schema node, the ingest as property 
is specified as `ls:ingestAs=edge` then the data ingestion will behave as ingesting the `array` as an `edge`. 

Ingesting an `array` as an `edge` outputs a graph similar to: 
![Array as edge](array_as_edge.png) 
where the edge label connecting the parent to the child node, is value of the `attributeName` 
field in the key-value pair of the child node.
