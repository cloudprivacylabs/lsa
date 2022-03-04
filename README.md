[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/lsa?status.svg)](https://godoc.org/github.com/cloudprivacylabs/lsa)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/lsa)](https://goreportcard.com/report/github.com/cloudprivacylabs/lsa)
[![Build Status](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml/badge.svg?branch=main)](https://github.com/cloudprivacylabs/lsa/actions/workflows/CI.yml)
# Layered Schemas

Layered schemas is a semantic interoperability tool. It uses a schema
to define data elements, and additional layers (overlays) to define
semantic annotations for those data elements. These are open-ended
semantic annotations that can be used to mark data using various
ontologies (for instance, privacy/security attributes, retention
policies), localization information (labels, enumerations in different
languages), constraints (format, patterns, encoding). Different sets
of layers can be used to construct different variations of a schema to
account for internationalization, constraint variations, and different
use-cases.

A layered schema defines a mapping from structured data (in JSON, XML,
tabular format) to linked data. For instance, when a JSON document is
ingested using a layered schema, a labeled property graph is
constructed where every element in the schema is a node in the graph,
and each node is linked to its corresponding schema node describing
its semantics. This allows processing data based on its semantic
attributes instead of its name and/or location in the input. For
example, it is possible to use layered schema to ingest health data,
mark certain attributes as "identifiable", and then, remove all
attributes that are marked as such. The algorithm to remove
"identifiable" elements can be written without any knowledge of the
input structure. Same algorithm can be used to de-identify health data
as well as marketing data.

The main documentation site for layered schemas is:

https://layeredschemas.org

This Go module contains the reference implementation of the layered
schema specification. It contains a layered schema compiler to
slice/compose schemas, import JSON schemas, and annotate data from
JSON/CSV sources.

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


