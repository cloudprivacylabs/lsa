# Layered Schemas

Layered schema architecture is developed work with unified data
semantics regardless of the data capture locale, data representation,
format, or terminology. Traditional schemas (such as JSON/XML schemas)
describe the structure of data with limited semantic
information. Layered schema architecture slices a traditional schema
into multiple layers with each layer adding structural information and
semantic annotations. The schema base is the first layer that defines
the structure and the vocabulary used to capture and process data. The
data model can be flat for tabular, linked with other schemas, or
nested to represent more elaborate structures such as health
data. Schema layers add semantic annotations to the base model, or
modify structure of the underlying object defined in the schema
base. These annotations include:

  * Constraints: Required attributes, length limits,...
  * Format: Phone number with area code, date/time,...
  * Language: English, Spanish,...
  * Privacy classifications: Personally identifiable information, sensitive information,...
  * Retention policies: Attributes must be cleared after a period,...
  * Provenance: Data source, signatures, ...

Use-cases for layered schemas include:

  * **Data entry:** A layered schema can include overlays that have
    information to auto-generate data entry forms. These can be
    locale-specific overlays including valid options, entry format, UI
    labels and help text, etc.
  * **Data ingestion:** Semantics for data coming from multiple
    sources (APIs, file uploads, data entrty) can be defined using a
    unified vocabulary at the schema bases, and source-specific
    variations can be implemented as different sets of overlays.
  * **Granular privacy/security controls** Different privacy/security
    classifications can be applied to data elements based on the
    overlays used. For instance, data elements with a certain privacy
    classification can be masked in one context, and left untouched in
    another context. 
  * **De-identification** Attributes that might reveal the identity of
    data subjects can be removed from the data. Using different
    overlays, the set of attributes that will be removed from the data
    can be controlled.

## Schemas

A layered schema is a JSON-LD document. It describes the structure of
an object and semantic information for its data elements. Most of the
information/constraints encoded in JSON/XML schemas or CSV files can
be represented using layered schemas.

Consider this JSON schema:

```
{
  "type": "object",
  "properties": {
     "link": {
        "type": "string",
        "format": "uri"
     },
     "num": {
        "type" "number"
    }
  }
}
```

An equivalent layered schema is as follows:

```
{
  "@context": "http://schemas.cloudprivacylabs.com/layers.jsonld",
  "@type": "SchemaBase",
  "attributes": {
    "id1": {
       "attributeName": "link",
       "type": "string",
       "format": "uri"
    },
    "id2": {
       "attributeName": "num",
       "type": "number"
    }
  }
}
```
