# Ingesting and Processing FHIR Objects

This directory demonstrates how to ingest FHIR messages. The relevant
files are:

  * fhir.schema.json: The FHIR 4.0 JSON schema, dowloaded from
    https://www.hl7.org/fhir/downloads.html
  * fhir.bundle.json: The layered schema bundle defining all entities
    in the FHIR schema
  * simplebundle.json: A sample FHIR bundle resource
  * patient.json: A sample FHIR patient resource
  
## The Layered Schema Bundle

This file defines the layered schemas that will be created based on
the FHIR schema. It contains entries of the form:

```
{
    "typeNames" : {
        "https://hl7.org/fhir/Patient": {
            "jsonSchema": {
                "ref": "fhir.schema.json#/definitions/Patient",
                "layerId": "https://hl7.org/fhir/Patient"
            }
        },
        "https://hl7.org/fhir/Bundle": {
            "jsonSchema": {
                "ref": "fhir.schema.json#/definitions/Bundle",
                "layerId": "https://hl7.org/fhir/Bundle"
            }
        },
       ...
```

This snippet shows two entities named `Patient` and
`Bundle`. The JSON schema definitions for these entities are
given in the `jsonSchema.ref` key. For `Patient`, the schema
definition is in the `fhir.schema.json` file, at location
`/definitions/Patient`:

```
fhir.schema.json:
{
  ...
  "definitions": {
    ...
    "Patient": {
      "description": "Demographics and other administrative information ...",
      "properties": {
        "resourceType": {
          "description": "This is a Patient resource",
          "const": "Patient"
        },
        "id": {
        ...
```

Using this JSON schema, `layers` generates a new schema base. That
schema base is assigned the id given in `"layerId":
"https://hl7.org/fhir/Patient"`.

For example, to ingest a FHIR resource of type `Patient`:

```
layers ingest json --bundle fhir.bundle.json --type https://hl7.org/fhir/Patient patient.json
```

This operation does the following:

  * Reads fhir.bundle.json, and generates layered schemas for each typename
  * Ingests `patient.json` file using the layered schema for `Patient`
  

## Ingesting FHIR messages

To ingest a sample patient, use:

```
layers ingest json --bundle fhir.bundle.json --type https://hl7.org/fhir/Patient patient.json
```

This will ingest the `Patient` record based on the schema with data
nodes connected to schema nodes using `ls:instanceOf` edges.

```
layers ingest json --bundle fhir.bundle.json --type https://hl7.org/fhir/Patient patient.json --embedSchemaNodes
```

This will ingest `Patient` record, merging schema nodes with data nodes.

Similarly, to ingest a FHIR Bundle:

```
layers ingest json --bundle fhir.bundle.json --type https://hl7.org/fhir/Bundle simplebundle.json --embedSchemaNodes
```

The result of this operation can be seen in [bundle.png](bundle.png) or [bundle.svg](bundle.svg).

Note that the FHIR bundle contains a polymorphic array containing
different entities. The layered schemas used to ingest those entities
are define in the `fhir.bundle.json` file.
