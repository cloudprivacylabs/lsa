# FHIR

This example illustrates working with health data in FHIR JSON format.
The FHIR 4.0 schema is included in this directory as
well:`fhir.schema.json`

## Importing FHIR schema 

To process data conforming to FHIR schema, we first import the FHIR
schema into layered schema format. Each layered schema describes an
entity, so we have to find all top-level FHIR entities first:

```
cat fhir.schema.json |jq '.definitions | keys[]'|sed 's/\"//g' > entities
```

This will give all the keys. Manually edit the output to remove
values, like string, date, etc. 

The an import specification is created:

```
{
    "entities": [
    {"ref":"fhir.schema.json#/definitions/Account","name":"Account","id":"http://hl7.org/fhir/Account"},
    {"ref":"fhir.schema.json#/definitions/Account_Coverage","name":"Account_Coverage","id":"http://hl7.org/fhir/Account_Coverage"},
    {"ref":"fhir.schema.json#/definitions/Account_Guarantor","name":"Account_Guarantor","id":"http://hl7.org/fhir/Account_Guarantor"},
    ...
    ]
}
```

Each entry in `entities` points to a location in the FHIR schema where the object is defined.

Then specify how the schemas are created:
```
{
    "entities": [ ...],
    "schemaManifest": "repo/{{.name}}_schema.json",
    "schemaId": "http://hl7.org/fhir/{{.name}}/schema",
    "objectType": "http://hl7.org/fhir/{{.name}}",
    "layers": [
        {
            "@id": "http://hl7.org/fhir/{{.name}}/base",
            "file": "repo/{{.name}}_base.json",
            "type": "Schema"
        }
    ]
```

This is a Go template evaluated for each entity defined in the
Entities array. As a result of this, 

  * a SchemaManifest is created in `repo/{{.name}}_schema.json`
  * The schema ID and type are derived from entity names
  * Each layer is specified in the `layers` array as separate
    files. This example only creates a schema base.
    
To import all FHIR entities:
```
layers import jsonschema fhir-import.json
```


## Ingesting data

The file `simplebundle.json` contains a simple FHIR bundle record. To ignest:

```
layers ingest json --repo repo/ simplebundle.json  --schema http://hl7.org/fhir/Bundle
```

This command will read the input file using the schema
`http://hl7.org/fhir/Bundle`, and output the annotated and expanded
document.

The FHIR bundle processing illustrates polymorphic references. A FHIR bundle has the following format:

```
{
  "entry": [
     {
       "resource": {
         "resourceType": "Immunization",
         ...
       },
       "resource": {
         "resourceType": "Patient",
         ...
       },
       ...
     }
  ]
}
```

Each `resource` is a different type of object, and the type is
determined by `resourceType` field. Data ingestion process discovers
the type of the `resource` using schema constraints, loads the
relevant schema, and applies it.

## Converting data using templates

The `vax.tmpl` contains a Go template that operates on the ingested
graph to output parts of a vaccine certificate. The template accesses
the data graph nodes to extract information:

```
{{- $patient := ginstanceOf .g "http://hl7.org/fhir/Patient"  | first }}
```

The above finds a node such that:
```
Data Node  --- instanceOf ---> Schema Node
                               type: http://hl7.org/fhir/Patient
```

This template follows the path `-- attributes --> name`
```
{{- $patientName := gpath $patient "https://lschema.org/data/object#attributes" "https://lschema.org/attributeName=name" | first }}
```

This template extracts the birthdate from a patient node by following `-- attributes --> birthDate`:
```
{{ (gpath $patient  "https://lschema.org/data/object#attributes" "https://lschema.org/attributeName=birthDate"  | first).GetValue }}"
```


To process the template:
```
layers template --template vax.tmpl --graph graphfile
```
where the `graphfile` is the ingested data.
