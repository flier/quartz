package quartz

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTriggerKey(t *testing.T) {
	Convey("Given a TriggerKey", t, func() {
		key := NewTriggerKey("name")

		So(key.Name(), ShouldEqual, "name")
		So(key.Group(), ShouldEqual, DEFAULT_GROUP)
		So(key.String(), ShouldEqual, "DEFAULT.name")
	})

	Convey("Given a grouped TriggerKey", t, func() {
		key := NewGroupTriggerKey("name", "group")

		So(key.Name(), ShouldEqual, "name")
		So(key.Group(), ShouldEqual, "group")
		So(key.String(), ShouldEqual, "group.name")
	})

	Convey("Given a uniqued TriggerKey", t, func(c C) {
		key := NewUniqueKey("")

		c.Printf("unique key: %s", key)

		So(len(key.Name()), ShouldEqual, 41)
		So(key.Group(), ShouldEqual, DEFAULT_GROUP)
	})
}

func TestTriggerBuilder(t *testing.T) {
	Convey("Given a TriggerBuilder, then build Trigger", t, func() {
		b := &TriggerBuilder{}

		Convey("TriggerBuilder -> Trigger.TriggerBuilder()", func() {
			So(b.Build().TriggerBuilder(), ShouldResemble, b)
		})

		Convey("WithIdentity -> Trigger.Key()", func() {
			b.WithIdentity("name")

			So(b.Build().Key().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("WithGroupIdentity -> Trigger.Key()", func() {
			b.WithGroupIdentity("name", "group")

			So(b.Build().Key().String(), ShouldEqual, "group.name")
		})

		Convey("WithTriggerKey -> Trigger.Key()", func() {
			b.WithTriggerKey(NewTriggerKey("name"))

			So(b.Build().Key().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("WithDescription -> Trigger.Description()", func() {
			b.WithDescription("desc")

			So(b.Build().Description(), ShouldEqual, "desc")
		})

		Convey("WithPriority -> Trigger.Priority()", func() {
			b.WithPriority(1)

			So(b.Build().Priority(), ShouldEqual, 1)
		})

		Convey("StartAt -> Trigger.StartTime()", func() {
			ts := time.Now()

			b.StartAt(ts)

			So(b.Build().StartTime(), ShouldResemble, ts)
		})

		Convey("EndAt -> Trigger.EndTime()", func() {
			ts := time.Now()

			b.EndAt(ts)

			So(b.Build().EndTime(), ShouldResemble, ts)
		})

		Convey("WithSchedule -> Trigger.ScheduleBuilder()", func() {
			sb := &SimpleScheduleBuilder{10 * time.Second, 100}

			b.WithSchedule(sb)

			So(b.Build().ScheduleBuilder(), ShouldResemble, sb)
		})

		Convey("ForJob -> Trigger.JobKey()", func() {
			b.ForJob("name")

			So(b.Build().JobKey().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("ForGroupJob -> Trigger.JobKey()", func() {
			b.ForGroupJob("name", "group")

			So(b.Build().JobKey().String(), ShouldEqual, "group.name")
		})

		Convey("ForJobKey -> Trigger.JobKey()", func() {
			b.ForJobKey(NewJobKey("name"))

			So(b.Build().JobKey().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("ForJobDetail -> Trigger.JobKey()", func() {
			b.ForJobDetail(&jobDetail{key: NewJobKey("name")})

			So(b.Build().JobKey().String(), ShouldEqual, "DEFAULT.name")
		})

		Convey("UsingJobData -> Trigger.JobDataMap()", func() {
			b.UsingJobData("key", "value")

			So(b.Build().JobDataMap().Get("key"), ShouldEqual, "value")
		})

		Convey("UsingJobDataMap -> Trigger.JobDataMap()", func() {
			m := NewJobDataMap()
			m.Put("key", "value")

			b.UsingJobDataMap(m)

			So(b.Build().JobDataMap().Get("key"), ShouldEqual, "value")
		})

		Convey("SetJobDataMap -> Trigger.JobDataMap()", func() {
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
