# git-sync

![](https://img.shields.io/badge/golang-1.5.2-blue.svg?style=flat)
[![GoDoc](https://godoc.org/github.com/takecy/git-sync?status.svg)](https://godoc.org/github.com/takecy/git-sync)

git-sync is `git fetch` or `git pull` alias.  
Run a command to the git repository all in the current directory.

<br/>
### Usage
##### via Go
```shell
$ go get github.com/takecy/git-sync
```
##### via Binary  
[Download](https://github.com/takecy/git-sync/releases)  
and copy to your `$PATH`.

<br/>
Print help.
```
$ git-sync
Usage:
  git-sync <command> [options]

Commands:
  fetch   Alias for <git fetch>.
  pull    Alias for <git pull>.

Options:
  same of git.
```

<br/>
#### Example
```shell
$ tree
.
├── .Hoge
├── repo_a
├── repo_b
└── repo_c
```

It will be performed on the repo_a, repo_b, and repo_c.  
```shell
$ git-sync fetch --all -p
```

## License
MIT
