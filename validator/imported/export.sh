#!/bin/bash

REPO_DIR=./graphql-js
PATH=./node_modules/.bin:$PATH
EXPORTER_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $EXPORTER_ROOT

GIT_REF=origin/master

if [[ -f "$EXPORTER_ROOT/graphql-js-commit.log" ]] ; then
  GIT_REF=$(cat "$EXPORTER_ROOT/graphql-js-commit.log")
fi
echo $GIT_REF

if [[ -d "$REPO_DIR" ]] ; then
    echo "fetching graphql-js with ${GIT_REF}"
    cd $REPO_DIR
    git fetch origin master
    git checkout "$GIT_REF"
    git reset --hard
else
    echo "cloning graphql-js with ${GIT_REF}"
    git clone --no-tags --single-branch -- https://github.com/graphql/graphql-js $REPO_DIR
    cd $REPO_DIR
    git checkout "$GIT_REF"
fi
git rev-parse HEAD > $EXPORTER_ROOT/graphql-js-commit.log

cd $EXPORTER_ROOT

echo "installing js dependencies"
npm install

echo "exporting tests"
babel-node ./export.js
