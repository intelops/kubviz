package scheduler

import (
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/intelops/go-common/logging"
	"github.com/intelops/kubviz/agent/config"
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

func InitScheduler(config *rest.Config, js nats.JetStreamContext, cfg config.AgentConfigurations, clientset *kubernetes.Clientset) (s *Scheduler) {
	log := logging.NewLogger()
	s = NewScheduler(log)
	if cfg.OutdatedInterval != "" && cfg.OutdatedInterval != "0" {
		sj, err := NewOutDatedImagesJob(config, js, cfg.OutdatedInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Outdated", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.GetAllInterval != "" && cfg.GetAllInterval != "0" {
		sj, err := NewKetallJob(config, js, cfg.GetAllInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("GetALL", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.KubeScoreInterval != "" && cfg.KubeScoreInterval != "0" {
		sj, err := NewKubescoreJob(clientset, js, cfg.KubeScoreInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("KubeScore", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.RakkessInterval != "" && cfg.RakkessInterval != "0" {
		sj, err := NewRakkessJob(config, js, cfg.RakkessInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Rakkess", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.KubePreUpgradeInterval != "" && cfg.KubePreUpgradeInterval != "0" {
		sj, err := NewKubePreUpgradeJob(config, js, cfg.KubePreUpgradeInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("KubePreUpgrade", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.TrivyImageInterval != "" && cfg.TrivyImageInterval != "0" {
		sj, err := NewTrivyImagesJob(config, js, cfg.TrivyImageInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Trivyimage", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.TrivySbomInterval != "" && cfg.TrivySbomInterval != "0" {
		sj, err := NewTrivySbomJob(config, js, cfg.TrivySbomInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Trivysbom", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.TrivyClusterScanInterval != "" && cfg.TrivyClusterScanInterval != "0" {
		sj, err := NewTrivyClusterScanJob(js, cfg.TrivyClusterScanInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Trivycluster", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	return
}
