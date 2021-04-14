# FHIR

This example illustrates working with health data in FHIR JSON format.
The FHIR 4.0 schema is included in this directory as
well:`fhir.schema.json`

## Importing FHIR schema

The first step is to import the FHIR schema into layered schema
format. This is done using the `fhir.json` import specification. This
specification has three sections:

```
{
  [ entities ],
  [ layers ],
  common
}
```

To import the schema:
```
layers import json fhir.json
```

This will read the `fhir.json` import specification, build a schema
for each entity, and then slice that entity using the `layers`
specification and write the output to a file. With the current
configuration, all schema layers will be written under `repo/`.

### Entities 

`entities` is an array where each entry defines an object:
```
        {
            "ref": "./fhir.schema.json#/definitions/Account",
            "name": "Account",
            "id": "http://hl7.org/fhir/Account"
        },
```

  * `ref`: This is the pointer in the schema showing where the object definition is
  * `name`: This is the object name to use for the resulting schema
  * `id`: This is the id of the resulting schema
  
### Layers

`layers` is an array specifying the layers to be sliced from a schema

```
        {
            "@id": "http://hl7.org/fhir/{{.name}}/base",
            "terms": [
                "http://layeredschemas.org/v1.0/attr/name"
            ],
            "file": "repo/{{.name}}_base.jsonld",
            "includeEmpty": true,
            "type": "Schema"
        },
```

  * `id`: The @id of the layer
  * `terms`: The schema terms to include in this layer. The above
    example will only include attribute names.
  * `file`: The output file to write the layer
  * `includeEmpty`: This is only necessary for schema bases to list
    all available attributes even though they have no annotations
  * `type`: `Schema` or `Overlay`
  
The layer specification elements are Go templates, so they can be
generated from the attributes of the current schema. These variable
are avaiable:

  * `{{.name}}`: The name given in the `entities` entry
  * `{{.ref}}`: The `ref` attribute of the entry

Each layer is generated for each entity in the `entities` array.

### Common

This section includes the attributes common to all layers:

```
    "schemaManifest": "repo/{{.name}}_schema.jsonld",
    "schemaId": "http://hl7.org/fhir/{{.name}}/schema",
    "objectType": "http://hl7.org/fhir/{{.name}}",
```

  * `schemaManifest`: File name of the generated schema manifest
  * `schemaId`: The @id of the generated schema
  * `objectType`: The object type of the generated schema and layers


## Ingesting data

There are two example data files, one containing a `Patient`, and
another containing ` Bundle` record.

To ingest the `Patient` record, use:

```
layers ingest json --repo repo/ Patient.json  --schema http://hl7.org/fhir/Patient
```

This command will read `Patient.json` file using the schema
`http://hl7.org/fhir/Patient`, and output the annotated and expanded
document.

To ingest the `Bundle` record, use:

```
layers ingest json --repo repo/ simplebundle.json  --schema http://hl7.org/fhir/Bundle
```

This will do the same with `simplebundle.json` and
`http://hl7.org/fhir/Bundle` schema.

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
relevant schema, and applies it. For instance, the constraint defined
in the `Patient` schema is:

```
      {
        "@id": "http://hl7.org/fhir/Patient.resourceType",
        "@type":"http://layeredschemas.org/v1.0/Value"
        "http://layeredschemas.org/v1.0/attr/enumeration": [ "Patient" ]
        "http://layeredschemas.org/v1.0/attr/required": true
      }, 
```

Because of this constraint, only objects that contain `"resourceType":
"Patient"` are recognized as a `Patient` object.
