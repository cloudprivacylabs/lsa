---
title: "Value"
type: "term"
see:
  - label: ingestAs
    url: /ingestAs
  - label: attributeName
    url: /attributeName
  - label: attributeIndex
    url: /attributeIndex
---

# Value

{{% termheader %}}
Term: https://lschema.org/Value
Type: Node label
Use: Schema nodes, ingested data nodes
{{% /termheader %}}

In a schema, a `Value` node defines an attribute that has a value, or
in some cases, a set of values. A `Value` schema node is a terminal
node with no child attributes. When data elements are ingested, a
`Value` node usually holds the ingested value. A JSON key-value pair,
a JSON array element, an XML attribute, or an XML element with only
text children can be represented as a `Value` node.

## Schema Model

If a schema node is declared with `ls:Value` label, `ls:Attribute` label is
automatically added.

![Value node model](value_node_model.png)

## JSON-LD Schema Representation

The following JSON-LD schema fragment shows a `Value`:

```
{
  "@type": "Value",
  "@id": "myValueIdId",
  "attributeName": "valueName"
}
```

## Ingested Data Model

Data ingestion behavior is controlled by the value of `ls:ingestAs`
property specified in the schema node.

### `ingestAs = node` (default)

TODO

### `ingestAs = edge`

TODO
