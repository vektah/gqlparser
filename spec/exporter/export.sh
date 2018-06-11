#!/bin/bash

REPO_DIR=./graphql-js
PATH=./node_modules/.bin:$PATH
EXPORTER_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $EXPORTER_ROOT

if [[ -d "$REPO_DIR" ]] ; then
    echo "fetching latest graphql-js"
    cd $REPO_DIR
    git fetch origin master
    git checkout origin/master
    git reset --hard
else
    echo "cloning graphql-js"
    git clone --no-tags --single-branch -- https://github.com/graphql/graphql-js $REPO_DIR
    cd $REPO_DIR
    git checkout origin/master
fi

cd $EXPORTER_ROOT

echo "installing js dependencies"
npm install

echo "exporting tests"
babel-node ./export.js