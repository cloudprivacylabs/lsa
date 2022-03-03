---
title: "Layered Schema Architecture"
subtitle: "Semantic interoperability for structured data"
menu: 
  main:
    title: Home
    weight: 1

---

{{<figure src="layers_ingestion.png" class="text-center my-3">}} 

Layered schema technology enables interoperability between
heterogeneous systems by harmonizing data semantics for data capture,
processing, and exchange. It uses a schema to define data elements and
interchangeable layers (overlays) to define semantics that can vary
between different sources. Data can be ingested from disparate systems
that use different conventions and ontologies, and then converted into
a knowledge graph.

### Use Case: Semantic Data Warehouse

A traditional data warehouse uses source specific ETL scripts to
normalize and ingest data. Maintanence of such scripts heavily rely on
organically developed tools and know-how. A semantic data warehouse
based on the layered schema architecture replaces such ETL scripts
with source specific schema variants. Each schema variant is composed
of a schema base that describes the common data elements with a set of
overlays. Overlays add the metadata necessary to adjust the ingestion
process to account for the variations specific to a data
source. Ingested data is stored as a knowledge graph. The knowledge
graph can be further processed using a terminology library and
semantic pipelines for analytics and AI applications.

{{<figure src="dw-fanin-fanout-sm.png" class="text-center my-3">}} 

### Use Case: Interoperability Across Domains

There are multipls domain-specific data exchange standards, such as
FHIR for health data exchange. An application working with data from
multiple domains usually have to continuously translate between
different standards. An applications built using the layered schema
architecture uses standard schemas available for each domain, with
layers defining translations and crosswalks between domain-specific
standards.

### Use Case: Executable Data Exchange Policies



### Data Display and Entry


