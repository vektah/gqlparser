package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/lexer"
)

func TestParserUtils(t *testing.T) {
	t.Run("test lookaround", func(t *testing.T) {
		p := newParser("asdf 1.0 turtles")
		require.Equal(t, "asdf", p.peek().Value)
		require.Equal(t, "asdf", p.expectKeyword("asdf").Value)
		require.Equal(t, "asdf", p.prev.Value)
		require.Nil(t, p.err)

		require.Equal(t, "1.0", p.peek().Value)
		require.Equal(t, "1.0", p.peek().Value)
		require.Equal(t, "1.0", p.expect(lexer.Float).Value)
		require.Equal(t, "1.0", p.prev.Value)
		require.Nil(t, p.err)

		require.True(t, p.skip(lexer.Name))
		require.Nil(t, p.err)

		require.Equal(t, lexer.EOF, p.peek().Kind)
		require.Nil(t, p.err)
	})

	t.Run("test many can read array", func(t *testing.T) {
		p := newParser("[a b c d]")

		var arr []string
		p.many(lexer.BracketL, lexer.BracketR, func() {
			arr = append(arr, p.next().Value)
		})
		require.Nil(t, p.err)
		require.Equal(t, []string{"a", "b", "c", "d"}, arr)

		require.Equal(t, lexer.EOF, p.peek().Kind)
		require.Nil(t, p.err)
	})

	t.Run("test many return if open is not found", func(t *testing.T) {
		p := newParser("turtles are happy")

		p.many(lexer.BracketL, lexer.BracketR, func() {
			t.Error("cb should not be called")
		})
		require.Nil(t, p.err)
		require.Equal(t, "turtles", p.next().Value)
	})

	t.Run("test many will stop on error", func(t *testing.T) {
		p := newParser("[a b c d]")

		var arr []string
		p.many(lexer.BracketL, lexer.BracketR, func() {
			arr = append(arr, p.next().Value)
			if len(arr) == 2 {
				p.error(p.peek(), "boom")
			}
		})
		require.EqualError(t, p.err, "Syntax Error: boom (line 1, column 6)")
		require.Equal(t, []string{"a", "b"}, arr)
	})

	t.Run("test errors", func(t *testing.T) {
		p := newParser("foo bar")

		p.next()
		p.error(p.peek(), "test error")
		p.error(p.peek(), "secondary error")

		require.EqualError(t, p.err, "Syntax Error: test error (line 1, column 5)")

		require.Equal(t, "foo", p.peek().Value)
		require.Equal(t, "foo", p.next().Value)
		require.Equal(t, "foo", p.peek().Value)
	})

	t.Run("unexpected error", func(t *testing.T) {
		p := newParser("1 3")
		p.unexpectedError()
		require.EqualError(t, p.err, "Syntax Error: Unexpected Int \"1\" (line 1, column 1)")
	})

	t.Run("unexpected error", func(t *testing.T) {
		p := newParser("1 3")
		p.unexpectedToken(p.next())
		require.EqualError(t, p.err, "Syntax Error: Unexpected Int \"1\" (line 1, column 1)")
	})

	t.Run("expect error", func(t *testing.T) {
		p := newParser("foo bar")
		p.expect(lexer.Float)

		require.EqualError(t, p.err, "Syntax Error: Expected Float, found Name (line 1, column 1)")
	})

	t.Run("expectKeyword error", func(t *testing.T) {
		p := newParser("foo bar")
		p.expectKeyword("baz")

		require.EqualError(t, p.err, "Syntax Error: Expected \"baz\", found Name \"foo\" (line 1, column 1)")
	})
}

func newParser(input string) parser {
	return parser{lexer: lexer.New(input)}
}
