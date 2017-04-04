#!/usr/bin/env bash

set -e

for d in $(go list ./... | grep -v vendor); do
    go test -race -v $d
done