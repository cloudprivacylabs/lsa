# ls:entityId - Entity unique identifier

```
https://lschema.org/entityId
```

If a node has `ls:entityId` property, that node is a unique identifier
field for that entity. `ls:entityId` is a marker term. Its contents
are ignored. 

An entity id node must be accessible from the root of the entity using
only attribute nodes. That is, an entity id cannot be under an array.
