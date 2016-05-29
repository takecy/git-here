# git-sync

[![Build Status](https://drone.io/github.com/takecy/git-sync/status.png)](https://drone.io/github.com/takecy/git-sync/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/takecy/git-sync)](https://goreportcard.com/report/github.com/takecy/git-sync)

![](https://img.shields.io/badge/golang-1.6.2-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-sync)
![](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)

git-sync is Run a command to all git repositories in the current directory.

<br/>
### Usage
##### via Go
```shell
$ go get github.com/takecy/git-sync
```
##### via Binary  
[Download](https://github.com/takecy/git-sync/releases) for your environment.  
and copy binary to your `$PATH`.

##### Print help
```
$ git-sync
Usage:
  git-sync [original_options] <git_command> [git_options]

Original Options:
  --target  Specific target directory with regex.
  --ignore  Specific ignore directory with regex.

Commands:
  version     Print version.
  <command>   Same as git command. (fetch, pull, status...)

Options:
  Same as git.
```
##### Default target directories
```shell
$ tree
.
├── .Hoge        // ignore (start from comma)
├── repo_a       // target
├── dir
│   └── repo_b   // ignore
└── repo_c       // target
```

<br/>
#### Examples
```shell
$ git-sync fetch --all -p
```
```shell
$ git-sync --target ^cool-tool pull
```
```shell
$ git-sync --target ^cool-tool --ignore ^wip-command pull
```

<br/>
## License
[MIT](./LICENSE)
