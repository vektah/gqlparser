package lexer

import (
	"math"
	"strings"
)

// blockStringValue produces the value of a block string from its parsed raw value, similar to
// Coffeescript's block string, Python's docstring trim or Ruby's strip_heredoc.
//
// This implements the GraphQL spec's BlockStringValue() static algorithm.
func blockStringValue(raw string) string {
	lines := strings.Split(raw, "\n")

	commonIndent := math.MaxInt32
	for i, line := range lines {
		if i == 0 {
			// The first line is the text right after the opening quotes, so
			// its indentation is part of the value, not a common indentation.
			continue
		}
		indent := leadingWhitespace(line)
		if indent < len(line) && indent < commonIndent {
			commonIndent = indent
			if commonIndent == 0 {
				break
			}
		}
	}

	if commonIndent != math.MaxInt32 && len(lines) > 0 {
		for i := 1; i < len(lines); i++ {
			if len(lines[i]) < commonIndent {
				lines[i] = ""
			} else {
				lines[i] = lines[i][commonIndent:]
			}
		}
	}

	start := 0
	end := len(lines)

	for start < end && leadingWhitespace(lines[start]) == math.MaxInt32 {
		start++
	}

	for start < end && leadingWhitespace(lines[end-1]) == math.MaxInt32 {
		end--
	}

	return strings.Join(lines[start:end], "\n")
}

func leadingWhitespace(str string) int {
	for i, r := range str {
		if r != ' ' && r != '\t' {
			return i
		}
	}
	// this line is made up entirely of whitespace, its leading whitespace doesnt count.
	return math.MaxInt32
}
