---
title: Model and Syntax
---

# Schema Model

A layered schema is a labeled property graph that defines a data
structure. The nodes of the graph represent attributes (data
elements), and the edges between those nodes represent relationships
between the attributes. Layered schemas and overlays can be
represented using graph JSON, graph YAML, JSON-LD. Layered schemas can
also be imported from JSON schemas and CSV specifications.

The nodes of a layered schema graph contains *labels* that represent
the node type (not the value type), and *properties* that are semantic
annotations for the node. The edges of a layered schema graph contains
a *label* that represents the relationship between the nodes, and
*properties* that are semantic annotations for the edge.

**Name spaces**: The JSON-LD specification of layered schemas use the
`https://lschema.org/` namespace. Some other common namespaces such as
`xsd` ( http://www.w3.org/2001/XMLSchema/ ), `json`
( https://json-schema.org/ ) are also recognized by the LSA tooling.

## Schema/Overlay Header

 A `Schema` or an `Overlay` node is the root node of a layered schema. The
 schema/overlay root node is connected to the root node of the layer
 with a `layer` edge. The schema/overlay node defines the schema and
 any metadata related to the schema. The schema/overlay layer root
 node defines the root node of the data object.
 
{{<figure src="schemaroot.png" class="text-center my-3">}} 

**Labels**

`https://lschema.org/Schema`
: This is a schema node.

`https://lschema.org/Object`
: This is an object node that contains other attributes.

`https://lschema.org/Attribute`
: This is a schema attribute node. When a layered schema is processed,
  the `Attribute` label is added to all the attribute nodes.

**Properties**

`id`
: The node identifier. For the schema/overlay root node, this is the
  schema/overlay identifier. For attribute nodes, this is the
  attribute identifier. A non-empty id is required for every attribute
  node of a schema or an overlay. Note the use of `valueType` with a
  namespace as the layer root id. This is because the attribute node
  id for a layer root is the type id of an entity. For example, the
  layer root node defines a `Person` object (`valueType`), and the
  type id is `https://test.org/Person` (layer root node `id`).

`https://lschema.org/encoding`
: Optional annotation that specifies the data encoding. If specified, data
  processed using this schema will be read using the defined encoding.
  If unspecified, the native encoding of the platform will be used.
  
`https://lschema.org/valueType`
: Required annotation that specifies the type name of the object
  defined by the schema. The `valueType` specified at the schema root
  node is copied to the layer root node when schema is loaded. This
  annotation is optional for overlays. If `valueType` is specified for
  an overlay, it can only be composed with a schema that has the same
  `valueType`.

`https://lschema.org/entityIdFields`
: Optional annotation that specifies the unique identifier(s)
attribute ids for the entity. Contents of this property can be a
string value, or string array. Note that this is given under `layer`.


The JSON-LD representation for a schema is as follows:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://schema_id",
  "valueType": "Person",
  "encoding": "utf-8",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "entityIdFields": "https://test.org/Person/id",
    "attributes": {
    ...
    }
}
{{</highlight>}}

An overlay compatible with this schema is as follows:
{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Overlay",
  "@id": "https://ovl_id",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
    ...
    }
}
{{</highlight>}}

## Overlay Specific Syntax

Overlays can include these additional information:

### `compose`

`https://lschema.org/compose` specifies how to combine annotations of
the overlay attributes with the schema. Possible values are:

`set` 
: All terms will be combined as a set with the composed schema
annotations.
`list`
: The terms of this overlay will be added to the composed schema
  annotations. Duplications may occur, ordering is preserved.
`override`
: The terms of this overlay overrides the schema terms.
  
As an example, consider the overlay:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Overlay",
  "@id": "https://ovl_id",
  "valueType": "https://example.org/Person",
  "compose": "override",
  "layer": {
    "@type": "Object",
    "@id": "https://example.org/Person",
    "attributes": {
       "https://example.org/Person/firstName": {
         "@type": "Value",
         "pattern": "[a-zA-Z]+"
       }
    }
  }
}
{{</ highlight >}}
This overlay overrides the `pattern` in the composed schema for `firstName`. 

### `attributeOverlays`

This is a convenient way to compose semantics for individual attribute
without specifying the full path. Attributes listed under `layer` term
must match the path of the underlying schema to modify an
attribute. Attributes listed under
`https://lschema.org/attributeOverlays` only match by attribute id.

As an example, consider the following overlay:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Overlay",
  "@id": "https://ovl_id",
  "valueType": "https://hl7.org/fhir/Patient",
  "attributeOverlays": [
    {
       "@id": "https://hl7.org/fhir/Patient/name/*/given/*",
       "@type": "Value",
      "pattern": "[a-zA-Z]+"
    }
  ]
}
{{</ highlight >}}

This overlay defines the `pattern` for patient given name, which is an
array field under `/name/*/given`. Without `attributeOverlays`, the
only way to define this overlay is to specify all attributes in the
path: `Patient`, `Patient/name`, `Patient/name/*`, and
`Patient/name/*/given`.

## Attributes

