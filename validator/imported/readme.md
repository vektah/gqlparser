# graphql.js spec importer

These specs have been generated from the testsuite in [graphql/graph-js](https://github.com/graphql/graphql-js).

Direct modifications should not be made to most of this directory, instead take a look at the exporter.

```shell script
# update to latest
$ rm graphql-js-commit.log && ./export.sh

# re-generate with known revision
$ ./export.sh
```

You will then need to manually update the [`validator/prelude.graphql`](./validator/prelude.graphql) file, [`validator/schema.go`](./validator/schema.go) file, and possibly other files relevant to those specific changes like [./validator/rules/](./validator/rules/), etc.

Please in your PR description note the git release tag that corresponds to the graphql-js commit.
