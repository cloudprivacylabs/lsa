---
title: Value sets
---

# Value Sets and Dictionaries

Value sets and dictionaries are used to assign normalized values to
input data that may come in different variations. Let's start with an
example. Consider the following data sets:

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

When ingested, the first row becomes:

{{<figure src="ingested.png" class="text-center my-3" >}} 

We now add the valueset annotations using an overlay:

{{< highlight json >}}
{
    "@context": "https://lschema.org/ls.json",
    "@id": "https://example.org/Person/vs_overlay",
    "@type": "Overlay",
    "valueType": "https://example.org/Person",
    "attributeOverlays": [
       {
          "@id": "https://example.org/Person/gender",
          "vsValuesets": "gender",
          "vsContext": "https://example.org/Person",
          "vsResultValues": "https://example.org/Person/normalized_gender"
       }
    ]
}
{{</highlight>}}


`"@id": "https://example.org/Person/gender"`: This is the attribute
where the valueset information is overlayed. The value of `gender`
field will be used to lookup in the valueset.

`"vsValuesets": "gender"`: This specifies the name of the valueset to
use. In this case, the "gender" valueset will be used to lookup the
value.

`"vsContext": "https://example.org/Person"`: This annotation gives the
closes ancestor node of the `gender` node that is an instance of the
`https://example.org/Person` schema node. In this case, it is the root
node corresponding to the current row in the input file. The context
node is the common parent node that contains all the values that will
be looked up, and the root node for all the found values. In this
case, the result of the valueset lookup will be inserted under this
`vsContext` node.

`"vsResultValues": "https://example.org/Person/normalized_gender"`:
This gives the schema node ID under the context node that will receive
the lookup result.

The valueset lookup will run for each `gender` node. In the above
example, the `gender` node has value `female`, so this will be looked
up in the `gender` valueset, and the result will be `1`. This result
will be inserted as a new node `normalized_gender` under the context
node, which is the root node for the `Person`. When exported as a CSV 
file, this will result in:

```
person_id,gender,normalized_gender
1,female,1
2,male,2
3,male,2
4,unknown,0
5,F,1
6,M,2
7,F,1
```
