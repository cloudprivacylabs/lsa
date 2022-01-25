# Reference Fields

A schema may specify fields that are references to entities. These
references can specify an inclusion of an entity, or a link to an
entity. 

This is a reference that defines an attribute of a schema as another
entity:

```
"attr": {
  "@type": "Reference",
  "reference": "otherEntity",
  "attributeName": "name"
}
```

This is a reference that defines a link to another entity:

```
"attr": {
  "@type": "Reference",
  "reference": "otherEntity",
  "fields": [ "$otherEntityId" ],
  "targetFields" [ "$id" ],
  "direction": "->",
  "label": "https://lschema.org/has"
}
```

This definition will look for an entity of type `otherEntity` in the
ingested data, whose `id` field is equal to the value of
`otherEntityId` field of the source entity. If found, a link with
label `label` added from the source entity to the target `otherEntity`
because the direction is `->`.
