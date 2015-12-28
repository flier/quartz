package quartz

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type Cloneable interface {
	Clone() interface{}
}

type Sortable interface {
	sort.Interface
}

type CompareFunc func(lhs, rhs interface{}) int

type StringKeys []interface{}

func (keys StringKeys) Len() int { return len(keys) }

func (keys StringKeys) Less(i, j int) bool {
	return strings.Compare(keys[i].(string), keys[j].(string)) < 0
}

func (keys StringKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

type MapEntry interface {
	Key() string

	Value() interface{}
}

type Map interface {
	Cloneable

	Empty() bool

	Len() int

	Keys() []string

	Values() []interface{}

	Entries() []MapEntry

	Contains(key string) bool

	Get(key string) interface{}

	Put(key string, value interface{})

	PutAll(m Map)

	Remove(key string) interface{}
}

type DirtyFlagMap interface {
	Map

	Dirty() bool

	ClearDirtyFlag()
}

type Set interface {
	Empty() bool

	Len() int

	Keys() []interface{}

	Contains(item interface{}) bool

	Add(item interface{})

	Remove(item interface{}) bool
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

func NewDirtyFlagMap() DirtyFlagMap {
	return &dirtyFlagMap{entries: make(map[string]interface{})}
}

func (m *dirtyFlagMap) Dirty() bool { return m.dirty }

func (m *dirtyFlagMap) ClearDirtyFlag() { m.dirty = false }

func (m *dirtyFlagMap) Empty() bool { return len(m.entries) == 0 }

func (m *dirtyFlagMap) Len() int { return len(m.entries) }

func (m *dirtyFlagMap) Keys() (keys []string) {
	for key, _ := range m.entries {
		keys = append(keys, key)
	}

	if keys != nil {
		sort.Sort(sort.StringSlice(keys))
	}

	return
}

func (m *dirtyFlagMap) Values() (values []interface{}) {
	for _, key := range m.Keys() {
		values = append(values, m.entries[key])
	}

	return
}

func (m *dirtyFlagMap) Entries() (entries []MapEntry) {
	for _, key := range m.Keys() {
		entries = append(entries, &mapEntry{key, m.entries[key]})
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

func (m *dirtyFlagMap) Remove(key string) interface{} {
	value, exists := m.entries[key]

	delete(m.entries, key)

	if exists {
		m.dirty = true
	}

	return value
}

func (m *dirtyFlagMap) Clone() interface{} {
	clone := dirtyFlagMap{
		entries: make(map[string]interface{}),
		dirty:   m.dirty,
	}

	for _, entry := range m.Entries() {
		clone.Put(entry.Key(), entry.Value())
	}

	return &clone
}

type hashSet map[interface{}]bool

func NewHashSet() Set { return make(hashSet) }

func (s hashSet) Empty() bool { return len(s) == 0 }

func (s hashSet) Len() int { return len(s) }

func (s hashSet) Keys() (keys []interface{}) {
	for key, _ := range s {
		keys = append(keys, key)
	}

	return
}

func (s hashSet) Contains(key interface{}) bool {
	_, exists := s[key]

	return exists
}

func (s hashSet) Add(key interface{}) {
	s[key] = true
}

func (s hashSet) Remove(key interface{}) bool {
	_, exists := s[key]

	delete(s, key)

	return exists
}

type treeSet struct {
	items   []interface{}
	compare CompareFunc
}

func NewTreeSet(compare CompareFunc) Set {
	return &treeSet{
		compare: compare,
	}
}

func (s *treeSet) Empty() bool { return len(s.items) == 0 }

func (s *treeSet) Len() int { return len(s.items) }

func (s *treeSet) Less(i, j int) bool {
	return s.compare(s.items[i], s.items[j]) <= 0
}

func (s *treeSet) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s *treeSet) Keys() (keys []interface{}) {
	return s.items
}

func (s *treeSet) Contains(item interface{}) bool {
	n := sort.Search(len(s.items), func(i int) bool {
		return s.compare(item, s.items[i]) <= 0
	})

	return n < len(s.items) && s.compare(s.items[n], item) == 0
}

func (s *treeSet) Add(item interface{}) {
	n := sort.Search(len(s.items), func(i int) bool {
		return s.compare(item, s.items[i]) <= 0
	})

	fmt.Printf("Add %s @ %d, %v", item, n, s.items)

	if n == len(s.items) {
		s.items = append(s.items, item)
	} else if s.compare(item, s.items[n]) != 0 {
		items := make([]interface{}, len(s.items)+1)

		copy(items[:n], s.items[:n])
		items[n] = item
		copy(items[n+1:], s.items[n:])

		s.items = items
	}
}

func (s *treeSet) Remove(item interface{}) bool {
	n := sort.Search(len(s.items), func(i int) bool {
		return s.compare(item, s.items[i]) <= 0
	})

	if n < len(s.items) && s.compare(s.items[n], item) == 0 {
		s.items = append(s.items[:n], s.items[n+1:]...)

		return true
	}

	return false
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
