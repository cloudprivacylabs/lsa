[{"type":"http://layeredschemas.org/Schema","id":"http://example.org/Contact/schema","targetType":["http://example.org/Contact"],"payload":[{"@id":"http://example.org/Contact/schema","@type":["http://layeredschemas.org/Schema"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Contact/type","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"type"}]},{"@id":"http://example.org/Contact/value","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"value"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Contact"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Contact/ovl1","targetType":["http://example.org/Contact"],"payload":[{"@id":"http://example.org/Contact/ovl1","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Contact/value","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org/PhoneNumber"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Contact"}]}]},{"type":"http://layeredschemas.org/SchemaManifest","id":"http://example.org/Contact/schemaManifest","targetType":["http://example.org/Contact"],"payload":[{"@id":"http://example.org/Contact/schemaManifest","@type":["http://layeredschemas.org/SchemaManifest"],"http://layeredschemas.org/SchemaManifest/overlays":[{"@list":[{"@id":"http://example.org/Contact/ovl1"}]}],"http://layeredschemas.org/SchemaManifest/schema":[{"@id":"http://example.org/Contact/schema"}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Contact"}]}]},{"type":"http://layeredschemas.org/SchemaManifest","id":"http://example.org/Person2/schema","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person2/schema","@type":["http://layeredschemas.org/SchemaManifest"],"http://layeredschemas.org/SchemaManifest/overlays":[{"@list":[{"@id":"http://example.org/Person/ovl2"}]}],"http://layeredschemas.org/SchemaManifest/schema":[{"@id":"http://example.org/Person/schemaBase"}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/Schema","id":"http://example.org/Person/schemaBase","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/schemaBase","@type":["http://layeredschemas.org/Schema"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"firstName"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"lastName"}]},{"@id":"http://example.org/Person/contact","@type":["http://layeredschemas.org/Array"],"http://layeredschemas.org/Array/items":[{"@type":["http://layeredschemas.org/Reference"],"http://layeredschemas.org/Reference/reference":[{"@id":"http://example.org/Contact"}]}],"http://layeredschemas.org/attr/name":[{"@value":"contact"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Person/bit","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/bit","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#BIT"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#BIT"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/SchemaManifest","id":"http://example.org/Person_bit/schema","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person_bit/schema","@type":["http://layeredschemas.org/SchemaManifest"],"http://layeredschemas.org/SchemaManifest/overlays":[{"@list":[{"@id":"http://example.org/Person/ovl1"},{"@id":"http://example.org/Person/bit"}]}],"http://layeredschemas.org/SchemaManifest/schema":[{"@id":"http://example.org/Person/schemaBase"}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Person/dpv","targetType":["http://example.org/Person","http://www.w3.org/ns/dpv#DataSubject"],"payload":[{"@id":"http://example.org/Person/dpv","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://www.w3.org/ns/dpv#hasPersonalDataCategory":[{"@id":"http://www.w3.org/ns/dpv#Name"},{"@id":"http://www.w3.org/ns/dpv#Identifying"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://www.w3.org/ns/dpv#hasPersonalDataCategory":[{"@id":"http://www.w3.org/ns/dpv#Name"},{"@id":"http://www.w3.org/ns/dpv#Identifying"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"},{"@id":"http://www.w3.org/ns/dpv#DataSubject"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Person/ovl1","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/ovl1","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Person/ovl1","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/ovl1","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/Overlay","id":"http://example.org/Person/ovl2","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/ovl2","@type":["http://layeredschemas.org/Overlay"],"http://layeredschemas.org/Object/attributes":[{"@id":"http://example.org/Person/firstName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"first"}],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]},{"@id":"http://example.org/Person/lastName","@type":["http://layeredschemas.org/Value"],"http://layeredschemas.org/attr/name":[{"@value":"last"}],"http://layeredschemas.org/attr/privacyClassification":[{"@value":"http://example.org#PII"}]}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/SchemaManifest","id":"http://example.org/Person/schema","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person/schema","@type":["http://layeredschemas.org/SchemaManifest"],"http://layeredschemas.org/SchemaManifest/overlays":[{"@list":[{"@id":"http://example.org/Person/ovl1"}]}],"http://layeredschemas.org/SchemaManifest/schema":[{"@id":"http://example.org/Person/schemaBase"}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]},{"type":"http://layeredschemas.org/SchemaManifest","id":"http://example.org/Person_dpv/schema","targetType":["http://example.org/Person"],"payload":[{"@id":"http://example.org/Person_dpv/schema","@type":["http://layeredschemas.org/SchemaManifest"],"http://layeredschemas.org/SchemaManifest/overlays":[{"@list":[{"@id":"http://example.org/Person/dpv"}]}],"http://layeredschemas.org/SchemaManifest/schema":[{"@id":"http://example.org/Person/schemaBase"}],"http://layeredschemas.org/targetType":[{"@id":"http://example.org/Person"}]}]}]
