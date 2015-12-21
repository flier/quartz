package quartz

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
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
	Key() JobKey

	Description() string

	JobDataMap() JobDataMap

	JobBuilder() *JobBuilder
}

type JobDataEntry interface {
	Key() string

	Value() interface{}
}

type JobDataMap interface {
	Dirty() bool

	ClearDirtyFlag()

	Empty() bool

	Size() int

	Keys() []string

	Values() []interface{}

	Entries() []JobDataEntry

	Contains(key string) bool

	Get(key string) interface{}

	Put(key string, value interface{}) JobDataMap

	PutAll(dataMap JobDataMap) JobDataMap
}

const (
	DEFAULT_GROUP = "DEFAULT"
)

type JobKey []string

func NewJobKey(name string) JobKey {
	return []string{DEFAULT_GROUP, name}
}

func NewJobGroupKey(name, group string) JobKey {
	return []string{group, name}
}

func NewUniqueKey(group string) JobKey {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	buf := make([]byte, 16)

	rand.Read(buf)

	hash := md5.Sum([]byte(group))

	name := fmt.Sprintf("%s-%s", hex.EncodeToString(hash[12:]), hex.EncodeToString(buf))

	return []string{group, name}
}

func (key JobKey) Group() string  { return key[0] }
func (key JobKey) Name() string   { return key[1] }
func (key JobKey) String() string { return strings.Join(key, ".") }

type jobDetail struct {
	key     JobKey
	desc    string
	dataMap JobDataMap
	builder *JobBuilder
}

func (d *jobDetail) Key() JobKey { return d.key }

func (d *jobDetail) Description() string { return d.desc }

func (d *jobDetail) JobDataMap() JobDataMap { return d.dataMap }

func (d *jobDetail) JobBuilder() *JobBuilder { return d.builder }

type jobDataEntry struct {
	key   string
	value interface{}
}

func (e *jobDataEntry) Key() string { return e.key }

func (e *jobDataEntry) Value() interface{} { return e.value }

type jobDataMap struct {
	entries map[string]interface{}
	dirty   bool
}

func NewJobDataMap() JobDataMap {
	return &jobDataMap{entries: make(map[string]interface{})}
}

func (m *jobDataMap) Dirty() bool { return m.dirty }

func (m *jobDataMap) ClearDirtyFlag() { m.dirty = false }

func (m *jobDataMap) Empty() bool { return len(m.entries) == 0 }

func (m *jobDataMap) Size() int { return len(m.entries) }

func (m *jobDataMap) Keys() (keys []string) {
	for key, _ := range m.entries {
		keys = append(keys, key)
	}

	sort.Sort(sort.StringSlice(keys))

	return
}

func (m *jobDataMap) Values() (values []interface{}) {
	for _, value := range m.entries {
		values = append(values, value)
	}

	return
}

func (m *jobDataMap) Entries() (entries []JobDataEntry) {
	for key, value := range m.entries {
		entries = append(entries, &jobDataEntry{key, value})
	}

	return
}

func (m *jobDataMap) Contains(key string) bool {
	_, exists := m.entries[key]

	return exists
}

func (m *jobDataMap) Get(key string) interface{} { return m.entries[key] }

func (m *jobDataMap) Put(key string, value interface{}) JobDataMap {
	if v, exists := m.entries[key]; !exists || v != value {
		m.entries[key] = value
		m.dirty = true
	}

	return m
}

func (m *jobDataMap) PutAll(dataMap JobDataMap) JobDataMap {
	for _, entry := range dataMap.Entries() {
		m.Put(entry.Key(), entry.Value())
	}

	return m
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
	b.Key = NewJobGroupKey(name, group)

	return b
}

func (b *JobBuilder) WithKey(key JobKey) *JobBuilder {
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
