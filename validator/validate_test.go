package validator

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/errors"
	yaml "gopkg.in/yaml.v2"
)

func TestValidate(t *testing.T) {
	s, err := gqlparser.LoadSchema(`
		type User {
			name: String!
		}
	`)
	require.Nil(t, err)

	req, err := gqlparser.ParseQuery(`fragment subFieldNotDefined on User { Name }`)
	require.Nil(t, err)

	errs := Validate(s, &req)

	fmt.Println(errs)
}

type Spec struct {
	Name   string
	Rule   string
	Schema int
	Query  string
	Errors []errors.Validation
}

func TestSpec(t *testing.T) {
	var rawSchemas []string
	readYaml("../spec/validation/schemas.yml", &rawSchemas)

	var schemas []*gqlparser.Schema
	for _, schema := range rawSchemas {
		schema, err := gqlparser.LoadSchema(schema)
		if err != nil {
			panic(err)
		}
		schemas = append(schemas, schema)
	}

	t.Run("FieldsOnCorrectType", runSpec(schemas, "../spec/validation/FieldsOnCorrectType.yml"))
	t.Run("FragmentsOnCompositeTypes", runSpec(schemas, "../spec/validation/FragmentsOnCompositeTypes.yml"))
}

func runSpec(schemas []*gqlparser.Schema, filename string) func(t *testing.T) {
	var specs []Spec
	readYaml(filename, &specs)
	return func(t *testing.T) {
		for _, spec := range specs {
			if len(spec.Errors) == 0 {
				spec.Errors = nil
			}
			t.Run(spec.Name, func(t *testing.T) {
				query, err := gqlparser.ParseQuery(spec.Query)
				require.Nil(t, err)
				errs := Validate(schemas[spec.Schema], &query)

				for i := range spec.Errors {
					// todo fixme. These arent currently supported.
					spec.Errors[i].Rule = spec.Rule
					spec.Errors[i].Locations = nil
				}
				require.Equal(t, spec.Errors, errs)
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
