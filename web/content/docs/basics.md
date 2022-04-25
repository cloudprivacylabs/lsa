---
title: Basic Concepts
---

# Motivation: Schema Layers

Traditional schemas define the shape of data (attributes, nesting,
constraints, etc.) The following is a JSON schema:


{{< highlight json >}}
{
   "type": "object",
   "properties": {
      "firstName": {
         "type": "string"
      },
      "age": {
         "type": "number"
      }
   }
}
{{</highlight>}}

It defines an object containing two attributes, `firstName` which is a
string value, and `age` which is a number. A data object that conforms to a schema is an **instance** of that
schema. The following JSON object is an instance of the above schema:

{{< highlight json >}}
{
  "firstName": "John",
  "age": 21
}
{{</highlight>}}

Schemas are primarily used for data validation and automated code
generation. Once data objects are validated using a schema, they can
be processed easily using native types without further validation. In
the above example, a data processing application can convert the
`firstName` and `age` to a platform-specific value after the object is
validated.

Schemas usually do not contain much semantic information. For
instance, the schema above does not include the fact that `firstName`
and `age` are personally identifiable information. We can `Overlay`
that information onto the JSON schema using another layer:

{{< highlight json >}}
{
   "type": "object",
   "properties": {
      "firstName": {
         "x-ls": {
            "privacyFlags": "PII"
         }
      },
      "age": {
         "x-ls": {
            "privacyFlags": "PII"
         }
      }
   }
}
{{</highlight>}}

When composed with the original schema, this gives:

{{< highlight json >}}
{
   "type": "object",
   "properties": {
      "firstName": {
         "type": "string",
         "x-ls": {
            "privacyFlags": "PII"
         }
      },
      "age": {
         "type": "number",
         "x-ls": {
            "privacyFlags": "PII"
         }
      }
   }
}
{{</highlight>}}

This composite schema now contains metadata that defines some semantic
attributes of the underlying data. Using this schema, we can write a
data ingestion algorithm that represent data elements as the nodes of
a labeled property graph:


{{<figure src="example-lpg1.png" class="text-center my-3">}} 

Note that the schema information is also embedded into the ingested
graph.

We can add processing directives to control how data elements are
ingested. For instance, the following overlay:

{{< highlight json >}}
{
   "type": "object",
   "properties": {
      "firstName": {
         "x-ls": {
            "ingestAs": "property"
         }
      },
      "age": {
         "x-ls": {
            "ingestAs": "edge"
         }
      }
   }
}
{{</highlight>}}

results in the following graph:

{{<figure src="example-lpg2.png" class="text-center my-3">}} 


Layered schema architecture extends the idea of overlaying semantic
layers onto a schema base. The architecture allows dealing with
different data sources that implement the same basic schema with
variations. Different sets of overlays can be composed with a base
schema to add semantic annotations that are layers used to process
data.

The canonical model for layered schemas is expressed using labeled
property graphs. Any textual input describing that graph can be used
as a layered schema. Because of this, layered schemas can be defined
by importing existing JSON/XML schemas, CSV specifications, or by
using more direct representations such as graph JSON objects and
JSON-LD documents. Layered schemas can describe complex data
structures that contain nested data fields, cyclic references, and
polymorphism.

## JSON-LD Representation

A JSON-LD layered schema looks like this:

{{< highlight json >}}
{
  "@context": "https://lschema.org/ls.json",
  "@type": "Schema",
  "valueType": "Person",
  "layer": {
    "attributes": {
      "firstName": {
         "@type": "Value",
         "valueType": "xsd:string",
      },
      "age": {
         "@type": "Value",
         "valueType": "json:number"
      }
    }
  }
}
{{</highlight>}}

This schema defines a data type `Person` containing two `Value`
attributes. A `Value` attribute holds a data value represented as a
string. `valueType` annotation defines the data type to interpret the
value stored for the attribute. 

An `Overlay` can be defined to add new data fields and annotate
existing ones. For instance, the following overlay adds the `lastName`
field and `Identifying` terms using the [Data Privacy Vocabulary](https://dpvcg.github.io/dpv/) to
first and last names.

{{< highlight json >}}
{
  "@context": "https://lschema.org/ls.json",
  "@type": "Overlay",
  "valueType": "Person",
  "layer": {
    "attributes": {
      "firstName": {
        "@type": "Value",
        "http://www.w3.org/ns/dpv##hasPersonalDataCategory": "http://www.w3.org/ns/dpv##Identifying"
      },
      "lastName": {
        "@type": "Value",
        "http://www.w3.org/ns/dpv##hasPersonalDataCategory": "http://www.w3.org/ns/dpv##Identifying"
      }
    }
  }
}
{{</highlight>}}

By composing a schema base with zero or more overlays, a **schema
variant** can be constructed. Different schema variants can be used to
ingest and process data from different sources. The schema variant
for the above schema and overlay is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/ls.json",
  "@type": "Schema",
  "valueType": "Person",
  "layer": {
    "attributes": {
      "firstName": {
        "@type": "Value",
        "valueType": "xsd:string",
        "http://www.w3.org/ns/dpv##hasPersonalDataCategory": "http://www.w3.org/ns/dpv##Identifying"
      },
      "lastName": {
        "@type": "Value",
        "http://www.w3.org/ns/dpv##hasPersonalDataCategory": "http://www.w3.org/ns/dpv##Identifying"
      },
      "age": {
        "@type": "Value",
        "valueType": "json:number"
     }
    }
  }
}
{{</highlight>}}
