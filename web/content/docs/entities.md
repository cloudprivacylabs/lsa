---
title: Entities
---

# Entities

Every schema defines an entity. The type of the entity is specified in
the `valueType` of the root node of the layer. An entity may have a
unique identifier. The unique identifier contains one or more
attributes. These attribute must not be included in array fields.

Below is a schema for a `https://example.org/Person` whose ID field is
`https://example/org/Person/id`:

{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/Person/schemaBase",
    "@type": "Schema",
    "valueType": "https://example.org/Person",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/Person",
        "entityIdFields": "https://example.org/Person/id",
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
                "@id": "http://example.org/Person/id",
                "@type": "Value",
                "attributeName": "id"
            },
            {
                "@id": "http://example.org/Person/contact",
                "@type": "Array",
                "attributeName": "contact",
                "arrayElements": {
                    "@type": "Reference",
                    "@id": "http://example.org/Person/contact/items",
                    "ref": "https://example.org/Contact"
                }
            }
        ]
    }
}
{{</highlight>}}

When a schema is compiled. the entity root nodes are marked with
`https://lschema.org/entitySchema` term that gives the schema ID for
the entity. The following graph shows the compiled schema for
`https://example.org/Person` entity. The root nodes for `Person` and
`Contact` entities are marked with `https://lschema.org/entitySchema`
annotation showing the schema for those entities. The `Person` entity
also declares the `https://example.org/Person/id` field as an entity
id. All the attributes under the `Person` root node up until the
`Contact` root node belong to the `Person` object. All the attributes
under the `Contact` root node belong to the `Contact` object.

![Compiled Person and Contact Entities](person_compiled.png)

When data elements are ingested, all nodes that are instances of
entity root nodes with nonempty `entityIdFields` annotation will have
the `https://lschema.org/entityId` field initialized to the entity
ID. If the `entityIdFields` is string value, then `entityId` will be
initialized to the content of that attribute. If `entityIfFields` is an
array, the `entityId` will be initialized as an array where every
element initialized from the value of the corresponding ID field.

## Linking entities

A schema may specify the creation of additional links between ingested
entities. Consider the following entity named `A`:

{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/A/schema",
    "@type": "Schema",
    "valueType": "https://example.org/A",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/A",
        "entityIdFields": "http://example.org/A/id",
        "attributes": {
            "http://example.org/A/id": {
               "@type": "Value"
            }
        }
    }
}
{{</ highlight>}}

Here is another entity `B` that contains a foreign key containing the
identifier for an `A` entity:

{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/B/schema",
    "@type": "Schema",
    "valueType": "https://example.org/B",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/B",
        "attributes": {
            "https://example.org/B/a_id": {
              "@type": "Value"
            }
        }
    }
}
{{</highlight>}}

When instances of `A` and `B` are ingested, we would like to create a
link from the `A` instance to its `B` instance. The following addition
to the `B` schema does this:

{{< highlight json >}}
{
    "@context": "https://layeredschemas.org/ls.json",
    "@id": "http://example.org/B/schema",
    "@type": "Schema",
    "valueType": "https://example.org/B",
    "layer": {
        "@type": "Object",
        "@id": "http://example.org/B",
        "attributes": {
            "https://example.org/B/a_id": {
              "@type": "Value"
            },
            "http://example.org/B/owner": {
                "@type": "Reference",
                "reference": "https://example.org/A/schema",
                "fk": "https://example.org/B/a_id",
                "link": "to",
                "ingestAs": "edge",
                "label": "owner",
                "multi": false
           }
        }
    }
}
{{</highlight>}}

Here, the `owner` field defines a reference to `A` using its schema
ID. The other fields define how the link is constructed, and how the
`A` instance is found:

  * `fk` specifies the foreign key field contained in `B` that
    contains the `A` entity id. If the `A` entity has multiple
    identifiers, the `fk` field must specify an array of fields.
  * `link` specifies the link direction. If `to`, the link is from `A`
    to `B`. If `from`, the link is from `B` to `A`.
  * `ingestAs` specifies how the link should be established. If
    `edge`, an edge is added between the root node of `B` and the root
    node of `A`. If `node`, a node is created for the `owner` field,
    and that node is connected to the root node of `A`.
  * `label` specifies the edge label. If omitted,
    `https://lschema.org/has` is used.
  * `multi` is `true` and there are multiple instances of `A` with the
    given identifier, the `B` entity is linked to all those matching
    `A` entities. If `multi` is false and multiple `A` entities match,
    an error is returned.
