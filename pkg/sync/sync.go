package sync

import (
	"errors"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type GitRepo struct {
	Url           string `yaml:"url"`
	ParentWorkDir string `yaml:"-"`
	Workdir       string `yaml:"-"`
}

type Sync struct {
	WorkDir       string     `yaml:"-"`
	Force         bool       `yaml:"force"`
	SourceOfTruth int        `yaml:"source"`
	Repos         []*GitRepo `yaml:"repos"`
}

func (r *GitRepo) Clone() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"fn":  "Clone",
		"url": r.Url,
		"dir": r.ParentWorkDir,
	})
	l.Debug("start")
	tmpDir, err := os.MkdirTemp(r.ParentWorkDir, "git-mirror")
	if err != nil {
		l.WithError(err).Error("failed to create temp dir")
		return err
	}
	r.Workdir = tmpDir
	l.Debugf("cloning into %s", tmpDir)
	cmd := exec.Command("git", "clone", "--mirror", r.Url, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		l.WithError(err).Error("failed to clone repo")
		return err
	}
	// pull all remote branches, including those not in the local repo
	l.Debug("pulling all branches")
	cmd = exec.Command("git", "fetch", "--all")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		l.WithError(err).Error("failed to fetch all branches")
		return err
	}
	l.Debug("done")
	return nil
}

func (r *GitRepo) SyncTo(dest *GitRepo) error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"src": r.Url,
		"dst": dest.Url,
		"fn":  "SyncTo",
	})
	l.Debug("start")
	pushErrors := make(chan error, 2)
	// push all remote branches to the destination
	// do not use mirror since we don't want to force push anything
	go func() {
		l.Debug("pushing branches")
		cmd := exec.Command("git", "push", "--all", dest.Url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = r.Workdir
		err := cmd.Run()
		if err != nil {
			pushErrors <- err
		}
		pushErrors <- nil
	}()
	go func() {
		// push all tags to the destination
		l.Debug("pushing tags")
		cmd := exec.Command("git", "push", "--tags", dest.Url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = r.Workdir
		err := cmd.Run()
		if err != nil {
			pushErrors <- err
		}
		pushErrors <- nil
	}()
	for i := 0; i < 2; i++ {
		err := <-pushErrors
		if err != nil {
			return err
		}
	}
	l.Debug("done")
	return nil
}

func (r *GitRepo) MirrorTo(dest *GitRepo) error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"src": r.Url,
		"dst": dest.Url,
		"fn":  "MirrorTo",
	})
	l.Debug("start")
	cmd := exec.Command("git", "push", "--mirror", "--force", dest.Url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.Workdir
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (r *GitRepo) cleanup() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"fn":  "cleanup",
	})
	l.Debug("start")
	return os.RemoveAll(r.Workdir)
}

func (s *Sync) syncMultidirectional() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"fn":  "syncMultidirectional",
	})
	l.Debug("start")
	for _, repo := range s.Repos {
		for _, dest := range s.Repos {
			if err := repo.SyncTo(dest); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sync) syncFromTo() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
	})
	l.Debug("start")
	sourceRepo := s.Repos[s.SourceOfTruth]
	for i, repo := range s.Repos {
		if i == s.SourceOfTruth {
			continue
		}
		err := sourceRepo.SyncTo(repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sync) mirrorFromTo() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
	})
	l.Debug("start")
	sourceRepo := s.Repos[s.SourceOfTruth]
	for i, repo := range s.Repos {
		if i == s.SourceOfTruth {
			continue
		}
		err := sourceRepo.MirrorTo(repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func cloneWorker(jobs <-chan *GitRepo, results chan<- error) {
	for repo := range jobs {
		results <- repo.Clone()
	}
}

func (s *Sync) Run() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
		"fn":  "Run",
	})
	l.Debug("start")
	cloneErrs := make(chan error, len(s.Repos))
	if s.WorkDir != "" {
		err := os.MkdirAll(s.WorkDir, 0755)
		if err != nil {
			return err
		}
	}
	for _, repo := range s.Repos {
		repo.ParentWorkDir = s.WorkDir
	}
	jobs := make(chan *GitRepo, len(s.Repos))
	workerCount := 10
	if len(s.Repos) < workerCount {
		workerCount = len(s.Repos)
	}
	for w := 1; w <= workerCount; w++ {
		go cloneWorker(jobs, cloneErrs)
	}
	for _, repo := range s.Repos {
		jobs <- repo
	}
	close(jobs)
	for i := 0; i < len(s.Repos); i++ {
		err := <-cloneErrs
		if err != nil {
			return err
		}
	}
	if s.SourceOfTruth == -1 && s.Force {
		return errors.New("cannot use force with multidirectional sync")
	}
	if s.SourceOfTruth == -1 {
		l.Info("multidirectional sync")
		return s.syncMultidirectional()
	}
	if s.Force {
		l.Info("force sync")
		return s.mirrorFromTo()
	}
	l.Info("sync")
	return s.syncFromTo()
}

func cleaupWorker(jobs <-chan *GitRepo, results chan<- error) {
	for j := range jobs {
		results <- j.cleanup()
	}
}

func (s *Sync) Cleanup() error {
	l := log.WithFields(log.Fields{
		"app": "git-mirror",
	})
	l.Debug("start")
	cleanupErrs := make(chan error, len(s.Repos))
	jobs := make(chan *GitRepo, len(s.Repos))
	workerCount := 10
	if len(s.Repos) < workerCount {
		workerCount = len(s.Repos)
	}
	for w := 1; w <= workerCount; w++ {
		go cleaupWorker(jobs, cleanupErrs)
	}
	for _, repo := range s.Repos {
		jobs <- repo
	}
	close(jobs)
	for i := 0; i < len(s.Repos); i++ {
		err := <-cleanupErrs
		if err != nil {
			return err
		}
	}
	return nil
}
