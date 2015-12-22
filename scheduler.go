package quartz

import (
	"time"
)

type Scheduler interface {
}

type ScheduleBuilder interface {
	Build() MutableTrigger
}

type SimpleScheduleBuilder struct {
	repeatInterval time.Duration
	repeatCount    int
}

func (b *SimpleScheduleBuilder) Build() MutableTrigger {
	return &simpleTrigger{
		repeatInterval: b.repeatInterval,
		repeatCount:    b.repeatCount,
	}
}
