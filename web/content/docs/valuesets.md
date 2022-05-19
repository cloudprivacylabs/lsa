---
title: Value sets
---

# Value Sets and Dictionaries

Value sets and dictionaries are used to assign normalized values to
input data that may come in different variations. As an example,
consider the following data sets:

Data Set A:
```
person_id,gender
1,female
2,male
3,male
4,unknown
```
Data Set B:
```
person_id,gender
5,F
6,M
7,F
```

A value set can be used to map these gender values to a predefined set
of normalized values. The following JSON file defines one valueset
named "gender", that maps "F" and "Female" values to "1", "M" and
"Male" values to "2", and any other value to "0":

{{< highlight json >}}
{
   "valuesets": [
      {
         "id" : "gender",
         "values": [
           {
              "values": ["F", "Female"],
              "result": "1"
          },
          {
             "values": ["M","Male"],
             "result": "2"
          },
          {
            "result": "0"
          }
       ]
     }
   ]
}
{{</highlight>}}

The schema to ingest data declares three data fields: `person_id`,
`gender`, and `normalized_gender`. The `normalized_gender` field will
receive the normalized value for the gender field.

{{< highlight json >}}
{
    "@context": "https://lschema.org/ls.json",
    "@id": "https://example.org/Person/schema",
    "@type": "Schema",
    "valueType": "https://example.org/Person",
    "layer": {
        "@type": "Object",
        "@id": "https://example.org/Person",
        "attributes": {
            "https://example.org/Person/id": {
                "@type": "Value",
                "attributeName": "person_id"
            },
            "https://example.org/Person/gender": {
                "@type": "Value",
                "attributeName": "gender"
            },
            "https://example.org/Person/normalized_gender": {
                "@type": "Value",
                "attributeName": "normalized_gender"
            }
        }
    }
}
{{</highlight>}}
