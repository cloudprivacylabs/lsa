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

When ingested, the first row becomes:

{{<figure src="ingested.png" class="text-center my-3" >}} 


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

## Value Set Processing

Value set processing is done using annotations declared on the schema
for a data element. All valueset annotations are in the
`https://lschema.org/vs/` namespace. 

To process value set lookups, LSA tooling scans the nodes of the
schema used to ingest data. All schema nodes that contain
`https://lschema.org/vs/valuesets` or `https://lschema.org/vs/context`
are processed.

The `https://lschema.org/vs/context` annotation gives the schema node
that serves as the anchor node for value set processing. Any nodes
required to do value set lookup can be found under the context node,
and the results of the valueset lookup will be placed under the
context node.

In the below figure, two `Person` objects are ingested. The value set
context is defined as the `https://example.org/Person` node. Thus,
every instance of `https://example.org/Person` node is set as the
valueset context node. That means, data required to perform valueset
lookups are available under these context nodes, and the results of
the valueset lookups will be placed under these context nodes as well.

If the context annotation is not given, then the node containing the
`valuesets` annotation is assumed to be the context node.

In the following example, the schema annotations corresponding to the
`gender` node are:

{{<highlight json>}}
{
   "vsValuesets": "gender",
   "vsContext": "https://example.org/Person",
   "vsResultValues": "https://example.org/Person/normalized_gender"
}
{{</highlight>}}


{{<figure src="context-nodes.png" class="text-center my-3" >}} 


The `https://lschema.org/vs/valuesets` annotations gives one or more
valueset ids to lookup values. In this example, the values will be
looked up in the valueset named `gender`.

The valueset lookup can be performed using one of more values
determined by the `https://lschema.org/vs/requestKeys` and
`https://lschema.org/vs/requestValue` annotations. If neither of these
are present, then the value of the node containing the valueset
annotations is used. In the above example, only the value of the
`gender` node is used for valueset lookup. This example will result in
two valueset lookups: the first lookup with `valuesets: gender` and
value `female`, and the second lookup with `valuesets: gender` and
value `M`. Using the valueset example given above, these will return
`1` and `2` respectively.

The `https://lschema.org/vs/resultValues` annotation determines where
these normalized values will be inserted. In the above exampe, this is
given as `https://example.org/Person/normalized_gender`. This means
that when the lookup is performed and the results are obtained,
instances of `https://example.org/Person/normalized_gender` schema
node will be created under the context node with values set from the
normalized values:

{{<figure src="processed-narrow.png" class="text-center my-3" >}} 

## Composite Values

When dealing with data that may include terms/codes from multiple
ontologies, it may make sense to perform valueset lookupe using
multiple values. For example, consider the following input data:

```
id, code_system, measure_name
1,LOINC,Body height
```

A valueset lookup can be performed on this input data to find the
matching LOINC code for body height, which is 8302-2. The input for
this valueset lookup contains two values: `code_system: LOINC` and
`measure_name: Body height`. The schema for this input looks like:

{{<highlight json>}}
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
                "attributeName": "id"
            },
            "https://example.org/Person/code_system": {
                "@type": "Value",
                "attributeName": "code_system"
            },
            "https://example.org/Person/measure_name": {
                "@type": "Value",
                "attributeName": "measure_name"
            },
            "https://example.org/Person/measure_code": {
                "@type": "Value",
                "attributeName": "measure_code",
                "vsValuesets": "measurements",
                "vsContext": "https://example.org/Person",
                "vsRequestKeys": [ 
                   "code_system", 
                   "measure_name" 
                ],
                "vsRequestValues": [
                   "https://example.org/Person/code_system",
                   "https://example.org/Person/measure_name"
                ],
                "vsResultKeys": [
                   "code"
                ],
                "vsResultValues": [
                   "https://example.org/Person/measure_code"
                ]
            }
        }
    }
}
{{</highlight>}}

`vsValuesets (https://lschema.org/vs/valuesets)` specify the value set
to use for lookup. In this example, it is "measurements".

