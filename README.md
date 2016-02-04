# git-sync

![](https://img.shields.io/badge/golang-1.5.3-blue.svg?style=flat-square)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/takecy/git-sync)
![](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)


git-sync is `git fetch` or `git pull` alias.  
Run a command to all git repositories in the current directory.

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

Commands:
  version  Print version.
  fetch    Alias for <git fetch>.
  pull     Alias for <git pull>.

Options:
  Same as git.

Original Options:
  --target-dir  Specific target directory with regex.
  --ignore-dir  Specific ignore directory with regex.
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
$ git-sync --target-dir ^cool-tool pull
```
```shell
$ git-sync --target-dir ^cool-tool --ignore-dir ^wip-command pull
```

<br/>
## License
[MIT](./LICENSE)