Attributes are data elements of the object described by the
schema. Each attribute must have a type, and an identifier that is
unique within the schema. Attribute types are:

  * [Value](#value)
  * [Object](#object)
  * [Array](#array)
  * [Reference](#reference)
  * [Composite](#composite)
  * [Polymorphic](#polymorphic)

### `Value`

A `Value` is a string of bytes whose content will be interpreted by a
program. The actual underlying value may have parts when interpreted
(such as a date field with year, month, day parts), but as long as the
schema processing is concerned, a `Value` field is atomic. A `Value`
attribute cannot have child attributes.

The following schema defines an object containing a value attribute:

{{<figure src="valueschema.png" class="text-center my-3" >}} 

The corresponding JSON-LD schema is:


{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
      "https://test.org/firstName": {
        "@type": "Value",
        "attributeName": "firstName"
      }
    }
}
{{</highlight>}}

The `attributeName` annotation is used during data ingestion or data
export to name the value. It may correspond to a JSON object key, or
the column name of tabular data.

### `Object`

An `Object` contains a set of named attributes. An object can be used
to represent a JSON object containing key-value pairs, an XML element
containing other elements, or a row of tabular data. An object
attribute can have `attributes` which is a set of attributes where
order is not significant, or `attributeList`, which is a set of
attributes where order is significant.

The following schema shows an object using `attributes`:
{{<figure src="objectschema1.png" class="text-center my-3" >}} 

The corresponding JSON-LD schema is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
      "https://test.org/firstName": {
        "@type": "Value",
        "attributeName": "firstName"
      },
      "https://test.org/lastName": {
        "@type": "Value",
        "attributeName": "lastName"
      }
    }
}
{{</highlight>}}

Below is the same schema using `attributeList`:

{{<figure src="objectschema2.png" class="text-center my-3" >}} 

And its JSON-LD representation is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributeList": [
      {
        "@id": "https://test.org/firstName",
        "@type": "Value",
        "attributeName": "firstName"
      },
      {
        "@id": "https://test.org/lastName",
        "@type": "Value",
        "attributeName": "lastName"
      }
    ]
  }
}
{{</highlight>}}

The `attributeIndex`es are added to the attribute nodes when schema is
loaded. The ordering of attributes and the `attributeIndex` values for
an object with `attributes` edges is nondeterministic. The ordering of
attributes for `attributeList` edges is fixed by the order in which
the object is defined.

### `Array`

An `Array` contains repeated attributes that share the same definition
(which can be polymorphic). Array attributes can be used to represent
JSON arrays, or XML elements (an XML element containing other elements
can be represented as both an object and an array). The array
definition contains the attribute specification for the array items.


{{<figure src="arrayschema.png" class="text-center my-3" >}} 

The JSON-LD schema for this is:


{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
      "https://test.org/addresses": {
        "@type": "Array",
        "attributeName": "addresses",
        "arrayElements": {
           "@id": "https://test.org/address",
           "@type": "Object"
        }
      }
    }
  }
}
{{</highlight>}}


### `Reference`

A `Reference` points to another entity defined by a schema or schema
variant. How the reference is resolved is implementation
dependent. The reference can be:

  * A reference to schema or schema variant using schema id
  * The value type of the referenced object, which is then resolved
    into a schema variant using a `Bundle`.

The reference implementation of layered schemas uses value type
references.

{{<figure src="referenceschema.png" class="text-center my-3" >}} 

The JSON-LD schema for this is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
      "https://test.org/address": {
        "@type": "Reference",
        "attributeName": "address",
        "ref": "Address"
      }
    }
  }
}
{{</highlight>}}

When compiled, the `Reference` in the schema will be resolved by
looking up the schema variant for `Address` type. Then, the reference
node in the schema will be composed with the root node of the
`Address` schema. For example, consider the following address schema:

{{<figure src="addressschema.png" class="text-center my-3" >}} 

After compilation, the schema looks like:

{{<figure src="compiled_address_schema.png" class="text-center my-3" >}} 


### `Composite`

A `Composite` attribute is a composition of other attributes. When a
schema containing composite attributes is compiled, all such
attributes are converted into `Object`s by combining the contents of its
components.

{{<figure src="compositeschema.png" class="text-center my-3" >}} 

The JSON-LD schema for this is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/person_schema",
  "valueType": "Person",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Person",
    "attributes": {
      "https://test.org/address": {
        "@type": "Composite",
        "attributeName": "address",
        "allOf": [
          {
            "@id": "https://test.org/address/base",
            "@type": "Reference",
            "ref": "BaseAddress"
          },
          {
            "@id": "https://test.org/address/state",
            "@type": "Value",
            "attributeName": "state"
          }
        ]
      }
    }
  }
}
{{</highlight>}}

When compiled, the attributes of `BaseAddress` and `state` will be
combined to make up a new `Object` node in place of the `Composite`
node.


### `Polymorphic`

A `Polymorphic` attribute can be one of the types of attributes listed
in its definition. The reference implementation of data ingestion
algorithm relies on attribute validators to determine the correct type
of the object being ingested, but other implementations may choose to
ingest data using different approaches.

{{<figure src="polymorphicschema.png" class="text-center my-3" >}} 

The JSON-LD schema for this is:

{{< highlight json >}}
{
  "@context": "https://lschema.org/v1/ls.json",
  "@type": "Schema",
  "@id": "https://test.org/account_schema",
  "valueType": "Account",
  "layer": {
    "@type": "Object",
    "@id": "https://test.org/Account",
    "attributes": {
      "https://test.org/Account/owner": {
        "@type": "Polymorphic",
        "attributeName": "owner",
        "anyOf": [
          {
            "@id": "https://test.org/Account/owner/person",
            "@type": "Reference",
            "ref": "Person"
          },
          {
            "@id": "https://test.org/Account/owner/organization",
            "@type": "Reference",
            "ref": "Organization"
          }
        ]
      }
    }
  }
}
{{</highlight>}}
