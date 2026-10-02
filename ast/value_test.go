package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueObject(t *testing.T) {
	t.Run("undefined variable", func(t *testing.T) {
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{Kind: Variable, Raw: "Var"}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		// Treated as absent so the field's own default can be applied later.
		require.NotContains(t, val.(map[string]any), "field")
	})

	t.Run("undefined variable with default", func(t *testing.T) {
		// The variable is not supplied in vars, but its definition carries a
		// default value, so the field is present with that default.
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{
				Kind: Variable,
				Raw:  "Var",
				VariableDefinition: &VariableDefinition{
					Variable:     "Var",
					DefaultValue: &Value{Kind: IntValue, Raw: "42"},
				},
			}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		m := val.(map[string]any)
		require.Contains(t, m, "field")
		require.Equal(t, int64(42), m["field"])
	})

	t.Run("explicit null", func(t *testing.T) {
		obj := &Value{Kind: ObjectValue, Children: ChildValueList{
			{Name: "field", Value: &Value{Kind: NullValue}},
		}}
		val, err := obj.Value(map[string]any{})
		require.NoError(t, err)
		m := val.(map[string]any)
		require.Contains(t, m, "field")
		require.Nil(t, m["field"])
	})
}

// TestValueStringQuotesAsGraphQL pins String() to graphql-js printString's escaping rather
// than Go's. strconv.Quote, which this used to call, emits \x1b, \a, \v and \U0010ffff —
// none of which the GraphQL grammar accepts — so its output could not always be read back.
//
// The cases below were checked against graphql-js 16.14.2 across U+0000-U+00FF plus astral
// code points. Anything named "slow path" exists because quoteString escapes nothing until
// it meets a byte that needs it: a character can therefore be printed by two different
// branches depending only on what else shares the string with it, and both must agree.
func TestValueStringQuotesAsGraphQL(t *testing.T) {
	cases := map[string]struct{ raw, want string }{
		"no escaping needed": {"plain", `"plain"`},

		// Only " and \ are escaped by doubling. A forward slash is not: GraphQL accepts \/
		// but does not require it, and graphql-js prints the bare character.
		"quote, backslash, slash": {"\"\\/", `"\"\\/"`},

		// The five controls with a short form. Note \v is absent on purpose: Go spells
		// U+000B that way, GraphQL has no such escape, so it has to go out as \u000B.
		"controls with a short escape": {"\b\f\n\r\t", `"\b\f\n\r\t"`},
		"controls without one": {
			"\x00\x07\x0b\x1b\x1f",
			`"\u0000\u0007\u000B\u001B\u001F"`,
		},

		"printable ascii bounds": {" ~", `" ~"`},

		// DEL and the C1 block are escaped even though they are not ASCII controls, and the
		// hex is upper-case. strconv.Quote wrote \u008a here, which parses but does not match.
		"del and c1 controls": {"\x7f\u0080\u009f", `"\u007F\u0080\u009F"`},

		// 0xC2 leads the escaped C1 block *and* the printable U+00A0-U+00BF, so the lead
		// byte alone cannot decide; U+00A0 must survive as itself.
		"non-breaking space is printable": {"\u00a0", "\"\u00a0\""},

		// Printable non-ASCII is passed through, not escaped. strconv.Quote agreed about
		// the emoji, but spelled the unprintable tag character \U000e0001 — a Go escape
		// the GraphQL grammar has no rule for, so that output did not parse at all.
		"multibyte passed through":      {" é", "\" é\""},
		"astral stays verbatim":         {"\U0001F600\U000E0001", "\"\U0001F600\U000E0001\""},
		"multibyte on the slow path":    {"\né", "\"\\né\""},
		"astral on the slow path":       {"\x00\U0001F600", "\"\\u0000\U0001F600\""},
		"c1 escape next to a multibyte": {"\u0080é", "\"\\u0080é\""},

		// Raw doesn't have to hold valid UTF-8 — callers build Values by hand. Formatting
		// must not be what corrupts them: a rune-wise walk would turn each bad byte into
		// U+FFFD, and only on the slow path, so the same byte would print two ways.
		"invalid utf-8 passed through":    {"\xff", "\"\xff\""},
		"invalid utf-8 on the slow path":  {"\x00\xff", `"\u0000` + "\xff" + `"`},
		"invalid utf-8 after a 0xC2 lead": {"\u00a0\xff", "\"\u00a0\xff\""},
		"truncated two-byte sequence":     {"\xc3", "\"\xc3\""},
		"lone 0xC2 is not a c1 escape":    {"\xc2", "\"\xc2\""},

		// Deciding which path runs is only ever an optimisation: treating a harmless byte
		// as escape-worthy costs a slower route to the same text, whereas overlooking one
		// puts it in the output raw. Every case above trips that decision on its first
		// byte, so only a lone trigger sandwiched in clean text can tell the two apart.
		"backslash is the only trigger":      {`a\b`, `"a\\b"`},
		"unit separator is the only trigger": {"a\x1fb", `"a\u001Fb"`},

		// A filter that recognised every byte *except* the backslash would still escape
		// this one correctly whenever something else dragged the string onto the slow
		// path. Only a string made of nothing else forces the filter to carry it alone.
		"nothing but backslashes": {`\\`, `"\\\\"`},

		// Text ahead of the first escape is copied in one go rather than a byte at a time,
		// and the space afterwards has to survive the per-byte loop that handles the rest.
		"clean text before the first escape": {"ab\nc d", `"ab\nc d"`},

		// 0xC2 introduces an escape only when the byte after it falls in 0x80-0x9F. For
		// anything else the pair is not a C1 control and both bytes stand on their own.
		"0xC2 before ascii": {"\xc2A", "\"\xc2A\""},
		"0xC2 before del":   {"\xc2\x7f", "\"\xc2\\u007F\""},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			v := &Value{Kind: StringValue, Raw: tc.raw}
			require.Equal(t, tc.want, v.String())
		})
	}
}
