package validator_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
	"github.com/vektah/gqlparser/validator"
	"gopkg.in/yaml.v2"
)

type Spec struct {
	Name   string
	Rule   string
	Schema int
	Query  string
	Errors gqlerror.List
}

type Deviation struct {
	Rule   string
	Errors []*gqlerror.Error
	Skip   string

	pattern *regexp.Regexp
}

func TestValidation(t *testing.T) {
	var rawSchemas []string
	readYaml("./imported/spec/schemas.yml", &rawSchemas)

	var deviations []*Deviation
	readYaml("./imported/deviations.yml", &deviations)
	for _, d := range deviations {
		d.pattern = regexp.MustCompile("^" + d.Rule + "$")
	}

	var schemas []*validator.Schema
	for i, schema := range rawSchemas {
		schema, err := gqlparser.LoadSchema(&ast.Source{Input: schema, Name: fmt.Sprintf("schemas.yml[%d]", i)})
		if err != nil {
			panic(err)
		}
		schemas = append(schemas, schema)
	}

	err := filepath.Walk("./imported/spec/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".spec.yml") {
			return nil
		}

		runSpec(t, schemas, deviations, path)
		return nil
	})
	require.NoError(t, err)
}

func runSpec(t *testing.T, schemas []*validator.Schema, deviations []*Deviation, filename string) {
	ruleName := strings.TrimSuffix(filepath.Base(filename), ".spec.yml")

	var specs []Spec
	readYaml(filename, &specs)
	t.Run(ruleName, func(t *testing.T) {
		for _, spec := range specs {
			if len(spec.Errors) == 0 {
				spec.Errors = nil
			}
			t.Run(spec.Name, func(t *testing.T) {
				for _, deviation := range deviations {
					if deviation.pattern.MatchString(ruleName + "/" + spec.Name) {
						if deviation.Skip != "" {
							t.Skip(deviation.Skip)
						}
						if deviation.Errors != nil {
							spec.Errors = deviation.Errors
						}
					}
				}

				_, err := gqlparser.LoadQuery(schemas[spec.Schema], spec.Query)
				var finalErrors gqlerror.List
				for _, err := range err {
					// ignore errors from other rules
					if err.Rule != spec.Rule {
						continue
					}
					finalErrors = append(finalErrors, err)
				}

				for i := range spec.Errors {
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

				if len(finalErrors) != len(spec.Errors) {
					t.Errorf("wrong number of errors returned\ngot:\n%s\nwant:\n%s", finalErrors.Error(), spec.Errors)
				} else {
					for i := range spec.Errors {
						expected := spec.Errors[i]
						actual := finalErrors[i]
						if actual.Rule != spec.Rule {
							continue
						}
						var errLocs []string
						if expected.Message != actual.Message {
							errLocs = append(errLocs, "message mismatch")
						}
						if len(expected.Locations) > 0 && len(actual.Locations) == 0 {
							errLocs = append(errLocs, "missing location")
						}
						if len(expected.Locations) > 0 && len(actual.Locations) > 0 {
							found := false
							for _, loc := range expected.Locations {
								if actual.Locations[0].Line == loc.Line {
									found = true
									break
								}
							}

							if !found {
								errLocs = append(errLocs, "line")
							}
						}

						if len(errLocs) > 0 {
							t.Errorf("%s\ngot:  %s\nwant: %s", strings.Join(errLocs, ", "), finalErrors[i].Error(), spec.Errors[i].Error())
						}
					}
				}

				if t.Failed() {
					t.Logf("name: '%s'", spec.Name)
					t.Log("\nquery:", spec.Query)
				}
			})
		}
	})
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
