---
title: Terms Reference
menu: 
  main:
    weight: 10
    parent: docs
---

# Terms Reference

## Layer terms

These terms are used for defining schemas or overlays.

### Top-level terms

[https://lschema.org/Schema](/Schema)

[https://lschema.org/Overlay](/Overlay)

[https://lschema.org/SchemaVariant](/SchemaVariant)

[https://lschema.org/targettype](/targettype)

[https://lschema.org/bundle](/bundle)

[https://lschema.org/charcaterEncoding](/charcaterEncoding)

[https://lschema.org/layer](/layer)



### Attribute Types

[https://lschema.org/Attribute](/Attribute)
: Defines an attribute node. The actual type of the attribute node is
    defined by one of the following terms. Defining a schema node with
    one of the following terms implies that the node is an `Attribute`
    node.

[https://lschema.org/Value](/Value)
: Defines an attribute that is a terminal node that has a value, such
  as a JSON key-value pair, an XML attribute, or an XML element that
  does not contain other elements.

[https://lschema.org/Object](/Object)
: Defines an attribute that groups other attributes as an ordered or
  unordered set, such as a JSON object, or an XML element containing
  other elements.
 
[https://lschema.org/Array](/Array)
: Defines an attribute that groups other attributes as an ordered and
  indexed set, such as a JSON array, or repeating XML elements.
 
 [https://lschema.org/Reference](/Reference)
: Defines an attribute whose type definition is in a different schema.

[https://lschema.org/Composite](/Composite)
: Defines an attribute that is composed of multiple components. 
 
[https://lschema.org/Polymorphic](/Polymorphic)
: Defines an attribute whose type can be one of several types.

### Annotations

[https://lschema.org/asProperty](/asProperty)

[https://lschema.org/asPropertyOf](/asPropertyOf)

[https://lschema.org/attributeIndex](/attributeIndex)

[https://lschema.org/attributeName](/attributeName)

[https://lschema.org/attributeValue](/attributeValue)

[https://lschema.org/defaultValue](/defaultValue)

[https://lschema.org/description](/description)

[https://lschema.org/entityId](/entityId)

[https://lschema.org/label](/label)


### Validation Terms

[https://lschema.org/validation/enumeration](/validation/enumeration)

[https://lschema.org/validation/jsonFormat](/validation/jsonFormat)

[https://lschema.org/validation/pattern](/validation/pattern)

[https://lschema.org/validation/required](/validation/required)
