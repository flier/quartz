package quartz

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	zero time.Time
)

const (
	REPEAT_INDEFINITELY = -1
)

// The base interface with properties common to all Triggers -
// use TriggerBuilder to instantiate an actual Trigger.
type Trigger interface {
	Key() TriggerKey

	JobKey() JobKey

	Description() string

	JobDataMap() JobDataMap

	Priority() int

	MayFireAgain() bool

	StartTime() time.Time

	EndTime() time.Time

	NextFireTime() time.Time

	PreviousFireTime() time.Time

	FireTimeAfter(afterTime time.Time) time.Time

	FinalFireTime() time.Time

	TriggerBuilder() *TriggerBuilder

	ScheduleBuilder() ScheduleBuilder
}

type MutableTrigger interface {
	Trigger

	SetKey(key TriggerKey)

	SetJobKey(key JobKey)

	SetDescription(desc string)

	SetPriority(priority int)

	SetStartTime(startTime time.Time) error

	SetEndTime(endTime time.Time) error

	SetJobDataMap(dataMap JobDataMap)
}

type OperableTrigger interface {
	Cloneable
	MutableTrigger

	SetNextFireTime(nextFireTime time.Time)

	SetPreviousFireTime(previousFireTime time.Time)
}

type TriggerKey []byte

func NewTriggerKey(name string) TriggerKey {
	return NewGroupTriggerKey(name, DEFAULT_GROUP)
}

func NewGroupTriggerKey(name, group string) TriggerKey {
	return TriggerKey(fmt.Sprintf("%s.%s", group, name))
}

func NewUniqueTriggerKey(group string) TriggerKey {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	return NewGroupTriggerKey(newUniqueName(group), group)
}

func (key TriggerKey) Name() string                 { return strings.Split(string(key), ".")[1] }
func (key TriggerKey) Group() string                { return strings.Split(string(key), ".")[0] }
func (key TriggerKey) String() string               { return string(key) }
func (key TriggerKey) Equals(other TriggerKey) bool { return bytes.Equal(key, other) }

type abstractTrigger struct {
	name     string
	group    string
	jobName  string
	jobGroup string
	desc     string
	dataMap  JobDataMap
	priority int
	key      TriggerKey
}

func (t *abstractTrigger) Key() TriggerKey {
	if t.key == nil {
		if t.name == "" {
			return nil
		}

		t.key = NewGroupTriggerKey(t.name, t.group)
	}

	return t.key
}

func (t *abstractTrigger) SetKey(key TriggerKey) {
	t.name = key.Name()
	t.group = key.Group()
	t.key = key
}

func (t *abstractTrigger) JobKey() JobKey {
	if t.jobName == "" {
		return nil
	}

	return NewGroupJobKey(t.jobName, t.jobGroup)
}

func (t *abstractTrigger) SetJobKey(key JobKey) {
	t.jobName = key.Name()
	t.jobGroup = key.Group()
}

func (t *abstractTrigger) FullName() string { return t.Key().String() }

func (t *abstractTrigger) FullJobName() string { return t.JobKey().String() }

func (t *abstractTrigger) Description() string { return t.desc }

func (t *abstractTrigger) SetDescription(desc string) { t.desc = desc }

func (t *abstractTrigger) JobDataMap() JobDataMap {
	if t.dataMap == nil {
		t.dataMap = NewJobDataMap()
	}

	return t.dataMap
}

func (t *abstractTrigger) SetJobDataMap(dataMap JobDataMap) { t.dataMap = dataMap }

func (t *abstractTrigger) Priority() int { return t.priority }

func (t *abstractTrigger) SetPriority(priority int) { t.priority = priority }

type simpleTrigger struct {
	abstractTrigger

	startTime        time.Time
	endTime          time.Time
	nextFireTime     time.Time
	previousFireTime time.Time
	repeatInterval   time.Duration
	repeatCount      int
	timesTriggered   int
	complete         bool
}

