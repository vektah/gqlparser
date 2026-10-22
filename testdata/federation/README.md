# Federation v1 fixtures

`federation_test.go` uses these files to check that gqlparser handles the
GraphQL that an Apollo Federation v1 deployment produces.

- `subgraphs/` holds three Federation v1 subgraph schemas, written the way a
  gqlgen subgraph declares them. `federation_v1.graphql` holds the federation
  definitions gqlgen adds to every v1 subgraph.
- `operations.graphql` holds client operations. Between them they use entities
  that span subgraphs, `@requires`, `@provides`, interfaces, unions, fragments,
  `@include`, variables, a mutation, introspection, and deprecated arguments,
  input fields, and enum values.
- `supergraph.graphql` and `fetches.json` are generated. Don't edit them.

To regenerate the generated files after changing a schema or an operation:

```sh
cd testdata/federation
npm install --no-package-lock
npm run compose
rm -r node_modules
```

`compose.mjs` composes the subgraphs with `@apollo/composition`, which turns
Federation v1 subgraphs into a `join/v0.3` supergraph. It then plans each
operation with `@apollo/query-planner` and writes every subgraph fetch from the
plans to `fetches.json`. A gateway built on these packages sends these fetches
to the subgraphs. Update `federation_test.go` to give each new fetch its
variables.
