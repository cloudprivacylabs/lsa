%.png: %.json
	layers dot $< | dot -Tpng -o $@

all: content/Object/object_node_model.png \
     content/Object/attributes_model.png \
     content/Object/attributelist_model.png \
     content/Value/value_node_model.png \
	   content/Array/array_node_model.png

