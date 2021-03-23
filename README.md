[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/lsa?status.svg)](https://godoc.org/github.com/cloudprivacylabs/lsa)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudprivacylabs/lsa)](https://goreportcard.com/report/github.com/cloudprivacylabs/lsa)

# Table of Contents

- [Layered Schemas](#layered-schemas)
  * [Example Operation](#example-operation)
- [What's in this Go Module?](#whats-in-this-go-module)
- [Schemas Layers](#schema-layers)
  * [@context](#context)
  * [Examples](#examples)
  * [Semantics](#semantics) 
  * [attributes](#attributes)
- [Schemas](#schemas)
- [Addressing Layers and Schemas](#addressing-layers-and-schemas)
  * [Strong Reference: Hash](#strong-reference-hash)
  * [Weak Reference: IRI](#weak-reference-iri)
- [Schema Bundles](#schema-bundles)
  * [Decentralized Strong Reference](#decentralized-strong-reference)

# Layered Schemas

Layered schema architecture is developed to work with unified data
semantics regardless of the data capture locale, data representation,
format, or terminology. Traditional schemas (such as JSON/XML schemas)
describe the structure of data with limited semantic
information. Layered schemas use structural constraints similar to
traditional schemas and open-ended semantic annotations to describe
data. Some elements of a layered schema are:

  * Constraints: Required attributes, length limits,...
  * Format: Phone number with area code, date/time,...
  * Language: English, Spanish,...
  * Privacy classifications: Personally identifiable information, sensitive information,...
  * Retention policies: Attributes must be cleared after a period,...
  * Provenance: Data source, signatures, ...

A schema can be "sliced" into multiple layers with each layer
including only some of the annotations. These layers can be replaced
with other layers representing variations in stuctural constraints,
different languages or notationan differences based on locale,
different security constraints based on context, etc.

The following figure illustrates an application stack based on layered
schema architectur. In this application, data objects (JSON, XML, CSV,
etc.) are "ingested" using an adapter. These adapters convert incoming
data into a common data model containing the ingested data element and
their associated attribute information coming from the schema (this
repository includes adapters for CSV and JSON data.) The application
processes the data using this common data model, and then exports it
using another adapter, which may use another schema for the operation.
For example, an application can ingest tabular data from a CSV file,
and then can generate a verifiable credential. Same
application can also ingest a JSON data object using the same terms to
generate a verifiable credential using the same logic.

![Data Processing with Layered Schemas](doc/layered-schemas-stack.png)


Use-cases for layered schemas include:

  * **Data entry:** A layered schema can include overlays that have
    information to auto-generate data entry forms. These can be
    locale-specific overlays including valid options, entry format, UI
    labels and help text, etc.
  * **Data ingestion:** Semantics for data coming from multiple
    sources (APIs, file uploads, data entrty) can be defined using a
    unified vocabulary at the schema bases, and source-specific
    variations can be implemented as different sets of overlays.
  * **Granular privacy/security controls** Different privacy/security
    classifications can be applied to data elements based on the
    overlays used. For instance, data elements with a certain privacy
    classification can be masked in one context, and left untouched in
    another context. 
  * **De-identification** Attributes that might reveal the identity of
    data subjects can be removed from the data. Using different
    overlays, the set of attributes that will be removed from the data
    can be controlled.
  

## Example Operation

A layered schema is a JSON-LD document. It describes the structure of
an object and semantic information for its data elements. Most of the
information/constraints encoded in JSON/XML schemas or CSV files can
be represented using layered schemas.

Consider this JSON schema:

```
{
  "type": "object",
  "properties": {
     "link": {
        "type": "string",
        "format": "uri"
     },
     "num": {
        "type" "number"
    }
  }
}
```

An equivalent layered schema is as follows:

```
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",               // mark this object as a schema layer
  "objectType": "TestObject",     // Name of the object defined by the schema
  "attributes": {                 // Attributes of the objec
    "id1": {                      // Attribute id
       "name": "link",   // Name of the attribute
       "type": "string",          // Type of the attribute (annotation)
       "format": "uri"            // Format of the attribute (annotation)
    },
    "id2": {                      // Attribute id
       "name": "num",    // Name of the attribute
       "type": "number"           // Type of the attribute (annotation)
    }
  }
}
```

We can slice this schema into three layers:
```
// Base layer - structure
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",
  "objectType": "TestObject",
  "attributes": {
    "id1": {
       "name": "link"
    },
    "id2": {
       "name": "num"
    }
  }
}

// Type layer
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",
  "objectType": "TestObject",
  "attributes": {
    "id1": {
       "type": "string"
    },
    "id2": {
       "type": "number"
    }
  }
}

// Format layer
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",
  "objectType": "TestObject",
  "attributes": {
    "id1": {
       "format": "uri"
    }
  }
}
```

The composition of these three layers is the original schema.

Suppose we ingest a data object using this schema:

```
{
  "link": "http://example.com?id=12345",
  "num": 1,
  "extraField": "test"
}
```

When ingested, this data object will be annotated using schema attributes:

```
{
  "attributes": {
    "TestObject.link": {                      // Generated attribute ID
     "attributeId": "id1"                     // Link to schema attribute
     "name": "link",                 // Name of the attribute in the input
     "type": "string",                        // Annotation embedded from the schema
     "format": "uri",                         // Annotation embedded from the schema
     "value": "http://example.com?id=12345",  // Value of the attribute
    },
    "TestObject.num": {
     "attributeId": "id2"
     "name": "num",
     "type": "number",
     "value": 1,
    },
    "TestObject.extraField": {                // This field does not exist in the schema
                                              // so there are no schema annotations
     "value": "test",                         // Value of the attribute
     "name": "extraField"            // Name of the attribute
    }
}
```

We can add a new layer to mark the first attribute with `PII`
(personally identifiable information)` flag:

```
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",
  "attributes": {
    "id1": {
       "privacyClassification": "PII"
    }
  }
}
```

Then the ingested document becomes:

```
{
  "attributes": {
    "TestObject.link": {
     "attributeId": "id1"
     "name": "link",
     "type": "string",
     "format": "uri",
     "privacyClassification" "PII",            // New annotation from the added layer
     "value": "http://example.com?id=12345",
    },
  ...
```

An application can select all attributes containing `PII` in
`privacyClassifications` and set their values to `null` to
de-identify a data object.

# What's in this Go Module?

This Go module has the following open-source components:

  * `pkg/ls`: This is the core package containing 
    * models for schema layers and schema
    * JSON-LD processing to slice and compose schema layers
    * document model used for data ingestion
  * `pkg/json`: The JSON adapter containing
    * JSON schema to schema layers conversion
    * JSON data ingestion
  * `pkg/csv`: The CSV adapter containing
    * CSV schema description to schema layers conversion
    * CSV data ingestion. This first convers the CSV document to a 
      flat JSON object, and uses the JSON adapter.
  * `layers`: This is a CLI for
    * importing JSON/CSV schemas
    * ingesting JSON/CSV data, and outputting annotated data as JSON-LD
    * Slicing and composing schema layers
    * JSON-LD processing (expand, flatten, frame, etc.)

# Schema Layers

The format of a schema layer is as follows:

```
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "Layer",
  "@id": "http://example.org/someEntity/base",
  "objectType": "someEntity",
  "attributes": {
    "key1": {},
    "key2": {
      "attributes": {
         "key2_1": {}
      }
    },
    "key3": {
      "privacyClassifications": ["https://someOntology/PII"]
    },
    "array": {
      "arrayItems": {
         "type": "string"
      }
    },
    ...
  ],
  
}
```


  * @type: This is a Layer document
  * @id: The ID for the layer
  * objectType: The object described by this layer.

## @context

The @context defines several types and terms. Each term has specific
semantics and algorithms associated with it.

[@context](schemas/layers.jsonld)

 * `Layer` type is used to describe a schema layer. 
 * `objectType`: This is the entity type defined by the schema base
 
 Object structure related terms:
 
 * `attributes` term is used in schema layers. It defines a nested
attribute structure.``
  * `@id` (for each attribute): Specifies the id for the attribute. It
    must be unique in the schema it is defined in.
  * `reference`: Specifies another object referenced by this object
  * `arrayItems`: If the defined attributes is an array, `arrayItems`
    specifies the structure of one element.
  * `allOf`: Specifies composition. The resulting element is the
    combination of the contents of the elements of this term.
  * `oneOf`: Specifies polymorphism. The resulting element is one of
    the elements of this term.
  * `name`: Name of this attribute as it appears in data.
  
Predefined attribute annotations are below. You can include
additional @contexts for cusom annotations.
  
  * `privacyClassification`: A set of flags associated with the
    term. Each flag can belong to an ontology that flags this
    attributes. This can be privacy/risk classifications, blinding
    identity, etc.
  * `encoding`: Character encoding for the attribute.
  * `type`: Data type
  * `format`: Expected data format
  * `pattern`: Expected data pattern
  * `label`: Prompt label when constructing a form for this object
  * `information`: Comments
  * `enumeration`: Enumerated options for the attribute

## Examples

A simple key/value pair is represented as:

```
"attributes": {
   "<key>": {},
   ...
}
```
or

```
"attributes": [
  {
    "@id": "<key>"
  },
  ...
]
```

The `key` is the value assigned to this attribute by the schema
author. Localized names can be given to this key using an overlay with term:

```
{
   "@id": "<key>",
   "name": "name"
}
```

Nested objects can be defined for keys:

```
"attributes": [
  {
    "@id": "name1",
    "name": "name"
  },
  {
    "@id": "obj",
    "attributes": [
       {
         "@id": "name2",
         "name": "name"
       }
    ]
  }
]
```

The above schema defines the following JSON document:

```
{
  "name": "...",
  "obj": {
    "name": "..."
  }
}
```

The `name` and `obj.name` refer to two different attributes with ids
"name1" and "name2" respectively.

An attribute has privacy classification:

```
    {
      "@id": "nfijh9i38ceSa",
      "privacyClassification": ["https://someOntology/PII"]
    }
```

For instance, this can be used to flag PII information based on BIT.

An attribute can be a reference to another schema. References are
open-ended, and they can be a
  
   * Reference using a DRI
   * Reference using target object type
   * Reference using a specific variant of a schema

```
{
   "@id": "patient",
   "reference": "http://someSite/Patient"
}
```

The above defines the field "patient" to be a "Patient" object, whose
schema is given in the `reference` value. This is a reference using
the object type, which does not specify a definite schema, thus an
application specific schema selection must be done. A reference using
a DRI would specify a definite schema.

An attribute can be a nested object:

```
{
   "@id": "nestedObject",
   "attributes": [
      {
        "@id": "nestedAttribute"
      },
      ...
   ]
}
```

An attribute can be an array:

```
{
  "@id": "valueArray",
  "arrayItems": {}
}
```

Instance:

```
"valueArray": [ "value1", "value2", ... ]
```

```
{
  "@id": "objectArray",
  "arrayItems": {
     "attributes": [
        {
          "@id": "key1"
        }
     ]
  }
}
```

Instance:

```
"objectArray": [ 
  { "key1": "value1" },
  { "key2": "value2" }
]
```

An attribute can be the composition of multiple objects:

```
{
  "@id": "p1",
  "allOf": [
    {
      "reference": "http://someObject"
    },
    {
      "attributes": [
        {
           "@id": "attr"
        }
      ]
    }
  ]
}
```

The above construct creates the `p1` attribute by using all
attributes of `http://someobject` and `attr`.

An attribute can be a polymorphic value:

```
{
   "@id": "p2",
   "oneOf": [
     {
      "reference": "http://obj1"
     },
     {
      "reference": "http"//obj2"
     }
    ]
}
```

This describes the `p2` attribute as either `http://obj1` or
`http://obj2`.

## Semantics 

Each term has well-defined semantics that include the meaning and
operations defined for that term. 

### attributes

The term `attributes` is a container where each node with an @id
defines a new attribute. An attribute can be one of:

  * Object: An  `attributes` term defines the nested object structure.
  * Array: An `arrayItems` term defines the structure of each array element.
  * Reference: A `reference` term links to another object. This can be
    a pointer to a schema manifest, or a pointer to an object whose
    schema can be derived based on the current processing context.
  * Composition: An `allOf` term lists the parts of the object
  * Polymorphism: A `oneOf` term lists the possible types.
  * A simple value: If none of the above exists, the value is a simple
    value.
    
The term `attributes` defines a `compose` algorithm that receives two
`attributes` and combines the contents of matching attributes:

Input 1 :
```
"attributes": {
  "k1": {
    "privacyClassification": ["flag1"]
  },
  "k2": {
    "attributes": {
      "k3":{}
    }
  }
}
```

Input 2 :
```
"attributes": {
  "k1": {
    "privacyClassification": [ "flag2" ]
  },
  "k3": {
    "name": "attr3"
  }
}
```
Result:

```
"attributes": {
  "k1": {
    "privacyClassification": [ "flag1", "flag2" ]
  },
  "k2": {
    "attributes": {
      "k3": {
        "name": "attr3"
      }
    }
  }
}
```

During the compose operation:
  * Attributes that are defined as `@set` in the context
    (`privacyClassification`) are combined
  * If a term is defined as `@list`, contents of the second input is
    appended to the first
  * For non-container terms, terms of input1 and input2 are merged,
    with input2 terms overwriting matching input1 terms

# Schemas

A schema specifies one of more schema layers:

```
{
  "@context": "http://schemas.cloudprivacylabs.com/schema.jsonld",
  "@type": "Schema",
  "@id": "schema Id",
  "issuedBy": "...",
  "issuerRole": "...",
  "issuedAt": "...",
  "purpose": "...",
  "classification": "...",
  "objectType": "...",
  "bundle": "...",
  "objectVersion": "...",
  "layers": [
     "layer1",
     "layer2",
     ...
   ]
}
```

The schema combined schema layers to create a schema that is
localized, adopted to a particular context/jurisdiction, and
versioned. It defines the entity type specified by the schema
(objectType), the version of the specification (objectVersion), and
optionally, adds a signature by the schema publisher for the schema
users to validate.


# Addressing Layers and Schemas

Schema layers may refer to other schemas through a `reference`:

```
"attributes": {
  "attr1": {
     "reference": <reference to other object>
  }
}
```

Such references are only valid if the referenced object is a schema
(that is, a `reference` cannot address a layer).


Schemas also refer to other layers:

```
{
  "@type": "Schema",
  "objectType": "SomeObject",
  "layers": [
     <layer1>,
     <layer2>
     ...
  ]
}
```

These references are only valid if the referenced objects are either
layers or schemas with the same `objectType`.

This feature raises the problem of resolving those references
correctly. This is done using a "Schema Registry". A schema registry
decides what schemas and layers will be used to satisfy a given
reference based on the context and the object type.



## Strong reference: Hash

This links a layer or object using a hash of the object:

```
{
  "@type": "Schema",
  "objectType": "SomeObject",
  "layers": [
     "sha256:a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"
  ]
}
```

This is a valid reference only if the referenced object is a layer or
another schema, and it has the same `objectType`.

Similarly:

```
{
  "@type": "Layer",
  "objectType": "SomeObject",
  "attributes": {
    "attr1": {
       "reference": "sha256:a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"
    } 
  }
}
```

This reference is valid only if it points to a schema. That schema can
be for a different type of object.

## Weak reference: IRI

This kind of reference may refer to multiple schemas:

```
{
  "@type": "Schema",
  "objectType": "SomeObject",
  "layers": [
     "http://example.org/SomeObject/layer/v1.2"
  ]
}
```

The schema registry can resolve this link based on its own
configuration. For instance, if a registry allows only unique version
numbers, the above link would resolve to a definite schema. The link
resolution can be dependent on the processing context. For instance,
when processing data in a specific jurisdiction, layers tagged with
that jurisdiction can be selected.


# Schema Bundles

A schema bundle combines the layers necessary to construct all valid
variations of a schema.

```
{
  "@context": "http://schemas.cloudprivacylabs.com/schema.jsonld",
  "@type": "SchemaBundle",
  "references": {
    "http://example.org/SomeObject/layer/v1.2": {
      "reference": "sha256:ab36367212..."
    },
    "http://example.org/SomeObject": {
      "reference": [
        "sha256:74639847fe736...",
        "sha256:84878344...",
        ...
      ]
    },
  ...
  }
}
```

The `references` specifies a set of strong references for each weak
reference. This means that any weak references encountered while
processing data using the schema, only one of the specified objects
can be used.

This structure cryptographically ties a schema to its `bundle`, and
also to all the layers contained in that `bundle`, so registry
users can be sure that only a known set of schema layers are used to
satisfy a weak reference to a schema.

## Decentralized Strong Reference

With decentralized operation it is important to ensure that data
processed using layers from one registry can be interpreted and
processed correctly when another registry is used at some other
context. A schema bundle ensures that all layers used to interpret
data are identical in every context:

```
{
  "@type": "Schema",
  "objectType": "SomeObject",
  "bundle": "sha256:87ab50c3dfec7afff0e5cd0559981665059c65d7a97d5f8374d294535740f534",
  "layers": [
       "http://example.org/SomeObject/layer/v1.2"
  ]
}
```

This schema has a reference to a `bundle` containing the strong
references to schema layers and schemas.
