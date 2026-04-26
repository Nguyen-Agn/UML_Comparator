package drawio

import (
	"strconv"
	"strings"
)

type StyleHelper struct{}

var _ IStyleHelper = (*StyleHelper)(nil)

func NewStyleHelper() *StyleHelper {
	return &StyleHelper{}
}

// IsStyleBitSet parses a semicolon-separated Draw.io style string and checks
// if a specific key's integer value has a bit set.
// Example: "fontStyle=5" & bit 4 -> true (1|4 = 5)
func (h *StyleHelper) IsStyleBitSet(style, key string, bit int) bool {
	parts := strings.Split(style, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, key+"=") {
			valStr := part[len(key)+1:]
			val, err := strconv.Atoi(valStr)
			if err == nil {
				return (val & bit) != 0
			}
		}
	}
	return false
}
