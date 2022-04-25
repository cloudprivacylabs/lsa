# Simple contact list

This example illustrates some layered schema featues using a simple
contact list. Each `Person` object contains the name of the person,
and a list of contacts, with a `type` and `value`. The data model is
as follows:

```
Person:
  firstName string
  lastName  string
  contact   []Contact
  
Contact:
  type string
  value string
```

Multiple schemas and overlays can be neatly packaged into a single
file called a `bundle` which specifies type names for entities and
their corresponding schema variants. A schema variant is a schema plus
zero or more overlays. The input will be ingested using the schema
variant that is listed for the type. If a schema in a bundle
references other type names (by using an attribute of type
`Reference`), the reference is resolved using the schema variants in
the bundle. The bundle is a JSON file:

```
{
  "types": {
    "typeName": {
       "schema": "schemaFile",
       "overlays": [
          "overlayFile","overlayFile"
       ]
    },
    "typeName": {
       "schema": "schemaFile",
       "overlays": [
          "overlayFile","overlayFile"
       ]
    },
    ...
}
```

The schema for `Person` is `person.schema.json`. It defines
`http://example.org/Person` entity, which includes the three fields
with `contact` being a reference:

```
"attributes": [
    {
        "@id": "http://example.org/Person/firstName",
        "@type": "Value",
        "attributeName":"firstName"
    },
    {
        "@id": "http://example.org/Person/lastName",
        "@type": "Value",
        "attributeName": "lastName"
    },
    {
        "@id": "http://example.org/Person/contact",
        "@type": "Array",
        "attributeName": "contact",
        "arrayElements": {
            "@type": "Reference",
            "@id": "http://example.org/Person/contact/items",
            "ref": "Contact"
        }
    }
]
```

The `Person` object contains a list of attributes and each attribute is mapped to a
`http://example.org/Person/<attrName>` term. The attribute that contains 
`http://example.org/Person/contact` is an array of references that point to `http://example.org/Contact`
objects.

The file `person-dpv.bundle.json` contains the schemas an overlays
that use the Privacy Data Vocabulary (https://w3c.github.io/dpv/dpv/)
terms. The file `person-pii.bundle.json` uses a `PII` tag mark fields
as personally identifiable information.

To compose the `Person` schema, use:

```
layers compose --bundle person-dpv.bundle.json --type Person
```

Similarly:

```
layers compose --bundle person-pii.bundle.json --type Person
```

These will print the `Person` schema with different compositions.

To resolve links between the schemas and get a complete schema object, use the `compile` option:

```
layers compile --bundle person-dpv.bundle.json --type Person
```

In a compiled schema, all other schema references are resolved an a
composite schema graph is returned.


The `person_sample.json` file contains a sample record. To ingest this:

```
layers ingest json --bundle person-dpv.bundle.json  --type Person  person_sample.json
```

Similarly:

```
layers ingest json --bundle person-dpv.bundle.json --type Person --embedSchemaNodes person_sample.json
```

This will output the ingested graph, with schema nodes embedded into document nodes.

Note that the `person_sample.json` file contains a field not defined
in the schema. By default, such fields will be ingested without an
associated schema node. To ignore such fields:


```
layers ingest json --bundle person-dpv.bundle.json --type Person --embedSchemaNodes --onlySchemaAttributes person_sample.json
```

## Applying layer onto ingested graph

It is possible to apply a layer onto an already ingested graph. For
example, ingest using dpv bundle:

```
layers ingest json --bundle person-dpv.bundle.json --type Person --embedSchemaNodes person_sample.json >dpv.json
```

The ingested graph contains DPV  annotations.

Then:

```
layer applylayer --bundle person-pii.bundle.json --type Person dpv.json
```

The output graph now contains both PII annotations and DPV
annotations.
