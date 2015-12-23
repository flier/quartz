package quartz

import (
	"time"
)

type Scheduler interface {
	Name() string

	Context() SchedulerContext

	Start() error

	StartDelayed(time.Duration) error

	Started() bool

	Standby() error

	InStandbyMode() bool

	Shutdown() error

	IsShutdown() bool

	MetaData() SchedulerMetaData

	CurrentlyExecutingJob() ([]JobExecutionContext, error)

	SetJobFactory(factory JobFactory)

	ScheduleJob(jobDetail JobDetail, trigger Trigger) (time.Time, error)

	Schedule(trigger Trigger) (time.Time, error)

	ScheduleJobs(triggersAndJobs map[JobDetail][]Trigger, replace bool) (time.Time, error)

	UnscheduleJob(key TriggerKey) (bool, error)

	UnscheduleJobs(keys []TriggerKey) (bool, error)

	RescheduleJob(key TriggerKey, trigger Trigger) (time.Time, error)

	AddJob(jobDetail JobDetail, replace bool) error

	DeleteJob(key JobKey) (bool, error)

	DeleteJobs(keys []JobKey) (bool, error)

	TriggerJob(key JobKey) error

	PauseJob(key JobKey) error

	PauseTrigger(key TriggerKey) error

	ResumeJob(key JobKey) error

	ResumeTrigger(key TriggerKey) error

	PauseAll() error

	ResumeAll() error

	GetTriggersOfJob(key JobKey) []Trigger

	GetJobDetail(key JobKey) JobDetail

	GetTrigger(key TriggerKey) Trigger

	CheckJobExists(key JobKey) bool

	CheckTriggerExists(key TriggerKey) bool

	Clear() error
}

type SchedulerContext interface {
	DirtyFlagMap
}

type SchedulerMetaData interface {
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

type QuartzScheduler struct {
}

type StdScheduler struct {
}
