package quartz

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJobKey(t *testing.T) {
	Convey("Given a JobKey", t, func() {
		key := NewJobKey("name")

		So(key.Name(), ShouldEqual, "name")
		So(key.Group(), ShouldEqual, DEFAULT_GROUP)
		So(key.String(), ShouldEqual, "DEFAULT.name")
	})

	Convey("Given a grouped JobKey", t, func() {
		key := NewGroupJobKey("name", "group")

		So(key.Name(), ShouldEqual, "name")
		So(key.Group(), ShouldEqual, "group")
		So(key.String(), ShouldEqual, "group.name")
	})

	Convey("Given a uniqued JobKey", t, func(c C) {
		key := NewUniqueKey("")

		c.Printf("unique key: %s", key)

		So(len(key.Name()), ShouldEqual, 41)
		So(key.Group(), ShouldEqual, DEFAULT_GROUP)
	})
}

func TestJobDetail(t *testing.T) {
	Convey("Given a JobDataMap", t, func() {
		m := NewJobDataMap()

		So(m, ShouldNotBeNil)
		So(m.Empty(), ShouldBeTrue)
		So(m.Len(), ShouldEqual, 0)
		So(m.Dirty(), ShouldBeFalse)

		Convey("Put a item", func() {
			m.Put("key", "value")

			So(m.Empty(), ShouldBeFalse)
			So(m.Len(), ShouldEqual, 1)
			So(m.Dirty(), ShouldBeTrue)
			So(m.Contains("key"), ShouldBeTrue)
			So(m.Get("key"), ShouldEqual, "value")
			So(m.Keys(), ShouldResemble, []string{"key"})
			So(m.Values(), ShouldResemble, []interface{}{"value"})

			entries := m.Entries()

			So(len(entries), ShouldEqual, 1)
			So(entries[0].Key(), ShouldEqual, "key")
			So(entries[0].Value(), ShouldEqual, "value")

			m.ClearDirtyFlag()

			So(m.Dirty(), ShouldBeFalse)
		})

		Convey("Put a DataMap", func() {
			d := NewJobDataMap()
			d.Put("foo", "bar")
			d.Put("abc", "def")
			d.Put("key", "another")

			m.PutAll(d)

			So(m.Empty(), ShouldBeFalse)
			So(m.Len(), ShouldEqual, 3)
			So(m.Dirty(), ShouldBeTrue)
			So(m.Contains("key"), ShouldBeTrue)
			So(m.Get("key"), ShouldEqual, "another")
			So(m.Keys(), ShouldResemble, []string{"abc", "foo", "key"})
			So(len(m.Entries()), ShouldEqual, 3)
		})
	})
}

func TestJobBuilder(t *testing.T) {
	Convey("Given a JobBuilder, then build JobDetail", t, func() {
		b := &JobBuilder{}

		Convey("JobBuilder -> JobDetail.JobBuilder()", func() {
			So(b.Build().JobBuilder(), ShouldEqual, b)
		})

		Convey("WithIdentity -> JobDetail.Key()", func() {
			b.WithIdentity("name")

			So(b.Build().Key().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("WithGroupIdentity -> JobDetail.Key()", func() {
			b.WithGroupIdentity("name", "group")

			So(b.Build().Key().String(), ShouldEqual, "group.name")
		})

		Convey("WithJobKey -> JobDetail.Key()", func() {
			b.WithJobKey(NewJobKey("name"))

			So(b.Build().Key().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("WithDescription -> JobDetail.Description()", func() {
			b.WithDescription("desc")

			So(b.Build().Description(), ShouldEqual, "desc")
		})

		Convey("UsingJobData -> JobDetail.JobDataMap()", func() {
			b.UsingJobData("key", "value")

			So(b.Build().JobDataMap().Get("key"), ShouldEqual, "value")
		})

		Convey("UsingJobDataMap -> JobDetail.JobDataMap()", func() {
			m := NewJobDataMap()
			m.Put("key", "value")

			b.UsingJobDataMap(m)

			So(b.Build().JobDataMap().Get("key"), ShouldEqual, "value")
		})

		Convey("SetJobDataMap -> JobDetail.JobDataMap()", func() {
			b.UsingJobData("nonexists", "value")

			m := NewJobDataMap()
			m.Put("key", "value")
			b.SetJobDataMap(m)

			dm := b.Build().JobDataMap()

			So(dm.Get("key"), ShouldEqual, "value")
			So(dm.Contains("nonexists"), ShouldBeFalse)
		})
	})
}
