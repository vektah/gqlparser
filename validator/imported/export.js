import fs from "fs";
import Module from "module";
import { testSchema } from "./graphql-js/src/validation/__tests__/harness";
import {
  GraphQLSchema,
  printIntrospectionSchema,
  printSchema
} from "./graphql-js/src";
import yaml from "js-yaml";

main()

function main() {
  exportSpecDefinitions()
  exportPrelude()
}

function exportPrelude() {
  //
  const generatedText = "# This file defines all the implicitly declared types that are required by the graphql spec. It is implicitly included by calls to LoadSchema\n"
  const builtinScalars = `
"The \`Int\` scalar type represents non-fractional signed whole numeric values. Int can represent values between -(2^31) and 2^31 - 1."
scalar Int

"The \`Float\` scalar type represents signed double-precision fractional values as specified by [IEEE 754](http://en.wikipedia.org/wiki/IEEE_floating_point)."
scalar Float

"The \`String\`scalar type represents textual data, represented as UTF-8 character sequences. The String type is most often used by GraphQL to represent free-form human-readable text."
scalar String

"The \`Boolean\` scalar type represents \`true\` or \`false\`."
scalar Boolean

"""The \`ID\` scalar type represents a unique identifier, often used to refetch an object or as key for a cache. The ID type appears in a JSON response as a String; however, it is not intended to be human-readable. When expected as an input type, any string (such as "4") or integer (such as 4) input value will be accepted as an ID."""
scalar ID
`

  const deferDirective = `
"Directs the executor to defer this fragment when the \`if\` argument is true or undefined."
directive @defer(
  "Deferred when true or undefined."
  if: Boolean = true,
  "Unique name"
  label: String
) on FRAGMENT_SPREAD | INLINE_FRAGMENT
\n`

  const customGeneratedContent = generatedText + builtinScalars + deferDirective

  const schema = new GraphQLSchema({});
  const output = customGeneratedContent + printIntrospectionSchema(schema);

  fs.writeFileSync("./prelude.graphql", output);
}

function exportSpecDefinitions() {
  let schemas = [];
  function registerSchema(schema) {
    for (let i = 0; i < schemas.length; i++) {
      if (schemas[i] === schema) {
        return i;
      }
    }
    schemas.push(schema);
    return schemas.length - 1;
  }

  function resultProxy(start, base = {}) {
    const funcWithPath = (path) => {
      const f = () => {};
      f.path = path;
      return f;
    };
    let handler = {
      get: function (obj, prop) {
        if (base[prop]) {
          return base[prop];
        }
        return new Proxy(funcWithPath(`${obj.path}.${prop}`), handler);
      },
    };

    return new Proxy(funcWithPath(start), handler);
  }

// replace empty lines with the normal amount of whitespace
// so that yaml correctly preserves the whitespace
  function normalizeWs(rawString) {
    const lines = rawString.split(/\r\n|[\n\r]/g);

    let commonIndent = 1000000;
    for (let i = 1; i < lines.length; i++) {
      const line = lines[i];
      if (!line.trim()) {
        continue;
      }

      const indent = line.search(/\S/);
      if (indent < commonIndent) {
        commonIndent = indent;
      }
    }

    for (let i = 1; i < lines.length; i++) {
      if (lines[i].length < commonIndent) {
        lines[i] = " ".repeat(commonIndent);
      }
    }
    return lines.join("\n");
  }

  const harness = {
    testSchema,

    expectValidationErrorsWithSchema(schema, rule, queryStr) {
      return resultProxy("expectValidationErrorsWithSchema", {
        toDeepEqual(expected) {
          tests.push({
            name: names.slice(1).join("/"),
            rule: rule.name.replace(/Rule$/, ""),
            schema: registerSchema(schema),
            query: normalizeWs(queryStr),
            errors: expected,
          });
        },
      });
    },
    expectValidationErrors(rule, queryStr) {
      return harness.expectValidationErrorsWithSchema(testSchema, rule, queryStr);
    },
    expectSDLValidationErrors(schema, rule, sdlStr) {
      return resultProxy("expectSDLValidationErrors", {
        toDeepEqual(expected) {
          // ignore now...
          // console.warn(rule.name, sdlStr, JSON.stringify(expected, null, 2));
        },
      });
    },
  };

  let tests = [];
  let names = [];
  const fakeModules = {
    mocha: {
      describe(name, f) {
        names.push(name);
        f();
        names.pop();
      },
      it(name, f) {
        names.push(name);
        f();
        names.pop();
      },
    },
    chai: {
      expect(it) {
        const expect = {
          get to() {
            return expect;
          },
          get have() {
            return expect;
          },
          get nested() {
            return expect;
          },
          equal(value) {
            // currently ignored, we know all we need to add an assertion here.
          },
          property(path, value) {
            // currently ignored, we know all we need to add an assertion here.
          },
        };

        return expect;
      },
    },
    "./harness": harness,
  };

  const originalLoader = Module._load;
  Module._load = function (request, parent, isMain) {
    return fakeModules[request] || originalLoader(request, parent, isMain);
  };

  fs.readdirSync("./graphql-js/src/validation/__tests__").forEach((file) => {
    if (!file.endsWith("-test.ts")) {
      return;
    }

    if (file === "validation-test.ts") {
      return;
    }

    require(`./graphql-js/src/validation/__tests__/${file}`);

    let dump = yaml.dump(tests, {
      skipInvalid: true,
      flowLevel: 5,
      noRefs: true,
      lineWidth: 1000,
    });
    fs.writeFileSync(`./spec/${file.replace("-test.ts", ".spec.yml")}`, dump);

    tests = [];
  });

  let schemaList = schemas.map((s) => printSchema(s));

  schemaList[0] += `
# injected becuase upstream spec is missing some types
extend type QueryRoot {
    field: T
    f1: Type
    f2: Type
    f3: Type
}

type Type {
    a: String
    b: String
    c: String
}
type T {
    a: String
    b: String
    c: String
    d: String
    y: String
    deepField: T
    deeperField: T
}`;

  // Required for TestValidation/KnownRootTypeRule/Valid_root_type_but_schema_is_entirely_empty test
  schemaList.push("")

  let dump = yaml.dump(schemaList, {
    skipInvalid: true,
    flowLevel: 5,
    noRefs: true,
    lineWidth: 1000,
  });
  fs.writeFileSync("./spec/schemas.yml", dump);
}
