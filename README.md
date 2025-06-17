# zync

A tool for syncing zsh histories in git.

WARNING: This is alpha software. Back up your shell history before trying!

## Installation

```sh
go install lesiw.io/zync@latest
```

## Configuration

```sh
export ZYNCREPO=git@github.com:your/private/repo.git
export ZYNCPASS=yoursupersecretpassword
```

## Usage

Run `zync` to sync. There is probably a way of adding this to your `.zshrc` so
that it runs automatically after each command, but I haven't worked it out yet!
