---
title: Composition and Slicing
---
# Composition and Slicing

Use a schema base to describe the data elements for a particular data
capture scenario, and compose it with different overlays to add
semantics that may differ based on data source. The composed schema is
a *schema variant*. The semantic annotations classify data, add
processing instructions, or add use-case specific metadata. 

If you have a schema variant, you can *slice* it to separate into a
schema with fewer annotations, and one or more overlays.

## Composition 

Composition operation combines a schema and overlays to create a new
variant of the schema, or combines multiple overlays to create a new
overlay that is a combination of several overlays. A schema cannot be
composed with another schema.

When composing schema layers, all layers must agree on the same
nonempty `valueType`. That is, composing two layers with different
nonempty `valueType` values is not valid. That means, a layer with no
`valueType` can be composed with any other layer.

The following are valid compositions:

 * `Schema(valueType=A) + Overlay_1() + Overlay_2(valueType=A)`
The result is a schema with `valueType=A`.
 * `Overlay_1(valueType=A) + Overlay_2(valueType=A)`
The result is an overlay with `valueType=A`.
   
### Composing Terms
    
A key part of the algorithm is composing the values of terms defined
in the attribute. It is possible for an implementation to define terms
that specify term-specific composition methods. The term composition
methods are as follows:
 
  * Set composition: This is the default term composition method. The
    set composition of two terms is the set union of their values. For
    example:

| Value 1 | Value 2 | Composition |
| ------- | ------- | ----------- |
| A       | [A, B]  | [A, B]      |
| A       | B       | [A, B]      |
| A       | [B, C]  | [A, B, C]   |

  * List composition: The list composition of two terms is the
    concatenation of the values of the second term to the first. For
    example:
    
| Value 1 | Value 2 | Composition |
| ------- | ------- | ----------- |
| A       | [A, B]  | [A, A, B]   |
| A       | B       | [A, B]      |
| A       | [B, C]  | [A, B, C]   |

  * Override composition: The value of the second term overrides the
    first. For example:
    
| Value 1 | Value 2 | Composition |
| ------- | ------- | ----------- |
| A       | [A, B]  | [A, B]      |
| A       | B       | [B]         |
| A       | [B, C]  | [B, C]      |

   * No composition: The value of the first term remains. For example:
   
| Value 1 | Value 2 | Composition |
| ------- | ------- | ----------- |
| A       | [A, B]  | [A]         |
| A       | B       | [A]         |
| A       | [B, C]  | [A]         |
      

### Algorithm

This algorithm recursively composes the attributed of the `source`
layer into `target` layer. Each attribute of the `source` layer is
looked up in the `target` by using the attribute ID. If the `target`
has an attribute with a matching ID, the two nodes are composed. If
the `target` does not have a matching attribute, then the new
attribute is added.

#### Examples

Overlay adds `valueType` to the base:

Base:
{{<highlight json>}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "b": {
        "@type": "Object"
     }
  }
}
{{</highlight>}}

Overlay:
{{<highlight json>}}
{
  "@id": "b",
  "@type": "Object",
  "valueType": "xs:date"
}
{{</highlight>}}

Result:
{{<highlight json>}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "b": {
        "@type": "Object",
        "valueType": "xs:date"
     }
  }
}
{{</highlight>}}


Overlay adds new attribute to the base:

Base:
{{<highlight json>}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "b": {
        "@type": "Object"
     }
  }
}
{{</highlight>}}

Overlay:
{{<highlight json>}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "c": {
       "@type": "Value"
    }
  }
}
{{</highlight>}}

Result:
{{<highlight json>}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "b": {
        "@type": "Object"
     },
     "c": {
        "@type": "Value"
     }
  }
}
{{</highlight>}}


## Slicing

Slicing operation creates new layers from an existing layer by
selecting a subset of the terms. It uses an `accept` operation that
selects the terms that will be included in the output.

### Algorithm

{{<highlight reStructuredText>}}
SliceAttribute(attr,accept)
  newAttribute:= new Attribute
  For each (term, value) in attr
    If accept(term)
      newAttribue[term]=value
  
  If attr is one of Object, Array, Composite, Polymorphic
    For each nestedComponent under attr
      SliceAttribute(nestedComponent,accept)
      
  If newAttribute is not empty, return newAttribute
{{</highlight>}}

### Example

Consider the following layer:

{{<highlight json>}}
"attributes": {
  "attr1": {
     "@type": "Value",
     "format": "url",
     "privacyClassifications": ["PII"]
  },
  "attr2": {
    "@type": "Object",
    "attributes": {
       "attr3": {
         "@type": "Value",
         "privacyClassifications": ["BIT"]
       }
    }
  }
}
{{</highlight>}}

Slicing this schema with an `accept` function that only accepts
`attributes`, `items`, `allOf`, `oneOf`, and `reference`:

{{<highlight json>}}
"attributes": {
  "attr1": {
     "@type": "Value",
  },
  "attr2": {
    "@type": "Object",
    "attributes": {
       "attr3": {
         "@type": "Value"
       }
    }
  }
}
{{</highlight>}}

Slicing this schema with an `accept` function that accepts `format`:

{{<highlight json>}}
"attributes": {
  "attr1": {
     "@type": "Value",
     "format": "url",
  }
}
{{</highlight>}}

Slicing this schema with an `accept` function that accepts `privacyClassifications`:

{{<highlight json>}}
"attributes": {
  "attr1": {
     "@type": "Value",
     "privacyClassifications": ["PII"]
  },
  "attr2": {
    "@type": "Object",
    "attributes": {
       "attr3": {
         "@type": "Value",
         "privacyClassifications": ["BIT"]
       }
    }
  }
}
{{</highlight>}}