`vsContext (https://lschema.org/vs/context)` specify the parent node
that contains the information for value set lookup. In this example,
it is the parent `Person` node.

`vsRequestKeys (https://lschema.org/vs/requestKeys)` and
`vsRequestValues (https://lschema.org/vs/requestValues)` are use to
construct a value set lookup request. The number of elements in
`vsRequestKeys` and `vsRequestValues` must be the same. The entries in
`vsRequestKeys` are used as the keys for the request lookup, and the
instance of the schema nodes for matching `vsReqestValues` under
`vsContext` are used as values. In this example, the value set lookup
request will be constructed as `code_system: "LOINC", measure_name:
"Body height"`. The keys `code_system` and `measure_name` are taken
from `vsRequestKeys`. The values for the corresponding keys are taken
from the nodes under `vsContext` that are instances of the schema node
ids given in `vsRequestValues`.

Once the valueset lookup is performed, the response will be used to
create new nodes under the `vsContext` node using a similar
mechanism. The `vsResultKeys (https://lschema.org/vs/resultKeys)` and
`vsResultValues (https://lschema.org/vs/resultValues)` define how new
nodes will be created. In this example, if the valueset lookup returns
`{code_system: LOINC, code: 8302-2}`, the `vsResultKeys` will select
only `code`, and create a new instance of
`https://example.org/Person/measure_code` using the value `8302-2`.

When this document is exported as CSV, the output becomes:

 
```
id, code_system, measure_name, code
1,LOINC,Body height, 8302-2
```

## Building Value Sets

Value sets can be built using a spreadsheet. The following spreadsheet
can be used to convert between languages and their codes (taken from
PCORNet valuesets). Note that the same code is repeated multiple times
if there are multiple descriptive texts for the same language. This
spreadsheet can be used to translate languages entered as text to
language codes. 

<table class="table table-sm table-bordered">
<thead>
<tr><th>CODE</th><th>DESCRIPTIVE_TEXT</th></tr>
</thead>
<tbody>
<tr><td>ACH</td><td>Acoli</td></tr>
<tr><td>ADA</td><td>	Adangme</td></tr>
<tr><td>ADY</td><td>	Adyghe</td></tr>
<tr><td>ADY</td><td>   Adygei</td></tr>
<tr><td>AFR</td><td>	Afrikaans</td></tr>
<tr><td>AIN</td><td>	Ainu</td></tr>
</tbody>
<table>

Possible valueset requests and responses are:

<table class="table table-sm table-bordered">
<thead>
<tr><th>Request</th><th>Response</th></tr>
</thead>
<tbody>
<tr><td><code>{DESCRIPTIVE_TEXT: adyegi}</code> </td><td> <code>{CODE: ADY, DESCRIPTIVE_TEXT: Adyegi}</code></td></tr>
<tr><td><code>{DESCRIPTIVE_TEXT: Ainu}</code> </td><td> <code>{CODE: AIN, DESCRIPTIVE_TEXT: Ainu}</code></td></tr>
</tbody>
</table>

Sometimes the input data is unreliable, and it may be necessary to
conduct a less restrictive search on valuesets. Such behavior can be
controlled using valueset options.

The following valueset will search the input value in the *code*
column first, and then in the *text* column, and return the result
found in the *code* column only. This will handle data that contains
both codes and textual descriptions as the input value.

<table class="table table-sm table-bordered">
<tbody>
<tr><td>options.lookupOrder</td><td>code</td><td>text</td></tr>
<tr><td>options.output</td><td colspan="2">code</td></tr>
<tr><th>code</th><th colspan="2">text</th></tr>
<tr><td>8532</td><td colspan="2">F</td></tr>
<tr><td>8532</td><td colspan="2">Female</td></tr>
<tr><td>8507</td><td colspan="2">M</td></tr>
<tr><td>8507</td><td colspan="2">Male</td></tr>
</tbody>
</table>
