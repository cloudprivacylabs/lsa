---
title: "Working with Layers: Schemas, Overlays, and Bundles"
---
# Working with Layers: Schemas, Overlays, and Bundles

Let's start by repeating some definitions:

A **schema** is a machine-readable document that describes a data
object.

An **overlay** modifies a schema by adding or changing semantic
annotations for schema attributes, so the schema can be adjusted for a
different use case.

A **schema variant** is a schema composed of a schema base and zero or
more overlays. Note that a schema base itself is a schema variant.

In many applications, data objects are interconnected. Thus, a schema
defining a data object have to refer to other schemas defining the
connected objects. LSA represents such references using the
`Reference` attribute type. For example, an application that deals
with `Person` objects may keep the contact information for each
`Person` as an array of `Contact` objects. In this scenario, we have
two objects: a `Person` that has an array field whose elements are
references to a `Contact` object. This example can be seen in [Github
repository](https://github.com/cloudprivacylabs/lsa/tree/main/examples/contact).

The `Person` schema is:
{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/Person/schemaBase",
    "@type": "Schema",
    "valueType": "https://example.org/Person",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/Person",
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
                "arrayElements": {
                    "@type": "Reference",
                    "@id": "http://example.org/Person/contact/items",
                    "ref": "https://example.org/Contact"
                }
            }
        }
    }
}
{{</highlight>}}

And the `Contact` schema is:
{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/Contact/schema",
    "@type": "Schema",
    "valueType": "https://example.org/Contact",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/Contact",
        "attributes": {
            "http://example.org/Contact/value": {
                "@type": "Value",
                "attributeName": "value"
            },
            "http://example.org/Contact/type": {
                "@type": "Value",
                "attributeName": "type"
            }
        }
    }
}
{{</highlight>}}

When dealing with `Person` data, we need access to both the `Person`
schema and the `Contact` schema. Also, when working with a `Person`
schema variant, we have to use a matching `Contact` variant. A
**bundle** connects those objects for a use case by defining the
schema variants for each schema in question. We may use a different
bundle that links different variants of `Person` and `Contact` for a
different use case.

For example, `person.bundle.json` looks like the JSON file below. It
defines the `https://example.org/Person` variant using
`person.schema.json` file. This particular schema variant will be used
to ingest a `Person` object. The `Contact`s of a `Person` will be
ingested using the variant specified in the same bundle.


{{< highlight json >}}
{
    "variants" : {
        "https://example.org/Person": {
            "schema" : "person.schema.json"
        },
        "https://example.org/Contact": {
            "schema": "contact.schema.json"
        }
    }
}
{{</highlight>}}

Now let's say we would like to add data privacy vocabulary terms to
the underlying schemas. We introduce two overlays for this:

The `person-dpv.overlay.json` defines the `Person` object as an
instance of `dpv:DataSubject`, and adds `dpv:Name` and
`dpv:Identifying` annotations to the `firstName` and `lastName`
fields:

{{< highlight json >}}
{
     "@context": [
       "https://layeredschemas.org/ls.json",
         { 
            "dpv": "http://www.w3.org/ns/dpv#",
             "hasPersonalDataCategory": {
               "@id":"dpv:hasPersonalDataCategory",
               "@type":"@id"
             }
        }
     ],
    "@id": "http://example.org/Person/dpv",
    "@type": "Overlay",
    "valueType": "https://example.org/Person",
    "layer": {
        "@type": [ "Object","dpv:DataSubject"],
        "@id": "http://example.org/Person",
        "attributes": [
            {
                "@id": "http://example.org/Person/firstName",
                "@type": "Value",
                "hasPersonalDataCategory": [ "dpv:Name", "dpv:Identifying" ]
            },
            {
                "@id": "http://example.org/Person/lastName",
                "@type": "Value",
                "hasPersonalDataCategory": [ "dpv:Name", "dpv:Identifying" ]
            }
        ]
    }
}
{{</highlight>}}

Similarly, the `contact-dpv.overlay.json` adds `dpv:TelephoneNumber`
and `dpv:Identifying` annotations to the `value` field of the contact.

{{< highlight json >}}
{
    "@context":[ "https://layeredschemas.org/ls.json",
         { 
            "dpv": "http://www.w3.org/ns/dpv#",
             "hasPersonalDataCategory": {
               "@id":"dpv:hasPersonalDataCategory",
               "@type":"@id"
             }
         }
               ],
    "@id": "http://example.org/Contact/ovl1",
    "@type": "Overlay",
    "valueType": "https://example.org/Contact",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/Contact",
        "attributes": {
            "http://example.org/Contact/value": {
                "@type": "Value",
                "hasPersonalDataCategory": [ "dpv:TelephoneNumber", "dpv:Identifying" ]
            }
        }
    }
}
{{</highlight>}}

We can define a new `person-dpv.bundle.json` by adding these overlays:

{{< highlight json >}}
{
    "variants" : {
        "https://example.org/Person": {
            "schema" : "person.schema.json",
            "overlays" : [
                {
                    "schema" : "person-dpv.overlay.json"
                }
            ]
        },
        "https://example.org/Contact": {
            "schema": "contact.schema.json",
            "overlays" : [
                {
                    "schema" : "contact-dpv.overlay.json"
                }
            ]
        }
    }
}
{{</highlight>}}

Alternatively, we can define a new bundle **based on** another
one. This new bundle only adds new overlays to the base bundle.

{{< highlight json >}}
{
    "base": "person.bundle.json",
    "variants" : {
        "https://example.org/Person": {
            "overlays" : [
                {
                    "schema" : "person-dpv.overlay.json"
                }
            ]
        },
        "https://example.org/Contact": {
            "overlays" : [
                {
                    "schema" : "contact-dpv.overlay.json"
                }
            ]
        }
    }
}
{{</highlight>}}
