# git-here

> `git-here(gh)` is Run git command to all repositories in the current directory.

<br/>

[![Build Status](https://travis-ci.org/takecy/git-here.svg?branch=master)](https://travis-ci.org/takecy/git-here)
[![Go Report Card](https://goreportcard.com/badge/github.com/takecy/git-here)](https://goreportcard.com/report/github.com/takecy/git-here)
![](https://img.shields.io/badge/golang-1.14.2-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-here)
![](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)

<br/>

## Usage
```
$ gh pull
```
```
$ gh fetch --all -p
```
```
$ gh --target ^cool-tool pull
```
```
$ gh --target ^cool-tool --ignore ^wip-command pull
```

### Default target directories
```shell
$ tree
.
├── .Hoge        // ignore (start from comma)
├── repo_a       // target
├── dir
│   └── repo_b   // ignore
└── repo_c       // target
```

## Install
### via Go
```shell
$ go get -u github.com/takecy/git-here/gh
```
### via Binary  
Download from [Release Page](https://github.com/takecy/git-here/releases) for your environment.  
and copy binary to your `$PATH`.

### Print usage
```
$ gh

Usage:
  gh [original_options] <git_command> [git_options]

Original Options:
  --target   Specific target directory with regex.
  --ignore   Specific ignore directory with regex.
  --timeout  Specific timeout of performed commnad during on one directory.
             5s, 10m...

Commands:
  version     Print version. Whether check new version exists, and ask you to upgrade to latest version.
  <command>   Same as git command. (fetch, pull, status...)

Options:
  Same as git.
```

<br/>

## Development

* Go 1.13+

#### Why this repository have vendor?
It is to simplify development. You can start right away just by cloning.

### Prepare
```
$ git clone git@github.com:takecy/git-here.git
$ cd git-here
$ DEBUG=* go run gh/main.go version
```

### Testing
```
$ make test
```

<br/>

## License
[MIT](./LICENSE)
