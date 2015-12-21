package quartz

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestJobKey(t *testing.T) {
	Convey("Given a JobKey", t, func() {
		key := NewJobKey("name")

		So(key.Name(), ShouldEqual, "name")
	})
}
