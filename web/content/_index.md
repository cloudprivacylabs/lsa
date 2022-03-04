---
title: "Layered Schema Architecture"
subtitle: "Semantic interoperability for structured data"
features:
  - title: Cross-domain Interoperability
    text: Integrate concepts from multiple domains in reusable layers.
  - title: Semantic Harmonization
    text: Ingest data from disparate sources and transform with semantic pipelines.
  - title: Analytics and AI
    text: Build knowledge graphs, research data sets from heterogeneous data.
---

{{<figure src="layers_ingestion.png" class="text-center my-3">}} 

Layered schema architecture enables interoperability between
heterogeneous systems for data capture, processing, and exchange.  It
uses a schema to define data elements and interchangeable overlays to
add semantic annotations.  Data can be ingested from disparate systems
using different sets of overlays and converted into a knowledge graph.

### Use Case: Interoperability Across Domains

Schemas and vocabularies are usually domain specific.  Achieving
interoperability across domains requires manual mappings, which is
further complicated by variations due to conventions and jurisdictions
The layered schema architecture allows ingesting data while annotating
data elements with mappings to other vocabularies relevant to the use
case. The annotated knowledge graph can then be translated into data
objects for different domains.


{{<figure src="vocab-mapping.png" class="text-center my-3">}} 

### Use Case: Semantic Data Warehouse

A traditional data warehouse uses source specific ETL scripts to
normalize and ingest data. Maintenance of such scripts relies heavily
on internally developed tools and know-how.  A semantic data warehouse
based on the layered schema architecture replaces such ETL scripts
with source specific schema variants, which can be reused in different
scenarios and can be shared.

{{<figure src="dw-fanin-fanout-sm.png" class="text-center my-3">}} 

### Use Case: Privacy-Conscious Data Exchange

Data exchange policies and user consent dictate what types of data can
be exchanged with whom and for what purposes. Traditionally this is
solved by domain-specific algorithms that decide what can be
shared. With the layered schema architecture, data elements are
classified using overlays based on policies and privacy settings, and
are filtered or redacted using semantic pipelines.

{{<figure src="data-exchange.png" class="text-center my-3">}} 

### Use Case: Data Display and Entry

Layered schema architecture allows building data entry/display
applications that harmonize semantics to implement variations for
language, locale, formatting, and jurisdiction. An application can use
locale-specific overlays to generate views and entry forms.

{{<figure src="layered-schema-data-capture-application.png" class="text-center my-3">}} 


