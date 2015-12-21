package quartz

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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
}

//
// Conveys the detail properties of a given Job instance.
// JobDetails are to be created/defined with JobBuilder.
//
type JobDetail interface {
	Key() JobKey
}

const (
	DEFAULT_GROUP = "DEFAULT"
)

type JobKey []string

func NewJobKey(name string) JobKey {
	return []string{DEFAULT_GROUP, name}
}

func NewJobGroupKey(group, name string) JobKey {
	return []string{group, name}
}

func NewUniqueName(group string) JobKey {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	buf := make([]byte, 16)

	rand.Read(buf)

	hash := md5.Sum([]byte(group))

	name := fmt.Sprintf("%s-%s", hex.EncodeToString(hash[12:]), hex.EncodeToString(buf))

	return []string{group, name}
}

func (key JobKey) Name() string { return key[1] }

func (key JobKey) Group() string { return key[0] }

func (key JobKey) String() string { return strings.Join(key, ".") }

//
// JobBuilder is used to instantiate JobDetails.
//
type JobBuilder struct {
}

func NewJob() *JobBuilder {
	return &JobBuilder{}
}

func (b *JobBuilder) Build() JobDetail {
	return nil
}
