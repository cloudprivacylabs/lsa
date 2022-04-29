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

## Annotating FHIR messages with Data Privacy Vocabulary

The [Data Privacy Vocabulary](https://w3c.github.io/dpv/dpv) enables
expressing machine-readable metadata about the use and processing of
personal data based on legislative requirements such as the General
Data Protection Regulation. Layered schemas can be used to annotate
FHIR messages with DPV terms.

For this purpose, a `Patient` overlay that adds DPV terms to patient
data fields can be defined as follows:


```
patient-dpv.overlay.json:
{
    "@context": [
        "https://lschema.org/ls.json",
        {
            "dpv":"http://www.w3id.org/dpv#"
        }
    ],
    "@type": "Overlay",
    "@id": "https://hl7.org/fhir/Patient/dpv-overlay",
    "valueType": "https://hl7.org/fhir/Patient",
    "attributeOverlays": [
        {
            "@id":"https://hl7.org/fhir/Patient",
            "@type":  "dpv:DataSubject"
        },
        {
            "@id":"https://hl7.org/fhir/Patient/name/*/family",
            "dpv:hasPersonalDataCategory": [
                {
                    "@id": "dpv:Identifying"
                },
                {
                    "@id": "dpv:Name"
                }
            ]
        },
        {
            "@id":"https://hl7.org/fhir/Patient/name/*/given/*",
            "dpv:hasPersonalDataCategory": [
                {
                    "@id": "dpv:Identifying"
                },
                {
                    "@id": "dpv:Name"
                }
            ]
        }
    ]
}
```

The `@context` defines `dpv:` as an alias for the DPV namespace `http://www.w3id.org/dpv#`

`"@id": "https://hl7.org/fhir/Patient/dpv-overlay"` is the identifier for this overlay.

`"valueType": "https://hl7.org/fhir/Patient"`: This overlay is for a FHIR patient.

`"attributeOverlays"`: This object contains the entries that address
attributes by ID, and annotate them.


The following will add `dpv:DataSubject` to the `Patient` labels:
```

        {
            "@id":"https://hl7.org/fhir/Patient",
            "@type":  "dpv:DataSubject"
        },
```

The following will add `dpv:Identifying` and `dpv:Name` personal data categories to patient family name attribute:

```
       {
            "@id":"https://hl7.org/fhir/Patient/name/*/family",
            "dpv:hasPersonalDataCategory": [
                {
                    "@id": "dpv:Identifying"
                },
                {
                    "@id": "dpv:Name"
                }
            ]
        },
```

The output graph has the following `Patient` node. The node has
`DataSubject` annotation as one of the node labels.  ```

```
    {
      "n": 34,
      "id": "http://example.org/root.entry.1.resource",
      "labels": [
        "http://www.w3id.org/dpv#DataSubject",
        "https://lschema.org/Object",
        "https://lschema.org/DocumentNode",
        "https://hl7.org/fhir/Patient"
      ],
      "properties": {
        "https://lschema.org/Reference/ref": "https://hl7.org/fhir/Patient",
        "https://lschema.org/attributeIndex": "0",
        "https://lschema.org/entitySchema": "https://hl7.org/fhir/Patient",
        "https://lschema.org/schemaNodeId": "https://hl7.org/fhir/Bundle/entry/*/resource/102"
      }
    },
```

To ingest data using the DPV annotations, use:

```
layers ingest json --bundle fhir-dpv.bundle.json --type https://hl7.org/fhir/Patient patient.json --embedSchemaNodes 
```

Since the overlay is defined for `Patient`, any data input that
contains a `Patient` will use the DPV annotations. For example, a FHIR
bundle containing a `Patient`:

```
layers ingest json --bundle fhir-dpv.bundle.json --type https://hl7.org/fhir/Bundle simplebundle.json --embedSchemaNodes 
```



