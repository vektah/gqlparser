#!/bin/bash

REPO_DIR=./graphql-js
PATH=$PATH:./node_modules/.bin

cd "$( dirname "${BASH_SOURCE[0]}" )"

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

echo "installing js dependencies"
npm install

echo "exporting tests"
cp export.js graphql-js
babel-node graphql-js/export.js