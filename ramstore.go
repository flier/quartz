package quartz

import (
	"fmt"
	"sync"
)

type TriggerState int

const (
	STATE_WAITING TriggerState = iota
	STATE_ACQUIRED
	STATE_EXECUTING
	STATE_COMPLETE
	STATE_PAUSED
	STATE_BLOCKED
	STATE_PAUSED_BLOCKED
	STATE_ERROR
)

func jobAlreadyExistsError(job JobDetail) error {
	return fmt.Errorf("Unable to store Job : '%s', because one already exists with this identification.", job.Key())
}

func triggerAlreadyExistsError(trigger Trigger) error {
	return fmt.Errorf("Unable to store Trigger with name: '%s' and group: '%s', "+
		"because one already exists with this identification.", trigger.Key().Name, trigger.Key().Group)
}

func jobPersistenceError(key JobKey) error {
	return fmt.Errorf("The job (%s) referenced by the trigger does not exist.", key.String())
}

type jobWrapper struct {
	jobDetail JobDetail
}

func (w *jobWrapper) Key() JobKey { return w.jobDetail.Key() }

type triggerWrapper struct {
	trigger OperableTrigger

	state TriggerState
}

func (w *triggerWrapper) Key() TriggerKey { return w.trigger.Key() }

func (w *triggerWrapper) JobKey() JobKey { return w.trigger.JobKey() }

type JobMap map[string]*jobWrapper

type TriggerMap map[string]*triggerWrapper

type RAMJobStore struct {
	lock                sync.Mutex
	jobsByKey           JobMap
	triggersByKey       TriggerMap
	jobsByGroup         map[string]JobMap
	triggersByGroup     map[string]TriggerMap
	timeTriggers        TreeSet
	triggers            []*triggerWrapper
	pausedTriggerGroups HashSet
	pausedJobGroups     HashSet
	blockedJobs         HashSet
}

func NewRAMJobStore() *RAMJobStore {
	return &RAMJobStore{
		jobsByKey:       make(JobMap),
		triggersByKey:   make(TriggerMap),
		jobsByGroup:     make(map[string]JobMap),
		triggersByGroup: make(map[string]TriggerMap),
		timeTriggers: TreeSet{
			compare: func(lhs, rhs interface{}) bool {
				return lhs.(Trigger).Key().String() < rhs.(Trigger).Key().String()
			},
		},
		pausedTriggerGroups: NewHashSet(),
		pausedJobGroups:     NewHashSet(),
		blockedJobs:         NewHashSet(),
	}
}

func (s *RAMJobStore) SchedulerStarted() error { return nil }

func (s *RAMJobStore) SchedulerPaused() {}

func (s *RAMJobStore) SchedulerResumed() {}

func (s *RAMJobStore) SupportsPersistence() bool { return false }

func (s *RAMJobStore) Clustered() bool { return false }

func (s *RAMJobStore) StoreJobAndTrigger(job JobDetail, trigger OperableTrigger) error {
	if err := s.StoreJob(job, false); err != nil {
		return err
	}

	if err := s.StoreTrigger(trigger, false); err != nil {
		return err
	}

	return nil
}

func (s *RAMJobStore) StoreJobsAndTriggers(triggersAndJobs map[JobDetail][]Trigger, replace bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !replace {
		for job, triggers := range triggersAndJobs {
			if s.CheckJobExists(job.Key()) {
				return jobAlreadyExistsError(job)
			}

			for _, trigger := range triggers {
				if s.CheckTriggerExists(trigger.Key()) {
					return triggerAlreadyExistsError(trigger)
				}
			}
		}
	}

	for job, triggers := range triggersAndJobs {
		if err := s.StoreJob(job, true); err != nil {
			return err
		}

		for _, trigger := range triggers {
			if err := s.StoreTrigger(trigger.(OperableTrigger), true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RAMJobStore) StoreJob(jobDetail JobDetail, replaceExisting bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	jw, exists := s.jobsByKey[jobDetail.Key().String()]

	if exists {
		if !replaceExisting {
			return jobAlreadyExistsError(jobDetail)
		}
	}

	if jw == nil {
		grpMap, exists := s.jobsByGroup[jobDetail.Key().Group()]

		if !exists {
			grpMap = make(JobMap)

			s.jobsByGroup[jobDetail.Key().Group()] = grpMap
		}

		jw = &jobWrapper{jobDetail}

		grpMap[jobDetail.Key().String()] = jw
		s.jobsByKey[jobDetail.Key().String()] = jw
	} else {
		jw.jobDetail = jobDetail
	}

	return nil
}

func (s *RAMJobStore) RemoveJob(key JobKey) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return false, nil
}

func (s *RAMJobStore) StoreTrigger(trigger OperableTrigger, replaceExisting bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, exists := s.triggersByKey[trigger.Key().String()]

	if exists {
		if !replaceExisting {
			return triggerAlreadyExistsError(trigger)
		}

		if _, err := s.removeTrigger(trigger.Key(), false); err != nil {
			return err
		}
	}

	if job, err := s.RetrieveJob(trigger.JobKey()); err != nil {
		return err
	} else if job == nil {
		return jobPersistenceError(trigger.JobKey())
	}

	tw := &triggerWrapper{trigger: trigger}

	s.triggers = append(s.triggers, tw)

	grpMap, exists := s.triggersByGroup[trigger.Key().Group()]

	if !exists {
		grpMap = make(TriggerMap)

		s.triggersByGroup[trigger.Key().Group()] = grpMap
	}

	grpMap[trigger.Key().String()] = tw

	s.triggersByKey[trigger.Key().String()] = tw

	if s.pausedTriggerGroups.Contains(trigger.Key().Group()) || s.pausedJobGroups.Contains(trigger.JobKey().Group()) {
		if s.blockedJobs.Contains(trigger.JobKey().String()) {
			tw.state = STATE_PAUSED_BLOCKED
		} else {
			tw.state = STATE_PAUSED
		}
	} else if s.blockedJobs.Contains(trigger.JobKey().String()) {
		tw.state = STATE_BLOCKED
	} else {
		s.timeTriggers.Add(tw)
	}

	return nil
}

func (s *RAMJobStore) RemoveTrigger(key TriggerKey) (bool, error) {
	return s.removeTrigger(key, true)
}

func (s *RAMJobStore) removeTrigger(key TriggerKey, removeOrphanedJob bool) (bool, error) {
	return false, nil
}

func (s *RAMJobStore) RemoveJobs(keys []JobKey) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	allFound := true

	for _, key := range keys {
		if found, err := s.RemoveJob(key); err != nil {
			return false, err
		} else {
			allFound = found && allFound
		}
	}

	return allFound, nil
}

func (s *RAMJobStore) RemoveTriggers(keys []TriggerKey) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	allFound := true

	for _, key := range keys {
		if found, err := s.RemoveTrigger(key); err != nil {
			return false, err
		} else {
			allFound = found && allFound
		}
	}

	return allFound, nil
}

func (s *RAMJobStore) RetrieveJob(key JobKey) (JobDetail, error) {
	return nil, nil
}

func (s *RAMJobStore) TriggersForJob(key JobKey) []OperableTrigger {
	return nil
}

func (s *RAMJobStore) CheckJobExists(key JobKey) bool {
	return false
}

func (s *RAMJobStore) CheckTriggerExists(key TriggerKey) bool {
	return false
}
