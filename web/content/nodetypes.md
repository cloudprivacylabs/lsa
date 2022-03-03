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
