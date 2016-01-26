# git-sync

![](https://img.shields.io/badge/golang-1.5.2+-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-sync)


git-sync is `git fetch` or `git pull` alias.  
Run a command to the git repositories in the current directory.

<br/>
### Usage
##### via Go
```shell
$ go get github.com/takecy/git-sync
```
##### via Binary  
[Download](https://github.com/takecy/git-sync/releases)  
and copy binary to your `$PATH`.

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
  Same as git.
```

<br/>
#### Example
```shell
$ tree
.
├── .Hoge        // ignore
├── repo_a       // target
├── dir
│   └── repo_b   // ignore
└── repo_c       // target
```

It will be performed on the repo_a, and repo_c.  
```shell
$ git-sync fetch --all -p
```
```shell
$ git-sync pull
```

<br/>
## License
MIT
