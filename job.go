package quartz

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

//
// The interface to be implemented by classes which represent a 'job' to be performed.
//
type Job interface {
	// Called by the Scheduler when a Trigger fires that is associated with the Job.
	Execute(context JobExecutionContext)
}

//
// A context bundle containing handles to various environment information,
// that is given to a JobDetail instance as it is executed,
// and to a Trigger instance after the execution completes.
//
type JobExecutionContext interface {
	Scheduler() Scheduler

	Trigger() Trigger

	JobInstance() Job

	JobDetail() JobDetail

	FireTime() time.Time

	ScheduledFireTime() time.Time

	PreviousFireTime() time.Time

	NextFireTime() time.Time

	JobRunTime() time.Duration

	Result() interface{}

	SetResult(interface{})

	MergedJobDataMap() JobDataMap

	Put(key string, value interface{})

	Get(key string) interface{}
}

//
// Conveys the detail properties of a given Job instance.
// JobDetails are to be created/defined with JobBuilder.
//
type JobDetail interface {
	Cloneable

	Key() JobKey

	Description() string

	Durable() bool

	JobDataMap() JobDataMap

	JobBuilder() *JobBuilder
}

type JobDataMap interface {
	DirtyFlagMap
}

type JobFactory interface {
	NewJob(scheduler Scheduler) (Job, error)
}

type JobKey []byte

func NewJobKey(name string) JobKey {
	return NewGroupJobKey(name, DEFAULT_GROUP)
}

func NewGroupJobKey(name, group string) JobKey {
	return JobKey(fmt.Sprintf("%s.%s", group, name))
}

func NewUniqueKey(group string) JobKey {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	return NewGroupJobKey(newUniqueName(group), group)
}

func (key JobKey) Name() string             { return strings.Split(string(key), ".")[1] }
func (key JobKey) Group() string            { return strings.Split(string(key), ".")[0] }
func (key JobKey) String() string           { return string(key) }
func (key JobKey) Equals(other JobKey) bool { return bytes.Equal(key, other) }

type jobDetail struct {
	key     JobKey
	desc    string
	durable bool
	dataMap JobDataMap
	builder *JobBuilder
}

func (d *jobDetail) Key() JobKey { return d.key }

func (d *jobDetail) Description() string { return d.desc }

func (d *jobDetail) Durable() bool { return d.durable }

func (d *jobDetail) JobDataMap() JobDataMap { return d.dataMap }

func (d *jobDetail) JobBuilder() *JobBuilder { return d.builder }

func (d *jobDetail) Clone() interface{} {
	clone := *d

	if d.dataMap != nil {
		clone.dataMap = d.dataMap.Clone().(JobDataMap)
	}

	return &clone
}

func NewJobDataMap() JobDataMap {
	return JobDataMap(NewDirtyFlagMap())
}

//
// JobBuilder is used to instantiate JobDetails.
//
type JobBuilder struct {
	Key         JobKey
	Description string
	DataMap     JobDataMap
}

func (b *JobBuilder) WithIdentity(name string) *JobBuilder {
	b.Key = NewJobKey(name)

	return b
}

func (b *JobBuilder) WithGroupIdentity(name, group string) *JobBuilder {
	b.Key = NewGroupJobKey(name, group)

	return b
}

func (b *JobBuilder) WithJobKey(key JobKey) *JobBuilder {
	b.Key = key

	return b
}

func (b *JobBuilder) WithDescription(desc string) *JobBuilder {
	b.Description = desc

	return b
}

func (b *JobBuilder) UsingJobData(key string, value interface{}) *JobBuilder {
	if b.DataMap == nil {
		b.DataMap = NewJobDataMap()
	}

	b.DataMap.Put(key, value)

	return b
}

func (b *JobBuilder) UsingJobDataMap(dataMap JobDataMap) *JobBuilder {
	if b.DataMap == nil {
		b.DataMap = NewJobDataMap()
	}

	b.DataMap.PutAll(dataMap)

	return b
}

func (b *JobBuilder) SetJobDataMap(dataMap JobDataMap) *JobBuilder {
	b.DataMap = dataMap

	return b
}

func (b *JobBuilder) Build() JobDetail {
	job := &jobDetail{
		key:     b.Key,
		desc:    b.Description,
		dataMap: b.DataMap,
		builder: b,
	}

	if job.key == nil {
		job.key = NewUniqueKey("")
	}

	return job
}
