package quartz

import (
	"sort"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDirtyFlagMap(t *testing.T) {
	Convey("Given a DirtyFlagMap", t, func() {
		m := NewDirtyFlagMap()

		So(m, ShouldNotBeNil)

		So(m.Dirty(), ShouldBeFalse)
		So(m.Empty(), ShouldBeTrue)
		So(m.Len(), ShouldEqual, 0)
		So(m.Contains("key"), ShouldBeFalse)
		So(m.Keys(), ShouldBeNil)

		Convey("Put a item", func() {
			m.Put("key", "value")

			So(m.Dirty(), ShouldBeTrue)
			So(m.Empty(), ShouldBeFalse)
			So(m.Len(), ShouldEqual, 1)
			So(m.Contains("key"), ShouldBeTrue)
			So(m.Get("key"), ShouldEqual, "value")
			So(m.Keys(), ShouldResemble, []string{"key"})

			Convey("Clear the dirty flag", func() {
				m.ClearDirtyFlag()

				So(m.Dirty(), ShouldBeFalse)
			})

			Convey("Remove the item", func() {
				So(m.Remove("key"), ShouldEqual, "value")
				So(m.Dirty(), ShouldBeTrue)
				So(m.Empty(), ShouldBeTrue)

				m.ClearDirtyFlag()

				So(m.Remove("key"), ShouldEqual, nil)
				So(m.Dirty(), ShouldBeFalse)
			})
		})

		Convey("Given another map", func() {
			other := &dirtyFlagMap{entries: map[string]interface{}{
				"key": "value",
				"foo": 0,
				"bar": 1,
			}}
			Convey("Put all items", func() {
				m.PutAll(other)

				So(m.Dirty(), ShouldBeTrue)
				So(m.Empty(), ShouldBeFalse)
				So(m.Len(), ShouldEqual, 3)
				So(m.Contains("key"), ShouldBeTrue)
				So(m.Get("key"), ShouldEqual, "value")
				So(m.Keys(), ShouldResemble, []string{"bar", "foo", "key"})
				So(m.Values(), ShouldResemble, []interface{}{1, 0, "value"})

				entities := m.Entries()

				So(entities, ShouldNotBeNil)
				So(len(entities), ShouldEqual, 3)
				So(entities[0].Key(), ShouldEqual, "bar")
				So(entities[0].Value(), ShouldEqual, 1)
			})

			Convey("Clone a map", func() {
				m := other.Clone().(*dirtyFlagMap)

				So(m.Dirty(), ShouldBeTrue)
				So(m.Empty(), ShouldBeFalse)
				So(m.Len(), ShouldEqual, 3)
				So(m.Contains("key"), ShouldBeTrue)
				So(m.Get("key"), ShouldEqual, "value")
				So(m.Keys(), ShouldResemble, []string{"bar", "foo", "key"})

				So(m.entries, ShouldNotEqual, other.entries)
				So(m.entries, ShouldResemble, other.entries)
			})
		})
	})
}

func TestHashSet(t *testing.T) {
	Convey("Given a HashSet", t, func() {
		s := NewHashSet()

		So(s, ShouldNotBeNil)
		So(s.Empty(), ShouldBeTrue)
		So(s.Len(), ShouldEqual, 0)

		Convey("Give some keys", func() {
			s.Add("key")
			s.Add("foo")
			s.Add("bar")
			s.Add("key")
			s.Add("bar")

			So(s.Empty(), ShouldBeFalse)
			So(s.Len(), ShouldEqual, 3)
			So(s.Contains("key"), ShouldBeTrue)
			So(s.Contains("foo"), ShouldBeTrue)
			So(s.Contains("bar"), ShouldBeTrue)
			So(s.Contains("any"), ShouldBeFalse)

			keys := s.Keys()

			sort.Sort(StringKeys(keys))

			So(keys, ShouldResemble, []interface{}{"bar", "foo", "key"})

			Convey("Remove a key", func() {
				So(s.Remove("key"), ShouldBeTrue)

				So(s.Len(), ShouldEqual, 2)
				So(s.Contains("key"), ShouldBeFalse)
				So(s.Remove("key"), ShouldBeFalse)

				So(s.Remove("foo"), ShouldBeTrue)

				So(s.Len(), ShouldEqual, 1)
				So(s.Contains("foo"), ShouldBeFalse)
				So(s.Remove("foo"), ShouldBeFalse)

				So(s.Remove("bar"), ShouldBeTrue)

				So(s.Len(), ShouldEqual, 0)
				So(s.Contains("bar"), ShouldBeFalse)
				So(s.Remove("bar"), ShouldBeFalse)
			})
		})
	})
}

func TestTreeSet(t *testing.T) {
	Convey("Given a TreeSet", t, func() {
		s := NewTreeSet(func(lhs, rhs interface{}) int {
			return strings.Compare(lhs.(string), rhs.(string))
		})

		So(s, ShouldNotBeNil)
		So(s.Empty(), ShouldBeTrue)
		So(s.Len(), ShouldEqual, 0)

		s.Add("key")
		So(s.Keys(), ShouldResemble, []interface{}{"key"})
		s.Add("foo")
		So(s.Keys(), ShouldResemble, []interface{}{"foo", "key"})
		s.Add("bar")
		So(s.Keys(), ShouldResemble, []interface{}{"bar", "foo", "key"})
		s.Add("key")
		So(s.Keys(), ShouldResemble, []interface{}{"bar", "foo", "key"})
		s.Add("bar")
		So(s.Keys(), ShouldResemble, []interface{}{"bar", "foo", "key"})

		So(s.Empty(), ShouldBeFalse)
		So(s.Len(), ShouldEqual, 3)
		So(s.Contains("key"), ShouldBeTrue)
		So(s.Contains("foo"), ShouldBeTrue)
		So(s.Contains("bar"), ShouldBeTrue)
		So(s.Contains("any"), ShouldBeFalse)

		Convey("Remove a key", func() {
			So(s.Remove("key"), ShouldBeTrue)

			So(s.Len(), ShouldEqual, 2)
			So(s.Contains("key"), ShouldBeFalse)
			So(s.Remove("key"), ShouldBeFalse)

			So(s.Remove("foo"), ShouldBeTrue)

			So(s.Len(), ShouldEqual, 1)
			So(s.Contains("foo"), ShouldBeFalse)
			So(s.Remove("foo"), ShouldBeFalse)

			So(s.Remove("bar"), ShouldBeTrue)

			So(s.Len(), ShouldEqual, 0)
			So(s.Contains("bar"), ShouldBeFalse)
			So(s.Remove("bar"), ShouldBeFalse)
		})
	})
}

func TestUniqueName(t *testing.T) {
	Convey("Given a unique name", t, func() {
		name := newUniqueName("test")

		So(len(name), ShouldEqual, 41)
		So(name, ShouldStartWith, "2627b4f6-")
	})
}