func (t *simpleTrigger) StartTime() time.Time { return t.startTime }

func (t *simpleTrigger) SetStartTime(startTime time.Time) error {
	if startTime.IsZero() {
		return errors.New("Start time cannot be null")
	}

	if !t.endTime.IsZero() && !startTime.IsZero() && t.endTime.Before(startTime) {
		return errors.New("End time cannot be before start time")
	}

	t.startTime = startTime

	return nil
}

func (t *simpleTrigger) EndTime() time.Time { return t.endTime }

func (t *simpleTrigger) SetEndTime(endTime time.Time) error {
	if !t.startTime.IsZero() && !endTime.IsZero() && t.startTime.After(endTime) {
		return errors.New("End time cannot be before start time")
	}

	t.endTime = endTime

	return nil
}

func (t *simpleTrigger) NextFireTime() time.Time { return t.nextFireTime }

func (t *simpleTrigger) SetNextFireTime(nextFireTime time.Time) { t.nextFireTime = nextFireTime }

func (t *simpleTrigger) PreviousFireTime() time.Time { return t.previousFireTime }

func (t *simpleTrigger) SetPreviousFireTime(previousFireTime time.Time) {
	t.previousFireTime = previousFireTime
}

func (t *simpleTrigger) FireTimeAfter(afterTime time.Time) time.Time {
	if t.complete {
		return zero
	}

	if t.timesTriggered > t.repeatCount && t.repeatCount != REPEAT_INDEFINITELY {
		return zero
	}

	if afterTime.IsZero() {
		afterTime = time.Now()
	}

	if t.repeatCount == 0 && afterTime.After(t.startTime) {
		return zero
	}

	if !t.endTime.IsZero() && t.endTime.Before(afterTime) {
		return zero
	}

	if afterTime.Before(t.startTime) {
		return t.startTime
	}

	numberOfTimesExecuted := int(afterTime.Sub(t.startTime)/t.repeatInterval) + 1

	if numberOfTimesExecuted > t.repeatCount && t.repeatCount != REPEAT_INDEFINITELY {
		return zero
	}

	fireTime := t.startTime.Add(time.Duration(numberOfTimesExecuted) * t.repeatInterval)

	if t.endTime.Before(fireTime) {
		return zero
	}

	return fireTime
}

func (t *simpleTrigger) FireTimeBefore(endTime time.Time) time.Time {
	if endTime.Before(t.startTime) {
		return zero
	}

	numFires := t.computeNumTimesFiredBetween(t.startTime, endTime)

	return t.startTime.Add(time.Duration(numFires) * t.repeatInterval)
}

func (t *simpleTrigger) MayFireAgain() bool { return !t.NextFireTime().IsZero() }

func (t *simpleTrigger) computeNumTimesFiredBetween(start, end time.Time) int {
	if t.repeatInterval < time.Millisecond {
		return 0
	}

	return int(end.Sub(start) / t.repeatInterval)
}

func (t *simpleTrigger) FinalFireTime() time.Time {
	if t.repeatCount == 0 {
		return t.startTime
	}

	if t.repeatCount == REPEAT_INDEFINITELY {
		if t.endTime.IsZero() {
			return zero
		}

		return t.FireTimeBefore(t.endTime)
	}

	lastTrigger := t.startTime.Add(time.Duration(t.repeatCount) * t.repeatInterval)

	if t.endTime.IsZero() || lastTrigger.Before(t.endTime) {
		return lastTrigger
	}

	return t.FireTimeBefore(t.endTime)
}

func (t *simpleTrigger) TriggerBuilder() *TriggerBuilder {
	return &TriggerBuilder{
		Key:             t.Key(),
		Description:     t.desc,
		StartTime:       t.startTime,
		EndTime:         t.endTime,
		Priority:        t.priority,
		JobKey:          t.JobKey(),
		DataMap:         t.dataMap,
		ScheduleBuilder: t.ScheduleBuilder(),
	}
}

