---
title: Composition and Slicing
---
# Composition and Slicing

To adapt a schema for a particular use case, compose it with overlays
that add semantic annotations for that use case. The resulting schema
is a *schema variant*. The semantic annotations classify data, add
processing instructions, or add use-case specific metadata. Use a
different set of overlays to adapt the same schema base for another
use case.

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

This algorithm composes the `source` layer into `target` layer. The
result is the `target` layer. The algorithm recursively processes the
source attributes, find the matching target attribute and composes the
two.

For a given `source` attribute `sourceAttr`, the `path(sourceAttr)`
refers to the sequence of attribute id's from `sourceAttr` to the
layer root. For example:

{{< highlight json >}}
{
  "@id": "a",
  "@type": "Object",
  "attributes": {
     "b": {
       "@type":"Object",
       "attributes": {
         "c": {}
       }
     }
  }
}
{{</highlight>}}

Above, `path(a) = a`, `path(b) = `a.b`, and `path(c) = a.b.c`.

In the below algorithm, an overlay node `o` matches the base layer
node `b` if `path(o)` is a suffix of `path(b)`. 



{{< highlight reStructuredText >}}
ComposeNode(target,source)
  ComposeTerms(target,source)
  For each source attribute node s
    Find target node t such that path(t) has path(s) as a suffix
    ComposeNode(t,s)

  Add all source non-attribute nodes connected to s into t
{{< /highlight>}}

This algorithm allows defining overlays that contains only the leaf
nodes without the intermediate steps. For example:

{{< highlight json >}}
{
  "@type": "Schema",
  "layer": {
    "@type": "Object",
    "attributes": {
      "obj": {
        "@type": "Object",
        "attributes": {
           "nestedAttr": {
              "@type": "Value"
           }
        }
    }
  }
}

{
  "@type": "Overlay",
  "layer": {
    "@type": "Object",
    "attributes": {
      "nestedAttr": {
        "@type":"Value",
        "descr": "description"
      }
    }
  }
}
{{</highlight>}}

The `nestedAttr` in the overlay has path `nestedAttr`, which matches
`obj.nestedAttr`, so the composition becomes:

{{<highlight json>}}
{
  "@type": "Schema",
  "layer": {
    "@type": "Object",
    "attributes": {
      "obj": {
        "@type": "Object",
        "attributes": {
           "nestedAttr": {
              "@type": "Value",
              "descr": "description"
           }
        }
      }
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
