package quartz

import (
	"fmt"
	"strings"
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
	timeTriggers        Set
	triggers            []*triggerWrapper
	pausedTriggerGroups Set
	pausedJobGroups     Set
	blockedJobs         Set
}

func NewRAMJobStore() *RAMJobStore {
	return &RAMJobStore{
		jobsByKey:       make(JobMap),
		triggersByKey:   make(TriggerMap),
		jobsByGroup:     make(map[string]JobMap),
		triggersByGroup: make(map[string]TriggerMap),
		timeTriggers: NewTreeSet(func(lhs, rhs interface{}) int {
			return strings.Compare(lhs.(Trigger).Key().String(), rhs.(Trigger).Key().String())
		}),
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

		s.removeTrigger(trigger.Key(), false)
	}

	if job := s.RetrieveJob(trigger.JobKey()); job == nil {
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

func (s *RAMJobStore) RemoveTrigger(key TriggerKey) bool {
	return s.removeTrigger(key, true)
}

func (s *RAMJobStore) removeTrigger(key TriggerKey, removeOrphanedJob bool) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, exists := s.triggersByKey[key.String()]

	if exists {
		delete(s.triggersByKey, key.String())

		if triggers, exists := s.triggersByGroup[key.Group()]; exists && len(triggers) > 0 {
			delete(triggers, key.String())

			if len(triggers) == 0 {
				delete(s.triggersByGroup, key.Group())
			}
		}

		var tw *triggerWrapper

		for i, trigger := range s.triggers {
			if !trigger.Key().Equals(key) {
				tw = trigger
				s.triggers = append(s.triggers[:i], s.triggers[i+1:]...)
			}
		}

		s.timeTriggers.Remove(tw)

		if removeOrphanedJob {
			jw, exists := s.jobsByKey[tw.JobKey().String()]
			triggers := s.TriggersForJob(tw.JobKey())

			if len(triggers) == 0 && exists && jw.jobDetail.Durable() {
				s.RemoveJob(jw.Key())
			}
		}
	}

	return exists
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
		allFound = s.RemoveTrigger(key) && allFound
	}

	return allFound, nil
}

func (s *RAMJobStore) RetrieveJob(key JobKey) JobDetail {
	s.lock.Lock()
	defer s.lock.Unlock()

	if jw, exists := s.jobsByKey[key.String()]; exists {
		return jw.jobDetail.Clone().(JobDetail)
	}

	return nil
}

func (s *RAMJobStore) RetrieveTrigger(key TriggerKey) OperableTrigger {
	s.lock.Lock()
	defer s.lock.Unlock()

	if tw, exists := s.triggersByKey[key.String()]; exists {
		return tw.trigger.Clone().(OperableTrigger)
	}

	return nil
}

func (s *RAMJobStore) TriggersForJob(key JobKey) (triggers []OperableTrigger) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, tw := range s.triggers {
		if tw.JobKey().Equals(key) {
			triggers = append(triggers, tw.trigger.Clone().(OperableTrigger))
		}
	}

	return
}

func (s *RAMJobStore) CheckJobExists(key JobKey) bool {
	s.lock.Lock()
	jw, exists := s.jobsByKey[key.String()]
	s.lock.Unlock()

	return exists && jw != nil
}

func (s *RAMJobStore) CheckTriggerExists(key TriggerKey) bool {
	s.lock.Lock()
	tw, exists := s.triggersByKey[key.String()]
	s.lock.Unlock()

	return exists && tw != nil
}
