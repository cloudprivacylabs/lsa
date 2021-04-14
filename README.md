[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/lsa?status.svg)](https://godoc.org/github.com/cloudprivacylabs/lsa)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/lsa)](https://goreportcard.com/report/github.com/cloudprivacylabs/lsa)

# Layered Schemas

Layered schemas architecture is developed to work with unified data
semantics regardless of the data capture locale, data representation,
format, or terminology. Traditional schemas (such as JSON/XML schemas)
describe the structure of data with limited semantic
information. Layered schemas use open-ended semantic annotations to
describe data. These annotations include:

  * Constraints: Required attributes, length limits,...
  * Format: Phone number with area code, date/time,...
  * Language: English, Spanish,...
  * Privacy classifications: Personally identifiable information, sensitive information,...
  * Retention policies: Attributes must be cleared after a period,...
  * Provenance: Data source, signatures, ...

A schema can be "sliced" into multiple layers with each layer
including only some of the annotations. These layers can be replaced
with other layers representing variations in stuctural constraints,
different languages or notational differences based on locale,
different security constraints based on context, etc.

The main documentation site for layered schemas is:

https://layeredschemas.org

This Go module contains the reference implementation of the layered
schema specification. It contains a layered schema compiler to
slice/compose schemas, import JSON schemas, annotate data from
JSON/CSV sources.

This Go module has the following open-source components:

  * `layers`: This is the layered schema compiler CLI for
    * importing JSON/CSV schemas
    * ingesting JSON/CSV data, and outputting annotated data as JSON-LD
    * Slicing and composing schema layers
    * JSON-LD processing (expand, flatten, frame, etc.)
  * `pkg/ls`: This is the core package containing 
    * models for schema layers and schema
    * JSON-LD processing to slice and compose schema layers
    * document model used for data ingestion
  * `pkg/json`: The JSON adapter containing
    * JSON schema to schema layers conversion
    * JSON data ingestion
  * `pkg/csv`: The CSV adapter containing
    * CSV schema description to schema layers conversion
    * CSV data ingestion. This first convers the CSV document to a 
      flat JSON object, and uses the JSON adapter.

The [examples/](examples/) directory contains several layered schema examples.
