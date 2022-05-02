[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/lsa?status.svg)](https://godoc.org/github.com/cloudprivacylabs/lsa)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/lsa)](https://goreportcard.com/report/github.com/cloudprivacylabs/lsa)
[![Build Status](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml)
# Layered Schemas

Layered Schema Architecture (LSA) enables semantic interoperability
between heterogeneous systems. LSA uses a schema (such as FHIR schemas
for health data) to define data structures, and overlays to add
semantics, contextual metadata, and processing directives. A schema
variant is the composition of the schema base and use-case specific
overlays. Different schema variants can be used to ingest data from or
export data to disparate systems. Each variant encodes source specific
metadata and rules to ingest data into a knowledge graph, or target
specific metadata and rules to translate the knowledge graph into
another format.

A layered schema defines a mapping from structured data (in JSON, XML,
tabular format) to a labeled property graph. For instance, when a JSON
document is ingested using a layered schema, a labeled property graph
is constructed by combining schema information with the input data
elements. The resulting graph is self-contained, and includes both the
input data and the schema. This allows processing data based on its
semantic attributes instead of its attributes names or other
structural properties. For example, it is possible to use layered
schema to ingest health data, mark certain attributes as
"identifiable", and then, remove all attributes that are marked as
such. The algorithm to remove "identifiable" elements can be written
without any knowledge of the input structure. Same algorithm can be
used to de-identify health data as well as marketing data.

The main documentation site for layered schemas is:

https://layeredschemas.org

This Go module contains the reference implementation of the layered
schema specification. It contains a layered schema processor to
slice/compose schemas, import JSON schemas, and ingest data from
JSON/CSV/XML sources.


## Overview

A traditional schema (a JSON schema, an XML schema, etc.) defines one
or more data types. Such schemas are primarily used for documentation,
data validation, and automated code generation. A schema includes
definitions about the shape of the data object such as the list of
fields, nested fields, or constraints on field contents. A layered
schema adds additional interchangeable layers to a traditional schema
so the same schema can be adjusted for different use cases, or can be
processed with different processing directives to handle data
variations. The layered schema architecture provides the framework to
extend traditional schemas using such layers.

### Schemas, overlays, and bundles

A *schema* defines a data structure (attributes). A schema has a
`valueType` which is the the type of data structure defined by the
schema (e.g. `Person`).

An *overlay* defines semantics for attributes, and may override schema
definitions. A `Person` schema can only be composed with `Person`
overlays.

A *schema variant* is a schema that is the result of the composition
of a schema with zero or more overlays. 

A *bundle* defines the schema variants for multiple `valueType`s. That
is, a bundle specifies all the schemas and overlays for each type of
object used in a particular use case. For instance, a bundle may
specify that `Person` means the composition of a `Person` schema and
an overlay containing terms from data privacy vocabulary. Data
ingested using this bundle will have annotations using data privacy
vocabulary terms. 

Below is an example schema for a `Person` data type containing
`firstName` and `lastName`.

```
person.schema.json:
{
  "@context": "https://lschema.org/ls.json",
  "@id": "https://example.org/Person/schema",
  "@type": "Schema",
  "valueType": "https://example.org/Person",
  "layer": {
    "@type": "Object",
    "@id": "https://example.org/Person",
    "attributes": {
      "https://example.org/Person/firstName": {
        "@type": "Value"
      },
      "https://example.org/Person/lastName": {
        "@type": "Value"
     }
   }
 }
}
```

The `Person` object is defined as an `Object` that contains two
`Value` attributes: `firstName` and `lastName`. These are the unique
attribute IDs. An overlay can be defined to add new semantics to these
attributes. The following overlay adds personal data category `Name`
to the `lastName` attribute:

```
person-dpv.overlay.json:
{
  "@context": "https://lschema.org/ls.json",
  "@id": "https://example.org/Person/dpv-ovl",
  "@type": "Overlay",
  "valueType": "https://example.org/Person",
  "layer": {
    "https://example.org/lastName": {
      "@type": "Value",
      "http://www.w3.org/ns/dpv#hasPersonalDataCategory": "http://www.w3.org/dpv#Name"
    }
  }
}
```

A bundle that defines `Person` by composing the person schema and the
dpv overlay is:

```
{
    "typeNames" : {
        "https://example.org/Person": {
            "schema" : "person.schema.json",
            "overlays" : [
                {
                    "schema" : "person-dpv.overlay.json"
                }
            ]
        }
    }
}
```

### Composing Schema Variants

A schema variant is a schema and zero or more overlays that modify the
semantics of the schema for a use-case. A schema variant is composed
by combining attributes of overlays with matching attributes of the
schema. By default, the terms specified for the attributes of the
overlays are combined as a set with the same terms of the schema. That
is, if a schema has

```
{
   "@id": <attributeId>,
   "@type": "Value",
   "someAnnotation": "someValue"
}
```

and the overlay has:

```
{
   "@id": <attributeId>,
   "@type": ["Value","someType"],
   "someAnnotation": "otherValue"
}
```

the schema variant will have:

```
{
   "@id": <attributeId>,
   "@type": ["Value","someType"],
   "someAnnotation": ["someValue", "otherValue"]
}
```

Some of the terms have override semantics, and can be used to modify
schema behavior using overlays. For example:

```
{
   "@id": <attributeId>,
   "@type": "Value",
   "valueType": "string"
}
```

The `valueType` can be overridden by an overlay:

```
{
   "@id": <attributeId>,
   "@type": "Value",
   "valueType": "xsd:date"
}
```

To compose schema variants using the `layers` binary, use:

```
layers compose --bundle person-dpv.bundle.json --type https://example.org/Person
```

This will compose and output the schema variant for `Person` as defined in the `person-dpv.bundle`.


### Compiling Schemas

A schema variant may reference other schema variants. For example, the
`Person` object in the `examples/contact` directory refers to a
`Contact` object to represent contact information. Compiling a schema
variant resolves these references and creates a single self-contained
schema. Such a schema can be used for data ingestion.

To compile the `Person` schema, use:

```
layers compile --bundle person-dpv.bundle.json --type https://example.org/Person
```

This operation will attach the `Contact` schema variant as configured
in the bundle to the `Person` variant, resulting in a schema that has
no external references.

### Ingesting Data

Data ingestion process parses an input object, combines the data
attributes with the matching schema attributes and builds a labeled
property graph (LPG). The LPG contains both the input object and the
relevant schema information, so it can be processed using the
annotations in the graph without any external references.

To ingest data using the `layers` binary, use:

```
layers ingest json person_sample.json --bundle person-dpv.bundle.json --type https://example.org/Person
```

This operation will output a graph containing the nodes for the input
as well as nodes for the schema. All the input nodes will have the
label `https://lschema.org/DocumentNode`, and all schema nodes will
have the label `https://lschema.org/Attribute`. Every input node will
be connected to its corresponding schema node with a
`https://lschema.org/instanceOf` edge. The input nodes will contain
`https://lschema.org/schemaNodeId` annotation giving the ID of the
schema node for the attribute.

The schema nodes and input nodes can be combined using the
`embedSchemaNodes` flag:

```
layers ingest json person_sample.json --bundle person-dpv.bundle.json --type https://example.org/Person --embedSchemaNodes
```

The output graph will have nodes containing only
`https://lschema.org/DocumentNode`s that include all the annotations
given in the corresponding schema node. The original schema node can
still be found using the `https://lschema.org/schemaNodeId`
annotation.


## Building

Once you clone the repository, you can build the schema compiler using
the Go build system.

```
cd layers
go build
```

This will build the `layers` binary in the current directory.

## Examples

The `examples/` directory contains some example data processing
scenarios.

## Commercial Support

Commercial support for this library is available from Cloud Privacy Labs: support@cloudprivacylabs.com


