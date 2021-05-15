package memgraph

import (
	"container/list"
)

// An indexes list is an iteratable map
type indexedList struct {
	index map[string]*list.Element
	list  list.List
}

func newIndexedList() indexedList {
	return indexedList{index: make(map[string]*list.Element)}
}

func (l *indexedList) add(key string, value interface{}) *list.Element {
	el := l.list.PushBack(value)
	l.index[key] = el
	return el
}

func (l *indexedList) remove(key string, el *list.Element) {
	l.list.Remove(el)
	delete(l.index, key)
}
