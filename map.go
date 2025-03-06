package problauncher

import (
	"container/list"
)

type entry[K comparable, V any] struct {
	key   K
	value V
}

type FixedSizeMap[K comparable, V any] struct {
	capacity int
	cache    map[K]*list.Element
	ll       *list.List
}

func NewFixedSizeMap[K comparable, V any](capacity int) *FixedSizeMap[K, V] {
	if capacity <= 0 {
		panic("capacity must be a positive integer")
	}
	return &FixedSizeMap[K, V]{
		capacity: capacity,
		cache:    make(map[K]*list.Element),
		ll:       list.New(),
	}
}

func (m *FixedSizeMap[K, V]) Set(key K, value V) {
	if ele, ok := m.cache[key]; ok {
		m.ll.MoveToFront(ele)
		ele.Value.(*entry[K, V]).value = value
		return
	}

	ele := m.ll.PushFront(&entry[K, V]{key, value})
	m.cache[key] = ele

	if len(m.cache) > m.capacity {
		m.removeOldest()
	}
}

func (m *FixedSizeMap[K, V]) Get(key K) (V, bool) {
	if ele, ok := m.cache[key]; ok {
		m.ll.MoveToFront(ele)
		return ele.Value.(*entry[K, V]).value, true
	}
	var zero V
	return zero, false
}

func (m *FixedSizeMap[K, V]) Remove(key K) bool {
	if ele, ok := m.cache[key]; ok {
		m.ll.Remove(ele)
		delete(m.cache, key)
		return true
	}
	return false
}

func (m *FixedSizeMap[K, V]) removeOldest() {
	ele := m.ll.Back()
	if ele != nil {
		m.ll.Remove(ele)
		kv := ele.Value.(*entry[K, V])
		delete(m.cache, kv.key)
	}
}

func (m *FixedSizeMap[K, V]) Len() int {
	return len(m.cache)
}
