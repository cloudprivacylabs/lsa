.DEFAULT_GOAL:= all

%.lr.png: %.lr.json
	layers dot $< --rankdir LR | dot -Tpng -o $@
%.png: %.json
	layers dot $< --rankdir TB | dot -Tpng -o $@
%.svg: %.json
	layers dot $< --rankdir TB | dot -Tsvg -o $@



all: content/docs/model_and_syntax/valueschema.png \
	content/docs/model_and_syntax/objectschema1.png \
	content/docs/model_and_syntax/objectschema2.png \
	content/docs/model_and_syntax/arrayschema.png \
	content/docs/model_and_syntax/referenceschema.png \
	content/docs/model_and_syntax/addressschema.png \
  content/docs/model_and_syntax/compiled_address_schema.png \
  content/docs/model_and_syntax/compositeschema.png \
  content/docs/model_and_syntax/polymorphicschema.png \
   content/docs/model_and_syntax/schemaroot.png \
  content/docs/basics/example-lpg1.png \
  content/docs/basics/example-lpg2.png \
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

