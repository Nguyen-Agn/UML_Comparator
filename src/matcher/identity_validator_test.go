package matcher

import (
	"testing"
)

func TestStandardIdentityValidator_IsValid(t *testing.T) {
	antonyms := NewAntonymDetector()
	validator := NewStandardIdentityValidator(antonyms)

	tests := []struct {
		name     string
		name1    string
		name2    string
		expected bool
	}{
		{"Exact match", "UserService", "UserService", true},
		{"Typo tolerant", "Account", "Acount", true}, // Not antonyms => allowed to pass
		
		{"Direct Prefix Antonyms", "EncryptService", "DecryptService", false}, // encrypt vs decrypt
		{"PascalCase Antonyms", "CreateUser", "DeleteUser", false}, // create vs delete
		{"Inter-position Antonyms", "UserLoginController", "UserLogoutController", false}, // login vs logout
		
		{"Standard Typo", "SystemManager", "SystmManagr", true}, // Handled by fuzzy matcher later
		
		{"Different Words", "Dog", "Cat", true}, // Not antonyms. Fuzzy match will reject them due to low score.
		{"camelCase opposites", "encodeBase64", "decodeBase64", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.IsValid(tc.name1, tc.name2)
			if result != tc.expected {
				t.Errorf("IsValid(%q, %q) = %v; want %v", tc.name1, tc.name2, result, tc.expected)
			}
		})
	}
}
