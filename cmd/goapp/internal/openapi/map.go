package openapi

import "iter"

type Map[T any, U comparable] struct {
	data []T
	hash map[U]int
}

func (m *Map[T, U]) Push(data T, key U) {
	if len(m.data) == 0 {
		m.hash = make(map[U]int)
	}
	m.hash[key] = len(m.data)
	m.data = append(m.data, data)
}

func (m *Map[T, U]) Get(key U) *T {
	return &m.data[m.hash[key]]
}
func (m *Map[T, U]) Index(key int) *T {
	return &m.data[key]
}

func (m *Map[T, U]) Length() int {
	return len(m.data)
}

func (m *Map[T, U]) Exists(key U) bool {
	_, ok := m.hash[key]
	return ok
}

func (m *Map[T, U]) Items() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range m.data {
			if !yield(v) {
				return
			}
		}
	}
}
