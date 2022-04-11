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

Multiple schemas and overlays can be neatly packaged 
into a single file called a `bundle` which specifies type names and their corresponding
schemas. The input will be ingested using the schema that is listed for the
type. The bundle is a JSON file:

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

To compose the `Person/schema` schema:

```
layers compose --repo repo/ http://example.org/Person/schema
```

The `person_sample.json` file contains a sample record. To ingest this:

```
layers ingest json --repo repo/  person_sample.json  --schema http://example.org/Person/schema
```

To ingest the `bundle` using the `Person` schema: 

```
layers ingest json --schema person.schema.json --bundle person-dpv.bundle.json --type Person
```