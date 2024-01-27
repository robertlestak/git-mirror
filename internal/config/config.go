package config

import (
	"os"
	"time"

	"git.shdw.tech/rob/git-mirror/pkg/sync"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	C *Config
)

type SyncConfig struct {
	Name     string        `yaml:"name"`
	Schedule time.Duration `yaml:"schedule"`
	Sync     sync.Sync     `yaml:"sync"`
	LastRun  time.Time     `yaml:"-"`
}

type Config struct {
	Workers int           `yaml:"workers"`
	WorkDir string        `yaml:"workdir"`
	Syncs   []*SyncConfig `yaml:"sync"`
}

func LoadFile(f string) error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"fn":  "LoadFile",
	})
	l.Debug("start")
	c := &Config{}
	fd, err := os.Open(f)
	if err != nil {
		l.WithError(err).Error("failed to open config file")
		return err
	}
	defer fd.Close()
	dec := yaml.NewDecoder(fd)
	err = dec.Decode(c)
	if err != nil {
		l.WithError(err).Error("failed to decode config file")
		return err
	}
	C = c
	for _, s := range C.Syncs {
		l.Debugf("sync: %+v", s)
	}
	return nil
}
