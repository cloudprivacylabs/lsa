## Lookup Tables

A lookup table provides the options to translate a value to a
canonical value.

![Lookup table graph representation](lookup-tables-graph-representation.png)


### JSON-LD Representation

Inline lookup tables are defined insude a field definition:

```
{ 
  "@id": "field",
  "@type": "Value",
  "lookupTable": {
    "elements": [
        {
            "options": [ "a","b",...],
            "caseSemsitive": false,
            "value": "resultValue",
            "error": "error msg"
        },
        {
            "value": "value for default case",
            "error": "error message for default case
        }
    ]
  }
}
```

The flattened representation looks like this:

```
...
{
  "@id": "fieldId",
  "@type": "https://lschema.org/Value",
  "https://lschema.org/lookupTable": {
    "@id": "_:b1"
  }
},
{
  "@id": "_:b1",
  "https://lschema.org/lookupTable/elements": {
    "@list": [
       {
         "@id": "_:b2"
       }
    ]
  }
},
{
  "@id": "_:b2",
  "https://lschema.org/lookupTable/element/options": [
    "a",
    "b"
  ],
  "https://lschema.org/lookupTable/element/value": "a"
},
```

The schema also accepts lookup table references:

```
{
    "@context": "https://lschema.org/v1/ls.json",
    "@id":"http://id",
    "@type": "Schema",
    "lookupTable": [
      {
         "@id": "http://tbl1",
         "elements": [
           {
              "options": ["a"],
              "value":"a"
           },
           {
              "options": ["b","c"],
              "value":"b"
           }
        ]
      }
    ],
    "layer" :{
       "@type": "Object",
       "attributes": {
         "a1": {
            "@type": "Value",
            "lookupTable": {
              "ref": "http://tbl1"
            }
         }
     }
  } 
}
```