func (t *simpleTrigger) ScheduleBuilder() ScheduleBuilder {
	return &SimpleScheduleBuilder{
		repeatInterval: t.repeatInterval,
		repeatCount:    t.repeatCount,
	}
}

// TriggerBuilder is used to instantiate Triggers.
type TriggerBuilder struct {
	Key                TriggerKey
	Description        string
	StartTime, EndTime time.Time
	Priority           int
	JobKey             JobKey
	DataMap            JobDataMap
	ScheduleBuilder    ScheduleBuilder
}

func (b *TriggerBuilder) WithIdentity(name string) *TriggerBuilder {
	b.Key = NewTriggerKey(name)

	return b
}

func (b *TriggerBuilder) WithGroupIdentity(name, group string) *TriggerBuilder {
	b.Key = NewGroupTriggerKey(name, group)

	return b
}

func (b *TriggerBuilder) WithTriggerKey(key TriggerKey) *TriggerBuilder {
	b.Key = key

	return b
}

func (b *TriggerBuilder) WithDescription(desc string) *TriggerBuilder {
	b.Description = desc

	return b
}

func (b *TriggerBuilder) WithPriority(priority int) *TriggerBuilder {
	b.Priority = priority

	return b
}

func (b *TriggerBuilder) StartAt(startTime time.Time) *TriggerBuilder {
	b.StartTime = startTime

	return b
}

func (b *TriggerBuilder) StartNow() *TriggerBuilder {
	b.StartTime = time.Now()

	return b
}

func (b *TriggerBuilder) EndAt(endTime time.Time) *TriggerBuilder {
	b.EndTime = endTime

	return b
}

func (b *TriggerBuilder) WithSchedule(scheduleBuilder ScheduleBuilder) *TriggerBuilder {
	b.ScheduleBuilder = scheduleBuilder

	return b
}

func (b *TriggerBuilder) ForJob(name string) *TriggerBuilder {
	b.JobKey = NewJobKey(name)

	return b
}

func (b *TriggerBuilder) ForGroupJob(name, group string) *TriggerBuilder {
	b.JobKey = NewGroupJobKey(name, group)

	return b
}

func (b *TriggerBuilder) ForJobKey(jobKey JobKey) *TriggerBuilder {
	b.JobKey = jobKey

	return b
}

func (b *TriggerBuilder) ForJobDetail(jobDetail JobDetail) *TriggerBuilder {
	b.JobKey = jobDetail.Key()

	return b
}

func (b *TriggerBuilder) UsingJobData(key string, value interface{}) *TriggerBuilder {
	if b.DataMap == nil {
		b.DataMap = NewJobDataMap()
	}

	b.DataMap.Put(key, value)

	return b
}

func (b *TriggerBuilder) UsingJobDataMap(dataMap JobDataMap) *TriggerBuilder {
	if b.DataMap == nil {
		b.DataMap = NewJobDataMap()
	}

	b.DataMap.PutAll(dataMap)

	return b
}

func (b *TriggerBuilder) SetJobDataMap(dataMap JobDataMap) *TriggerBuilder {
	b.DataMap = dataMap

	return b
}

func (b *TriggerBuilder) Build() Trigger {
	if b.ScheduleBuilder == nil {
		b.ScheduleBuilder = &SimpleScheduleBuilder{}
	}

	trigger := b.ScheduleBuilder.Build()

	trigger.SetDescription(b.Description)
	trigger.SetStartTime(b.StartTime)
	trigger.SetEndTime(b.EndTime)

	if b.Key == nil {
		b.Key = NewUniqueTriggerKey("")
	}

	trigger.SetKey(b.Key)

	if b.JobKey != nil {
		trigger.SetJobKey(b.JobKey)
	}

	trigger.SetPriority(b.Priority)

	if b.DataMap != nil {
		trigger.SetJobDataMap(b.DataMap)
	}

	return trigger
}
