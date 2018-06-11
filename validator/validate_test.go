package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
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
