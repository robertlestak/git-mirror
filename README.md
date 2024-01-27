# git-mirror

a simple cli and daemon tool to mirror git repositories between multiple git servers.

## git-mirror

### Configuration

```bash
Usage of git-mirror
  -dir string
        work dir
  -force
        force push
  -log-level string
        log level (default "log")
  -repo value
        repos to sync
  -source int
        source of truth (default -1)
```

`git-mirror` accepts 2 or more git repository urls and mirrors them to each other. By default, a mutlidirectional mirror is performed (assuming there are no merge conflicts).

A source-of-truth repo can be specified with the `-source` flag, where the number provided is the index of the repo in the `-repo` list. This repo will be used as the source of truth, and all other repos will be updated to match it. By default in source mode, it will not force push, and will only update the other repos if they are behind the source repo.

If the `-force` flag is provided, it will force push to all repos, regardless of whether they are behind or not. This can only be used with the `-source` flag.

## git-mirrord

`git-mirrord` is a daemon that will periodically run `git-mirror` on a list of repos. It is configured with a yaml file, and will run `git-mirror` on each repo in the list at the specified interval.

### config.yaml

```yaml
workers: 20 # number of workers to use, defaults to 10
sync:
- schedule: "10m"
  name: "hello-world"
  sync:
    repos:
    - url: https://github.com/example/hello-world
    - url: https://git.example.com/example/hello-world
- schedule: "5m"
  name: "another"
  sync:
    force: true # force mirror
    source: 0 # use the first repo as the source
    repos:
    - url: https://github.com/example/another
    - url: https://git.example.com/example/another
    - url: https://git.example2.com/example/another2
```