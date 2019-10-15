#!/bin/sh

set -e

for fi in `go list ./...`; do
  cmd="go test -count 5 -race -cover ${fi}"
  if [ `go test ${fi} --unknown | grep "adagio.integration"` ]; then
    cmd="$cmd -adagio.integration"
  fi

  eval $cmd
done

