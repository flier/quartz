package quartz

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
)

type MapEntry interface {
	Key() string

	Value() interface{}
}

type Map interface {
	Empty() bool

	Size() int

	Keys() []string

	Values() []interface{}

	Entries() []MapEntry

	Contains(key string) bool

	Get(key string) interface{}

	Put(key string, value interface{})

	PutAll(m Map)
}

type DirtyFlagMap interface {
	Map

	Dirty() bool

	ClearDirtyFlag()
}

type mapEntry struct {
	key   string
	value interface{}
}

func (e *mapEntry) Key() string { return e.key }

func (e *mapEntry) Value() interface{} { return e.value }

type dirtyFlagMap struct {
	entries map[string]interface{}
	dirty   bool
}

func (m *dirtyFlagMap) Dirty() bool { return m.dirty }

func (m *dirtyFlagMap) ClearDirtyFlag() { m.dirty = false }

func (m *dirtyFlagMap) Empty() bool { return len(m.entries) == 0 }

func (m *dirtyFlagMap) Size() int { return len(m.entries) }

func (m *dirtyFlagMap) Keys() (keys []string) {
	for key, _ := range m.entries {
		keys = append(keys, key)
	}

	sort.Sort(sort.StringSlice(keys))

	return
}

func (m *dirtyFlagMap) Values() (values []interface{}) {
	for _, value := range m.entries {
		values = append(values, value)
	}

	return
}

func (m *dirtyFlagMap) Entries() (entries []MapEntry) {
	for key, value := range m.entries {
		entries = append(entries, &mapEntry{key, value})
	}

	return
}

func (m *dirtyFlagMap) Contains(key string) bool {
	_, exists := m.entries[key]

	return exists
}

func (m *dirtyFlagMap) Get(key string) interface{} { return m.entries[key] }

func (m *dirtyFlagMap) Put(key string, value interface{}) {
	if v, exists := m.entries[key]; !exists || v != value {
		m.entries[key] = value
		m.dirty = true
	}
}

func (m *dirtyFlagMap) PutAll(o Map) {
	for _, entry := range o.Entries() {
		m.Put(entry.Key(), entry.Value())
	}
}

type HashSet map[string]bool

func NewHashSet() HashSet { return make(HashSet) }

func (s HashSet) Contains(key string) bool {
	_, exists := s[key]

	return exists
}

func (s HashSet) Add(key string) {
	s[key] = true
}

type TreeSet struct {
	items   []interface{}
	compare func(lhs, rhs interface{}) bool
}

func (s TreeSet) Len() int { return len(s.items) }

func (s TreeSet) Less(i, j int) bool {
	return s.compare(s.items[i], s.items[j])
}

func (s TreeSet) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s TreeSet) Contains(item interface{}) bool {
	n := sort.Search(len(s.items), func(i int) bool {
		return s.compare(item, s.items[i])
	})

	return n < len(s.items) && s.items[n] == item
}

func (s TreeSet) Add(item interface{}) {
	n := sort.Search(len(s.items), func(i int) bool {
		return s.compare(item, s.items[i])
	})

	s.items = append(append(s.items[:n], item), s.items[n:]...)
}

const (
	DEFAULT_GROUP = "DEFAULT"
)

func newUniqueName(group string) string {
	buf := make([]byte, 16)

	rand.Read(buf)

	hash := md5.Sum([]byte(group))

	return fmt.Sprintf("%s-%s", hex.EncodeToString(hash[12:]), hex.EncodeToString(buf))
}
