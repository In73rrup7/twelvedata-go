package twelvedata

import (
	"strconv"
	"strings"
)

// TDFloat64 handles float values from TwelveData API that may contain malformed strings.
// The API sometimes returns bare signs ("+", "-") or empty strings instead of valid numbers.
// TDFloat64 defaults to 0 for any unparseable value rather than failing the entire unmarshal.
type TDFloat64 float64

func (f *TDFloat64) UnmarshalJSON(data []byte) error {
	s := string(data)

	// Handle JSON null
	if s == "null" {
		*f = 0
		return nil
	}

	// Strip quotes if present (API returns string-encoded numbers like "123.45")
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Handle empty string or bare sign characters
	if s == "" || s == "+" || s == "-" {
		*f = 0
		return nil
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Default to 0 for any unparseable value
		*f = 0
		return nil
	}

	*f = TDFloat64(v)
	return nil
}
