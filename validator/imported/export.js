import fs from 'fs';
import Module from 'module';
import { testSchema } from './graphql-js/src/validation/__tests__/harness';
import { printSchema } from './graphql-js/src/utilities';
import { safeDump } from 'js-yaml';

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

function resultProxy(start) {
    let handler = {
        get: function(obj, prop) {
            if (typeof prop === 'symbol') {
                console.log("RET");
                return obj
            }
            return new Proxy({path: obj.path + "." + prop}, handler)
        },
    };

    return new Proxy({path:start}, handler);
}

// replace empty lines with the normal amount of whitespace
// so that yaml correctly preserves the whitespace
function normalizeWs(rawString) {
    const lines = rawString.split(/\r\n|[\n\r]/g);

    let commonIndent = 1000000;
    for (let i = 1; i < lines.length; i++) {
        const line = lines[i];
        if (line.trim() === '') continue;

        const indent = line.search(/\S/);
        if (indent < commonIndent) {
            commonIndent = indent;
        }
    }

    for (let i = 0; i < lines.length; i++) {
        if (lines[i].length < commonIndent) {
            lines[i] = ' '.repeat(commonIndent);
        }
    }
    return lines.join('\n');
}

const harness = {
    expectPassesRule(rule, queryString) {
        harness.expectPassesRuleWithSchema(testSchema, rule, queryString);
    },
    expectPassesRuleWithSchema(schema, rule, queryString, errors) {
        tests.push({
            name: names.slice(1).join('/'),
            rule: rule.name,
            schema: registerSchema(schema),
            query: normalizeWs(queryString),
            errors: [],
        });
    },
    expectFailsRule(rule, queryString, errors) {
        harness.expectFailsRuleWithSchema(testSchema, rule, queryString, errors);

        return resultProxy("errors")
    },
    expectFailsRuleWithSchema(schema, rule, queryString, errors) {
        tests.push({
            name: names.slice(1).join('/'),
            rule: rule.name,
            schema: registerSchema(schema),
            query: normalizeWs(queryString),
            errors: errors,
        });
    }
};


let tests = [];
let names = [];
const fakeModules = {
    'mocha': {
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
    'chai': {
        expect(it) {
            const expect = {
                get to() {
                    return expect
                },
                equal(value) {
                    // currently ignored, we know all we need to add an assertion here.
                },
            };

            return expect
        },
    },
    './harness': harness,
};

const originalLoader = Module._load;
Module._load = function(request, parent, isMain) {
    return fakeModules[request] || originalLoader(request, parent, isMain);
};

fs.readdirSync("./graphql-js/src/validation/__tests__").forEach(file => {
    if (!file.endsWith('-test.js')) {
        return
    }

    if (file === 'validation-test.js') {
        return
    }

    require('./graphql-js/src/validation/__tests__/' + file);

    let dump = safeDump(tests, {
        skipInvalid: true,
        flowLevel: 5,
        noRefs: true,
        lineWidth: 1000,
    });
    fs.writeFileSync("./spec/"+file.replace('-test.js', '.spec.yml'), dump);

    tests = [];
});

let schemaList = schemas.map(s => printSchema(s));


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

let dump = safeDump(schemaList, {
    skipInvalid: true,
    flowLevel: 5,
    noRefs: true,
    lineWidth: 1000,
});
fs.writeFileSync("./spec/schemas.yml", dump);
