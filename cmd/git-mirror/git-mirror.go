package main

import (
	"flag"
	"os"

	"git.shdw.tech/rob/git-mirror/pkg/sync"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

type repoFlags []string

func (i *repoFlags) String() string {
	return "my string representation"
}

func (i *repoFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
	})
	l.Debug("start")
	logLevel := flag.String("log-level", log.GetLevel().String(), "log level")
	var rFlags repoFlags
	flag.Var(&rFlags, "repo", "repos to sync")
	workDir := flag.String("dir", "", "work dir")
	force := flag.Bool("force", false, "force push")
	sourceOfTruth := flag.Int("source", -1, "source of truth")
	flag.Parse()
	ll, err := log.ParseLevel(*logLevel)
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
	l = log.WithFields(log.Fields{
		"app":    "git-mirror",
		"dir":    *workDir,
		"force":  *force,
		"source": *sourceOfTruth,
	})
	l.Debug("start")
	if *sourceOfTruth == -1 && *force {
		l.Fatal("cannot force push without a source of truth")
	}
	s := sync.Sync{
		WorkDir:       *workDir,
		Force:         *force,
		SourceOfTruth: *sourceOfTruth,
	}
	for _, r := range rFlags {
		s.Repos = append(s.Repos, &sync.GitRepo{
			Url: r,
		})
	}
	defer s.Cleanup()
	if err := s.Run(); err != nil {
		l.WithError(err).Error("error running sync")
		s.Cleanup()
		os.Exit(1)
	}
	l.Debug("end")
}
