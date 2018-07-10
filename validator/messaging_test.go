package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessaging(t *testing.T) {
	t.Run("orList", func(t *testing.T) {
		assert.Equal(t, "", orList())
		assert.Equal(t, "A", orList("A"))
		assert.Equal(t, "A or B", orList("A", "B"))
		assert.Equal(t, "A, B, or C", orList("A", "B", "C"))
		assert.Equal(t, "A, B, C, or D", orList("A", "B", "C", "D"))
		assert.Equal(t, "A, B, C, D, or E", orList("A", "B", "C", "D", "E", "F"))
	})

	t.Run("quotedOrList", func(t *testing.T) {
		assert.Equal(t, ``, quotedOrList())
		assert.Equal(t, `"A"`, quotedOrList("A"))
		assert.Equal(t, `"A" or "B"`, quotedOrList("A", "B"))
		assert.Equal(t, `"A", "B", or "C"`, quotedOrList("A", "B", "C"))
		assert.Equal(t, `"A", "B", "C", or "D"`, quotedOrList("A", "B", "C", "D"))
	})
}
