# git-sync

[![Build Status](https://travis-ci.org/takecy/git-sync.svg?branch=master)](https://travis-ci.org/takecy/git-sync)
[![Go Report Card](https://goreportcard.com/badge/github.com/takecy/git-sync)](https://goreportcard.com/report/github.com/takecy/git-sync)

![](https://img.shields.io/badge/golang-1.10.2-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-sync)
![](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)

git-sync is Run a command to all git repositories in the current directory.

<br/>

### Usage
##### via Go
```shell
$ go get github.com/takecy/git-sync/gs
```
##### via Binary  
[Download](https://github.com/takecy/git-sync/releases) for your environment.  
and copy binary to your `$PATH`.

##### Print usage
```
$ gs

Usage:
  gs [original_options] <git_command> [git_options]

Original Options:
  --target   Specific target directory with regex.
  --ignore   Specific ignore directory with regex.
  --timeout  Specific timeout of performed commnad during on one directory.
             5s, 10m...

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
$ gs fetch --all -p
```
```shell
$ gs --target ^cool-tool pull
```
```shell
$ gs --target ^cool-tool --ignore ^wip-command pull
```

<br/>

### Development
```
$ git clone git@github.com:takecy/git-sync.git
$ cd git-sync
$ DEBUG=* go run gs/main.go version
```

<br/>

## License
[MIT](./LICENSE)
