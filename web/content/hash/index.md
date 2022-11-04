---
title: "hash"
type: "term"
---

# Hashes

{{% termheader term="https://lschema.org/hash" key="hash" type="Node property" use="Schema nodes" %}}
Creates a hash value from the given fields.
{{% /termheader %}}

<code>hash</code> is used to set the value of an attribute as the hash
value of a group of attributes. If the term <code>hash</code> is used,
a sha256 hash is computed. You can also specify
<code>hash.sha1</code>, <code>hash.sha256</code> and
<code>hash.sha512</code> as a hash function. The contents of the field
can be a string, or a string array containing attribute ids.

```
{
   "@id": "testAttribute",
   "@type": "Value",
   "hash": [
      "attr1",
      "attr2"
    ]
}
```

The above declaration will set the value of `testAttribute` from the
sha256 hash of the values of `attr1` and `attr2`.

