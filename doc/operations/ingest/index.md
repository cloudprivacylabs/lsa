---
title: "Data Ingestion"
menu: 
  main:
    weight: 20
    parent: docs
---

# Data Ingestion


Ingest Object
* An `Object` contains a set of named attributes. An object can be used to represent a JSON object containing key-value pairs, or an XML element containing other elements. In the schema node, if the property ingest as property is specified as `ls:ingestAs=node` then the data ingestion will behave as ingesting the object as a node. Ingesting an object outputs a graph similar to: <img> where 
each parent node contains the edge label "has" to each of its child nodes. The graph structure of an ingested object type will be similar to <img>

Ingest Object as Edge
* In the schema node, if the property ingest as property is specified as `ls:ingestAs=edge` then the data ingestion will behave as ingesting the object as an edge. Ingesting an `Object` as an edge outputs a graph similar to: <img> where 
the edge label connecting the parent to the child node, is the name of the key in the key-value pair of the child node.

Ingest Array
* An `Array` contains repeated attributes. `Array` attributes can be used to represent JSON arrays, or XML elements (an XML element containing other elements can be represented as both an object and an array). The array definition contains the attribute specification for the array's items. In the schema node, if the property ingest as property is specified as `ls:ingestAs=node` then the data ingestion will behave as ingesting the array as a node. Ingesting an array outputs a graph similar to: <img> where 
each parent node contains the edge label "has" to each of its child nodes. The graph structure of an ingested array type will be similar to <img>

Ingest Array as Edge
* In the schema node, if the property ingest as property is specified as `ls:ingestAs=edge` then the data ingestion will behave as ingesting the array as an edge. Ingesting an array as an edge outputs a graph similar to: <img> where 
the edge label connecting the parent to the child node, is name of the key in the key-value pair of the child node

Ingest Value
* A `Value` is simply a string of bytes whose content will be interpreted by a program. The actual underlying value may have parts when interpreted (such as a date field with year, month, day parts), but as long as the schema processing is concerned, the Value field is atomic. The graph structure of an ingested value type will be similar to <img>

Ingest Value as Edge
* In the schema node, if the property ingest as property is specified as `ls:ingestAs=edge` then the data ingestion will behave as ingesting the value as an edge. Ingesting an array as an edge outputs a graph similar to: <img> where 
the edge label connecting the parent to the child node, is name of the key in the key-value pair of the child node
