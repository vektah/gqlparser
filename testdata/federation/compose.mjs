// Composes the Federation v1 subgraphs into supergraph.graphql and records
// the fetches the query planner sends each subgraph for operations.graphql
// in fetches.json.
import {readFileSync, writeFileSync} from 'node:fs';
import {parse} from 'graphql';
import {composeServices} from '@apollo/composition';
import {Supergraph, operationFromDocument} from '@apollo/federation-internals';
import {QueryPlanner} from '@apollo/query-planner';

const subgraphs = ['accounts', 'content', 'progress'];
const composed = composeServices(subgraphs.map(name => ({
  name,
  url: `http://${name}.invalid/query`,
  typeDefs: parse(readFileSync(`subgraphs/${name}.graphql`, 'utf8')),
})));
if (composed.errors) {
  console.error(composed.errors.map(String).join('\n'));
  process.exit(1);
}
writeFileSync('supergraph.graphql', composed.supergraphSdl);

const supergraph = Supergraph.build(composed.supergraphSdl);
const planner = new QueryPlanner(supergraph);
const operations = parse(readFileSync('operations.graphql', 'utf8'));
const fetches = [];
const collect = node => {
  if (!node) return;
  if (node.kind === 'Fetch') {
    fetches.push({subgraph: node.serviceName, query: node.operation});
  }
  [...(node.nodes ?? []), node.node, node.ifClause, node.elseClause].forEach(collect);
};
for (const def of operations.definitions) {
  if (def.kind !== 'OperationDefinition') continue;
  const operation = operationFromDocument(
    supergraph.apiSchema(), operations, {operationName: def.name.value});
  collect(planner.buildQueryPlan(operation).node);
}
writeFileSync('fetches.json', JSON.stringify(fetches, null, 2) + '\n');
