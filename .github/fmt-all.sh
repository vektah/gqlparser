#!/bin/sh

# see https://til.simonwillison.net/yaml/yamlfmt

function is_bin_in_path {
  builtin type -P "$1" &> /dev/null
}

export GOBIN="$HOME/go/bin"
! is_bin_in_path yamlfmt && GOBIN=$HOME/go/bin go install -v go install github.com/google/yamlfmt/cmd/yamlfmt@latest

# -formatter indentless_arrays=true,retain_line_breaks=true
yamlfmt \
  -conf .yamlfmt.yaml .