package matcher

import (
	"reflect"
	"testing"
)

func TestTokenizeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"PascalCase", "EncryptService", []string{"encrypt", "service"}},
		{"camelCase", "decryptService", []string{"decrypt", "service"}},
		{"snake_case", "encrypt_service", []string{"encrypt", "service"}},
		{"kebab-case", "decrypt-service", []string{"decrypt", "service"}},
		{"mixed_symbols", "User_Account-ID", []string{"user", "account", "id"}},
		{"acronyms", "XMLParserAPI", []string{"xml", "parser", "api"}},
		{"numbers_included", "OAuth2Client", []string{"o", "auth", "2", "client"}},
		{"numbers_as_separator", "user2Controller", []string{"user", "2", "controller"}},
		{"single_word", "Admin", []string{"admin"}},
		{"empty_string", "", []string{}},
		{"only_symbols", "_-*$", []string{}},
		{"already_separated", "account service", []string{"account", "service"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := TokenizeIdentifier(tc.input)
			
			// Handle edge case where deep equal fails on empty slice comparison map/struct
			if len(result) == 0 && len(tc.expected) == 0 {
				return // Pass
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Tokenize(%q) = %v; want %v", tc.input, result, tc.expected)
			}
		})
	}
}
