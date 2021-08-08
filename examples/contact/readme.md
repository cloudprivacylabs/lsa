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

The schema base for `Person` is `repo/person_base.jsonld`. It defines
`http://example.org/Person` entity, which includes the three fields
with `contact` being a reference:

```
"attributes": {
  "http://example.org/Person/firstName": {
      "@type": "Value",
      "attributeName":"firstName"
  },
  "http://example.org/Person/lastName": {
      "@type": "Value",
      "attributeName": "lastName"
  },
  "http://example.org/Person/contact": {
      "@type": "Array",
      "attributeName": "contact",
      "items": {
         "@type": "Reference",
          "reference": "http://example.org/Contact/schemaManifest"
      }
  }
}
```

Each attribute of `Person` object is mapped to a
`http://example.org/Person/<attrName>` term. The `contact` attribute
is an array of references that point to `http://example.org/Contact`
objects.

The schema variants for these objects are as follows:

 * `http://example.org/Person/schema`
: This schema marks the `firstName` and `lastName` with `PII` privacy classifications

 * `http://example.org/Person/schema`
: This schema marks the `firtsName` and `lastName` with `PII` and
  `BIT` privacy classifications.
  
Both variants marks the `contact.value` with `phoneNumber` privacy classification.

To compose the `Person/schema` variant:

```
layers compose --repo repo/ http://example.org/Person/schema
```

To compose the `Person_bit/schema` variant:

```
layers compose --repo repo/ http://example.org/Person_bit/schema
```

The `person_sample.json` file contains a sample record. To ingest this:

```
layers ingest json --repo repo/  person_sample.json  --schema http://example.org/Person/schema
```

To ingest the same file using the `Person_bit` schema:

```
layers ingest json --repo repo/  person_sample.json  --schema http://example.org/Person_bit/schema
```
