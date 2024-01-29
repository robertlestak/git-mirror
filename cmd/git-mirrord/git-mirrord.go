package main

import (
	"flag"
	"os"
	"time"

	"git.shdw.tech/rob/git-mirror/internal/config"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func syncWorker(jobs chan *config.SyncConfig, errors chan error) {
	for {
		s := <-jobs
		if s == nil {
			errors <- nil
			continue
		}
		if err := s.Sync.Run(); err != nil {
			s.LastRun = time.Now()
			errors <- err
			s.Sync.Cleanup()
			continue
		}
		s.LastRun = time.Now()
		s.Sync.Cleanup()
		errors <- nil
	}
}

func main() {
	l := log.WithFields(log.Fields{
		"app": "git-mirrord",
	})
	l.Debug("start")
	logLevel := flag.String("log-level", log.GetLevel().String(), "log level")
	configFile := flag.String("config", "config.yaml", "config file")
	workers := flag.Int("workers", 10, "number of workers")
	flag.Parse()
	ll, err := log.ParseLevel(*logLevel)
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
	l = log.WithFields(log.Fields{
		"app":  "git-mirrord",
		"conf": *configFile,
	})
	l.Debug("start")
	if err := config.LoadFile(*configFile); err != nil {
		l.WithError(err).Fatal("failed to load config")
	}
	if len(config.C.Syncs) == 0 {
		l.Fatal("no syncs configured")
	}
	workerCount := *workers
	if config.C.Workers > 0 {
		workerCount = config.C.Workers
	}
	if len(config.C.Syncs) < workerCount {
		workerCount = len(config.C.Syncs)
	}
	jobs := make(chan *config.SyncConfig, len(config.C.Syncs))
	errors := make(chan error, len(config.C.Syncs))
	for {
		l.Debug("main loop")
		for i := 0; i < workerCount; i++ {
			go syncWorker(jobs, errors)
		}
		var pendingJobCount int
		for _, s := range config.C.Syncs {
			sched := s.Schedule
			if config.C.WorkDir != "" {
				s.Sync.WorkDir = config.C.WorkDir
			}
			if s.LastRun.Add(time.Duration(sched)).After(time.Now()) {
				continue
			}
			pendingJobCount++
			jobs <- s
		}
		for i := 0; i < pendingJobCount; i++ {
			err := <-errors
			if err != nil {
				l.WithError(err).Error("sync failed")
			}
		}
		l.Debug("sleeping")
		time.Sleep(10 * time.Second)
	}
}
