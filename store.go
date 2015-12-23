package quartz

//
// The interface to be implemented by classes that want to provide a Job and Trigger storage mechanism for the QuartzScheduler's use.
type JobStore interface {
	SchedulerStarted() error

	SchedulerPaused()

	SchedulerResumed()

	Shutdown()

	SupportsPersistence() bool

	Clustered() bool

	StoreJobAndTrigger(job JobDetail, trigger OperableTrigger) error

	StoreJobsAndTriggers(triggersAndJobs map[JobDetail][]Trigger, replace bool) error

	StoreJob(job JobDetail, replaceExisting bool) error

	StoreTrigger(trigger OperableTrigger, replaceExisting bool) error

	RemoveJob(key JobKey) (bool, error)

	RemoveJobs(keys []JobKey) (bool, error)

	RetrieveJob(key JobKey) (JobDetail, error)

	RemoveTrigger(key TriggerKey) (bool, error)

	RemoveTriggers(keys []TriggerKey) (bool, error)

	ReplaceTrigger(key TriggerKey, trigger OperableTrigger) error

	RetrieveTrigger(key TriggerKey) (OperableTrigger, error)

	CheckJobExists(key JobKey) bool

	CheckTriggerExists(key TriggerKey) bool

	NumberOfJobs() int

	NumberOfTriggers() int

	TriggersForJob(key JobKey) []OperableTrigger

	PauseJob(key JobKey) error

	PauseTrigger(key TriggerKey) error

	ResumeJob(key JobKey) error

	ResumeTrigger(key TriggerKey) error

	PauseAll() error

	ResumeAll() error
}
