package main

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/intelops/go-common/logging"
)

type jobHandler interface {
	CronSpec() string
	Run()
}

type Scheduler struct {
	log       logging.Logger
	jobs      map[string]jobHandler
	cronIDs   map[string]cron.EntryID
	c         *cron.Cron
	cronMutex *sync.Mutex
}

func NewScheduler(log logging.Logger) *Scheduler {
	clog := cron.VerbosePrintfLogger(log.(logging.StdLogger))
	return &Scheduler{
		log:       log,
		c:         cron.New(cron.WithChain(cron.SkipIfStillRunning(clog), cron.Recover(clog))),
		jobs:      map[string]jobHandler{},
		cronIDs:   map[string]cron.EntryID{},
		cronMutex: &sync.Mutex{},
	}
}

func (t *Scheduler) AddJob(jobName string, job jobHandler) error {
	t.cronMutex.Lock()
	defer t.cronMutex.Unlock()
	_, ok := t.cronIDs[jobName]
	if ok {
		return errors.Errorf("%s job already exists", jobName)
	}
	spec := job.CronSpec()
	if spec == "" {
		return errors.Errorf("%s job has no cron spec", jobName)
	}
	entryID, err := t.c.AddJob(spec, job)
	if err != nil {
		return errors.WithMessagef(err, "%s job cron spec not valid", jobName)
	}

	t.jobs[jobName] = job
	t.cronIDs[jobName] = entryID
	t.log.Infof("%s job added with cron '%s'", jobName, spec)
	return nil
}

// RemoveJob ...
func (t *Scheduler) RemoveJob(jobName string) error {
	t.cronMutex.Lock()
	defer t.cronMutex.Unlock()
	entryID, ok := t.cronIDs[jobName]
	if !ok {
		return errors.Errorf("%s job not exist", jobName)
	}

	t.c.Remove(entryID)
	delete(t.jobs, jobName)
	delete(t.cronIDs, jobName)
	t.log.Infof("%s job removed", jobName)
	return nil
}

func (t *Scheduler) Start() {
	t.c.Start()
	t.log.Infof("Job scheduler started")
}

func (t *Scheduler) Stop() {
	t.c.Stop()
	t.log.Infof("Job scheduler stopped")
}

func (t *Scheduler) GetJobs() map[string]jobHandler {
	t.cronMutex.Lock()
	defer t.cronMutex.Unlock()
	return t.jobs
}
