---
title: Using Spreadsheets for Layered Schemas
---

# Using Spreadsheets for Layered Schemas

If the data model is relatively flat (without nested fields, arrays,
or polymorphism), it may be more convenient to work with spreadsheets
instead of JSON-LD files. 

An LSA spreadsheet specification looks like this:

<table class="table table-sm table-bordered">
  <tr class="table-primary">
    <td>https://lschema.org/valueType</td>
    <td>https://example.org/Person</td>
    <td></td>
    <td></td>
  </tr>
  <tr class="table-primary">
    <td>https://lschema.org/entityIdFields</td>
    <td>https://example.org/Person/id</td>
    <td></td>
    <td></td>
  </tr>
  <tr class="table-primary">
    <td>https://lschema.org/description</td>
    <td>This schema defines a Person object</td>
    <td></td>
    <td></td>
  </tr>
  
  <tr class="table-info">
  <td>@id</td><td>@type</td><td>https://lschema.org/valueType</td><td>http://www.w3.org/ns/dpv#hasPersonalDataCategory</td>
  </tr>
  
  <tr class="table-secondary">
  <td>https://example.org/Person/schema</td><td>Schema</td><td>true</td><td></td>
  </tr>
  <tr class="table-secondary">
  <td>https://example.org/Person/dpvoverlay</td><td>Overlay</td><td></td><td>true</td>
  </tr>
  <tr class="table-warning">
  <td>https://example.org/Person</td><td>Object</td><td></td><td></td>
  </tr>
  <tr  class="table-warning">
  <td>id</td><td>Value</td><td>string</td><td></td>
  </tr>
  <tr  class="table-warning">
  <td>firstName</td><td>Value</td><td>string</td><td>http://www.w3.org/ns/dpv#Identifying</td>
  </tr>
  <tr  class="table-warning">
  <td>lastName</td><td>Value</td><td>string</td><td>http://www.w3.org/ns/dpv#Identifying</td>
  </tr>
  <tr  class="table-warning">
  <td>dob</td><td>Value</td><td>xsd:date</td><td>http://www.w3.org/ns/dpv#Identifying</td>
  </tr>
</table>

The LSA specification contains four distinct sections. The first
section (colored purple) starts with the `valueType` or `https://lschema.org/valueType`
term in the first column. Any rows in the spreadsheet before this is
ignored, so if there are any documantation necessary they can be
included before this term. The second column gives the type of the
object defined by the schema. This section may include the
`entityIdFields` term that lists the identity fields. If there is more
than one, the additional fields must be listed in subsequent
columns. Note the naming of the identity fields: all the schema fields
are by default in the same namespace as the schema root. More on this
will be later. If there are additional schema-level metadata, they can
be included here as well. 

The second section (colored light blue) starts at the header row, which is identified by `@id`
in the first column, and `@type` in the second column. Subsequent
columns list the terms used to define the schema.

The third section (colored gray) lists the schema and the overlays defined in this
spreadsheet. In the previous section, the `@id` column specifies the id of the schema or
overlay, and the `@type` column specifies whether that column defines
a schema or an overlay. The remaining columns determine which terms
are included in that layer. If a cell has `true` value, then the
corresponding term is included in that layer. In the above example,
the schema `https://example.org/Person/schema` will include the
`valueType`, but not `hasPersonalDataCategory` term. The
`https://example.org/Person/dpvoverlay` overlay will include the
`hasPersonalDataCategory` term, but not the `valueType` term.

The fourth section (colored yellow) defines the schema attributes. The first attribute
must be an `Object` that is the root node of the schema/overlays. The
root node also specifies the namespace for the remaining
attributes. In the above example, all the attributes will be in the
`https://example.org/Person` namespace unless otherwise specified. The
remaining attributes are defined in the subsequent rows. Each
attribute can be a `Value` or a `Reference`. If the attribute id is an
absolute URL, then the given attribute id is used verbatim, otherwise
the attribute id is assigned relative to the schema root node. In the
above examples, the attributes are named
`https://example.org/Person/id`,
`https://example.org/Person/firstName`, etc.

If the spreadsheet contains multipe sheets, all the sheets are parsed
and layers are extracted. 

These layers can be referenced in a bundle using their layer ids.


{{< highlight json >}}
{
    "schemaSpreadsheets": [
        {
            "file":"person-schema.xlsx",
            "context": [ "https://lschema.org/ls.json" ]
        }
    ],
    "typeNames": {
        "https://example.org/Person": {
            "layerId": "https://example.org/Person/schema",
            "overlays": [
              {
                "layerId": "https://example.org/Person/dpvoverlay"
              }
            ]
        }
    }
}
{{</highlight>}}

The `schemaSpreadsheets` is an array listing all the spreadsheet
files. These can be `.xlsx` or `.csv` files. To enable JSON-LD name
expansion and context must be specified. This is necessary because
spreadsheet schema specification are first translated into JSON-LD,
and then expanded using the given context.

Once specified like this, the schemas and overlays defined in
spreadsheets can be accessed in the bundle using the `layerId` key.
The following command composes the `Person` schema:

```
layers compose --bundle person.bundle.json --type https://example.org/Person
```

