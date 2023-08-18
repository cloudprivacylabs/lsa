# Spreadsheets as Layered Schemas: OMOP schemas

This example demonstrates the use of spreadsheets as layered
schemas. The OMOP common data model is used in systematic analysis of
disparate observational databases storing electronic health
records. OMOP uses a relational model, and uses spreadsheets to
represent both the data model (schemas) and data. With minimal
editing, the same spreadsheets can be converted into layered schemas
to ingest OMOP data.

The file `omop-schemas.xlsx` contains a spreadsheet defining the
schema for the `Person` object.

An LSA schema specification starts with the `valueType` row. This row
specifies the type of the object specified by the spreadsheet.

```
https://lschema.org/valueType, https://ohdsi.org/omop/Person
```

Then, the attribute specification starts with the row containing `@id`
and `@type` as the first two cells. These two are mandatory, and they
specify the identity and type of the object specified. Additional
columns can be used to add semantic annotations to the schema.

The below specification includes the `valueType` as an additional
property for attributes.

```
@id, @type, https://lschena.org/valueType
```

The following lines include definitions of schemas and overlays. In
this example we have a single schema:

```
https://ohdsi.org/omop/Person/schena,Schema,TRUE
```

The first column becomes the schema id. The second colum speficies
that this is a schema. Third and subsequent columns determine if the
column is included in this layer or not. In this example, the
`valueType` will be included in the schema. 

As another example:

```
@id, @type, valueType, language1, language2
id1, Schema, true
id2, Overlay,false, true, false
id3. Overlay, false,false,true
```

The above configuration specifies three layers, a schema and two
overlays. The schema contains `valueType` for attributes. The first
overlay contains the annotations for `language1`, and the second
overlay contains the annotations for `language2`.

The remaining columns are attribute specifications:

```
https://ohdsi.org/omop/Person,Object
person_id,Value, integer
gender_concept_id,Value,integer
...
```

The first attribute *must* be an object, which determines the object
defined by this schema. It's id should be a URL. The remaining
attributes can be Value, or Reference attributes. Other types are not
allowed in a spereadsheet schema specification. If the id of the
attribute is not a URL, the URL of the root node will be used as a
namespace. In the above example, the attributes have ids
`https://ohdsi.org/omop/Person/person_id` and
`https://ohdsi.org/omop/Person/gender_concept_id`.

If the spreadsheet contains multipe sheets, all the sheets are parsed
and layers are extracted. These layers then can be used in a bundle.

``
{
    "schemaSpreadsheets": [
        {
            "file":"omop-schemas.xlsx",
            "context": [ "https://lschema.org/v1/ls.json" ]
        }
    ],
    "typeNames": {
        "https://ohdsi.org/omop/Person": {
            "layerId": "https://ohdsi.org/omop/Person/schema"
        }
    }
}
```

The `schemaSpreadsheets` is an array listing all the spreadsheet
files. These can be `.xlsx` or `.csv` files. To enable JSON-LD name
expansion, and context must be specified. This is necessary because
spreadsheet schema specification are first translated into JSON-LD,
and then expanded using the given context.

Once specified like this, the schemas and overlays defined in
spreadsheets can be accessed in the bundle using the `layerId` key.
The following command composes the `Person` schema:

```
layers compose --bundle omop.bundle.json --type https://ohdsi.org/omop/Person
```
