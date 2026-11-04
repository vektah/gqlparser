package lexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockStringValue(t *testing.T) {
	t.Run("removes uniform indentation from a string", func(t *testing.T) {
		result := blockStringValue(`
    Hello,
      World!

    Yours,
      GraphQL.`)

		require.Equal(t, "Hello,\n  World!\n\nYours,\n  GraphQL.", result)
	})

	t.Run("removes empty leading and trailing lines", func(t *testing.T) {
		result := blockStringValue(`


    Hello,
      World!

    Yours,
      GraphQL.

`)

		require.Equal(t, "Hello,\n  World!\n\nYours,\n  GraphQL.", result)
	})

	t.Run("removes blank and trailing newlines", func(t *testing.T) {
		result := blockStringValue(`

       
    Hello,
      World!

    Yours,
      GraphQL.
         

`)

		require.Equal(t, "Hello,\n  World!\n\nYours,\n  GraphQL.", result)
	})

	t.Run("does not dedent the first line", func(t *testing.T) {
		// The first line starts right after the opening quotes, so its
		// indentation is part of the value and is not the common indentation.
		require.Equal(t, "  a", blockStringValue("  a"))
		require.Equal(t, " a\nb", blockStringValue(" a\n  b"))
	})

	t.Run("does not take empty lines into account", func(t *testing.T) {
		require.Equal(t, "a\n\nb", blockStringValue("a\n\n b"))
		require.Equal(t, "a\n\nb", blockStringValue("a\n \n  b"))
	})

	t.Run("does not alter trailing spaces", func(t *testing.T) {
		result := blockStringValue(`

                
    Hello,      
      World!    
                
    Yours,      
      GraphQL.  
                

`)

		require.Equal(
			t,
			"Hello,      \n  World!    \n            \nYours,      \n  GraphQL.  ",
			result,
		)
	})
}
