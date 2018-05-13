# graphql/graphql-js text exporter


This exporter takes the [graphql/graph-js](https://github.com/graphql/graphql-js) specs and turns them into yaml files so that we can verify our implementation against their testsuite.

It is based on the great work done by @neelance for [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go).

### Regenerating the tests

Because this calls out to node and installs random dependencies I don't feel 
comfortable putting it into a go generate stanza.

You will need:
 - git
 - a recent version of node
 - npm
 - npx
 


```bash
./export.sh
```



It will clone the upstream repo, stub out the test runner and write out the specs in yaml.

The resulting files should be committed.