---
title: "Home"
---

Layered schemas technology is a semantic interoperability tool for
data capture, processing, and exchange. It uses a schema to define
data elements and additional layers (overlays) to define open-ended
semantic annotations. These annotations can be used to 

  * classify data using various ontologies, 
  * harmonize variations due to different conventions, locale, and
    vendor implementations,
  * add directives for validation and custom processing, 
  * add metadata.
  

A layered schema defines a mapping from structured data to a knowledge
graph while adding metadata. This enables processing and linking data
from multiple domains independent of representation. Added semantic
information allows interpreting and processing data without the
precise structure of the source data. For example, it is possible to
use layeres schemas to ingest clinical health data, merge it with
wearable device data, and then mark certain attributes as
"identifiable" so they can be excluded from certain data exchange
scenarios.

## Schema Vocabulary

### Node Types

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
