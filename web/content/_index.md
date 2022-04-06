---
title: "Layered Schema Architecture"
subtitle: "Semantic interoperability for data capture, processing, and exchange"
description: "Layered schemas data capture, semantic interoperability, knowledge graphs"
features:
  - title: Semantic Harmonization
    text: Ingest data from disparate sources. Harmonize, process and transform ingested data using semantic pipelines.
  - title: Cross-Domain Interoperability
    text: Annotate data with domain-specific vocabulary mappings, convert and exchange for specific targets
  - title: From Data to Knowledge
    text: Build knowledge graphs with contextual metadata. Create AI training sets from heterogeneous data.
#button:
#  link: https://playground.layeredschemas.org
#  text: Launch the Layered Schema Playground
---


Layered Schema Architecture (LSA) enables semantic interoperability
between heterogeneous systems. LSA uses a schema base (such as FHIR
schemas for health data) to define data structures, and use-case
specific interchangeable overlays to add semantics, contextual
metadata, and processing directives to build a schema
variant. Different schema variants can be used to ingest data from or
export data to disparate systems. Each variant encodes source specific
metadata and rules to ingest data into a knowledge graph, or target
specific metadata and rules to translate the knowledge graph into
another format.

## Use Cases
### Interoperability Across Domains

Schemas and vocabularies are usually domain specific.  Achieving
interoperability across domains requires manual mappings, which is
further complicated by variations due to conventions and jurisdictions
The layered schema architecture allows ingesting data from several
sources that may have variations due to vendor specific extensions or
jurisdiction. Ingested data can be annotated and translated into data
usable for other domains.


{{<figure src="vocab-mapping.png" class="text-center my-3">}} 

### Semantic Data Warehouse

A traditional data warehouse uses source specific ETL scripts to
normalize and ingest data. Maintenance of such scripts relies heavily
on internally developed tools and know-how.  A semantic data warehouse
based on the layered schema architecture replaces such ETL scripts
with source specific schema variants. The schema variants can be
reused in different scenarios and can be shared.

{{<figure src="dw-fanin-fanout-sm.png" class="text-center my-3">}} 

### Privacy-Conscious Data Exchange

Data exchange policies and user consent dictate what types of data can
be exchanged with whom and for what purposes. Traditionally this is
solved by domain-specific algorithms that decide what can be
shared. With the layered schema architecture, data elements are
classified using overlays based on policies and privacy settings, and
are filtered or redacted using semantic pipelines.

{{<figure src="data-exchange.png" class="text-center my-3">}} 

### Data Display and Entry

Layered schema architecture allows building data entry/display
applications that harmonize semantics to implement variations for
language, locale, formatting, and jurisdiction. An application can use
locale-specific overlays to generate views and entry forms.

{{<figure src="layered-schema-data-capture-application.png" class="text-center my-3">}} 


