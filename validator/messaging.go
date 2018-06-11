package validator

import "bytes"

// Given [ A, B, C ] return '"A", "B", or "C"'.
func quotedOrList(items ...string) string {
	itemsQuoted := make([]string, len(items))
	for i, item := range items {
		itemsQuoted[i] = `"` + item + `"`
	}
	return orList(itemsQuoted...)
}

// Given [ A, B, C ] return 'A, B, or C'.
func orList(items ...string) string {
	var buf bytes.Buffer

	for i, item := range items {
		if i != 0 {
			if i == len(items)-1 {
				buf.WriteString(" or ")
			} else {
				buf.WriteString(", ")
			}
		}
		buf.WriteString(item)
	}
	return buf.String()
}
