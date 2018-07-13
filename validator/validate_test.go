package validator

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
	"testing"

	"os"
	"path/filepath"

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
}

type Deviation struct {
	Rule   string
	Errors []errors.Validation
	Skip   string

	pattern *regexp.Regexp
}

func TestSpec(t *testing.T) {
	var rawSchemas []string
	readYaml("../spec/validation/schemas.yml", &rawSchemas)

	var deviations []*Deviation
	readYaml("../spec/validation/deviations.yml", &deviations)
	for _, d := range deviations {
		d.pattern = regexp.MustCompile("^" + d.Rule + "$")
	}

	var schemas []*gqlparser.Schema
	for _, schema := range rawSchemas {
		schema, err := gqlparser.LoadSchema(schema)
		if err != nil {
			panic(err)
		}
		schemas = append(schemas, schema)
	}

	err := filepath.Walk("../spec/validation/", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path)
		if info.IsDir() || !strings.HasSuffix(path, ".spec.yml") {
			return nil
		}

		ruleName := strings.TrimSuffix(filepath.Base(path), ".spec.yml")

		t.Run(ruleName, runSpec(schemas, deviations, path))
		return nil
	})
	require.NoError(t, err)
}

func runSpec(schemas []*gqlparser.Schema, deviations []*Deviation, filename string) func(t *testing.T) {
	var specs []Spec
	readYaml(filename, &specs)
	return func(t *testing.T) {
		for _, spec := range specs {
			if len(spec.Errors) == 0 {
				spec.Errors = nil
			}
			t.Run(spec.Name, func(t *testing.T) {
				for _, deviation := range deviations {
					if deviation.pattern.MatchString(spec.Name) {
						if deviation.Skip != "" {
							t.Skip(deviation.Skip)
						}
						if deviation.Errors != nil {
							spec.Errors = deviation.Errors
						}
					}
				}

				t.Logf("name: '%s'", spec.Name)

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

					// remove inconsistent use of ;
					spec.Errors[i].Message = strings.Replace(spec.Errors[i].Message, "; Did you mean", ". Did you mean", -1)
				}
				sort.Slice(spec.Errors, func(i, j int) bool {
					return strings.Compare(spec.Errors[i].Message, spec.Errors[j].Message) > 0
				})
				sort.Slice(finalErrors, func(i, j int) bool {
					return strings.Compare(finalErrors[i].Message, finalErrors[j].Message) > 0
				})
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
