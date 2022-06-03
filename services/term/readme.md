# Terminology Service Bridge

This small Python project is a web service that connects to a
terminology service to perform valueset and terminology lookups.  The
service responds to valueset lookup requests from the `layers`
program. When processing a valueset lookup, `layers` calls the service
with the name of the valueset table, and key-value pairs taken from
the document. The service uses the `queries.yaml` file to construct
queries, and returns the first found result. `layers` inserts these
values into the graph.

## Building and Configuration 

To build a Docker image, use

```
docker build  -t termservice:latest . 
```

This will package the service as a docker image.

To configure the service, edit the configuration files under
`cfg`. `database.ini` file contains database specific connection
information. `queries.yaml` contains the database queries.

```
valuesets:
  - tableId: measurement
    queries:
      - query: "select concept_id,concept_name,domain_id,vocabulary_id,concept_class_id,concept_code from vocabulary.concept where vocabulary_id=%(vocabulary)s and concept_code=%(code)s;"
        columns: 
              - code
              - vocabulary
      - query: "select concept_id,concept_name,domain_id,vocabulary_id,concept_class_id,concept_code from vocabulary.concept where vocabulary_id=%(vocabulary)s and concept_name=%(code)s;"
        columns: 
              - code
              - vocabulary
```

`valuesets` object lists all the different valuesets. The `tableId`
gives the valueset ID for each valueset. Layered schemas refer to
valuesets using this `tableId`. This example is a valueset for
`measurement`s.

`queries` is an array that lists queries to run. For each lookup, the
service runs these queries one by one until one of them returns a
nonempty resultset. This allows testing different lookup strategies
for data normalization. This example looks up the `code` in
`concept_code` column, and if that fails, it looks it up in
`concept_name` column. This is useful if the input data contains codes
or names of concepts in a single column.

For each query, the `query` object specifies the query and the column
that will be bound to that query. For each query, `columns` specifies
the request keys that will be bound to the query. 

For example, consider the following data input:

```
CODE,VOCAB
29463-7,LOINC
```

The layered schema can be defined to request `code=29463-7` and
`vocabulary=LOINC`. This will create a request of the form:

```
GET https://localhost:8011?tableId=measurement&code=29463-7&vocabulary=LOINC
```

Using this configuration, the database query becomes:

```
select concept_id,concept_name,domain_id,vocabulary_id,concept_class_id,concept_code from vocabulary.concept where vocabulary_id='LOINC' and concept_code='29463-7'
```

This results in:

```
{3025315, Body Weight, Measurement, LOINC, Clinical Observation, 29463-7}
```
which is returned as:
```
{
  "concept_id": "3025315",
  "concept_name": "Body Weight",
  "domain_id": "Measurement",
  "vocabulary_id": "LOINC",
  "concept_class_id": "Clinical Observation",
  "concept_code": "29463-7"
}
```

A valueset configuration file is used for `layers`:
```
valuesets.json:
{
    "services": {
        "measurement": "http://localhost:8011"
    },
   ...
}
```
This allows `layers` to find the server when a lookup for `measurement` is requested.
