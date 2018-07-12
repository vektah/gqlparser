package validator

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
	"gopkg.in/yaml.v2"
)

type Spec struct {
	Name   string
	Rule   string
	Schema int
	Query  string
	Errors []errors.Validation
	Skip   bool
}

func TestSpec(t *testing.T) {
	var rawSchemas []string
	readYaml("../spec/validation/schemas.yml", &rawSchemas)

	var deviations map[string]*Spec
	readYaml("../spec/validation/deviations.yml", &deviations)

	var schemas []*gqlparser.Schema
	for _, schema := range rawSchemas {
		schema, err := gqlparser.LoadSchema(schema)
		if err != nil {
			panic(err)
		}
		schemas = append(schemas, schema)
	}

	files, err := ioutil.ReadDir("../spec/validation/")
	if err != nil {
		panic(err)
	}

	excludes := []string{
		"schemas.yml",
		"deviations.yml",
	}
	ignores := []string{
		"ExecutableDefinitions",
		"NoUndefinedVariables",
		"NoUnusedVariables",
		"OverlappingFieldsCanBeMerged",
		"PossibleFragmentSpreads",
		"ProvidedRequiredArguments",
		"ValuesOfCorrectType",
		"VariablesAreInputTypes",
		"VariablesInAllowedPosition",
	}

file:
	for _, file := range files {
		fileName := file.Name()

		if !strings.HasSuffix(fileName, ".yml") {
			continue
		}

		for _, exclude := range excludes {
			if exclude == fileName {
				continue file
			}
		}

		ruleName := fileName[:len(fileName)-len(".yml")]

		for _, ignore := range ignores {
			if ignore == ruleName {
				t.Run(ruleName, func(t *testing.T) {
					t.SkipNow()
				})
				continue file
			}
		}

		t.Run(ruleName, runSpec(schemas, deviations, fmt.Sprintf("../spec/validation/%s", fileName)))
	}
}

func runSpec(schemas []*gqlparser.Schema, deviations map[string]*Spec, filename string) func(t *testing.T) {
	var specs []Spec
	readYaml(filename, &specs)
	return func(t *testing.T) {
		for _, spec := range specs {
			if len(spec.Errors) == 0 {
				spec.Errors = nil
			}
			t.Run(spec.Name, func(t *testing.T) {
				if deviation := deviations[spec.Name]; deviation != nil {
					if deviation.Errors != nil {
						spec.Errors = deviation.Errors
					}
					if deviation.Skip {
						t.SkipNow()
					}
				}

				query, err := gqlparser.ParseQuery(spec.Query)
				require.Nil(t, err)
				errs := Validate(schemas[spec.Schema], &query)

				var finalErrors []errors.Validation
				for _, err := range errs {
					// ignore errors from other rules
					if err.Rule != spec.Rule {
						continue
					}
					finalErrors = append(finalErrors, err)
				}

				// todo: location is currently not supported by the parser/
				for i := range spec.Errors {
					spec.Errors[i].Locations = nil
					spec.Errors[i].Rule = spec.Rule
				}
				assert.Equal(t, spec.Errors, finalErrors)

				if t.Failed() {
					t.Log("\nquery:", spec.Query)
				}
			})
		}
	}
}

func readYaml(filename string, result interface{}) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(b, result)
	if err != nil {
		panic(fmt.Errorf("unable to load %s: %s", filename, err.Error()))
	}
}
