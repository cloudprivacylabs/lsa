%.png: %.json
	layers dot $< | dot -Tpng -o $@
%.svg: %.json
	layers dot $< | dot -Tsvg -o $@

content/schemamodel.png: content/schemamodel.json
	layers dot $< --rankdir TB | dot -Tpng -o $@

all: content/schemamodel.png \
   content/Object/object_node_model.png \
	 content/Object/object_as_node_schema.png \
	 content/Object/object_as_node.png \
	 content/Object/object_as_edge_schema.png \
	 content/Object/object_as_edge.png \
     content/Object/attributes_model.png \
     content/Object/attributelist_model.png \
     content/Value/value_node_model.png \
	   content/Array/array_node_model.png \
	   content/Array/array_as_node.png \
	   content/Array/array_as_node_schema.png \
	   content/Array/array_as_edge_schema.png \
	   content/Array/array_as_edge.png \
	   content/Value/value_as_node.png \
	   content/Value/value_as_edge.png \
	   content/Value/value_as_node_schema.png  \
	   content/Value/value_as_edge_schema.png 

