package webserver_test

import (
	"testing"
)

// Test helper functions (simplified versions)
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 100 {
		return false
	}

	atCount := 0
	for _, char := range email {
		if char == '@' {
			atCount++
		}
	}

	return atCount == 1 && email[0] != '@' && email[len(email)-1] != '@'
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid Strong Password", "StrongPass123!", true},
		{"Too Short", "weak", false},
		{"No Uppercase", "lowercase123!", false},
		{"No Lowercase", "UPPERCASE123!", false},
		{"No Numbers", "NoNumbers!", false},
		{"No Special Characters", "NoSpecial123", false},
		{"Valid Minimum", "Valid1!", true},
		{"Empty Password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidPassword(tt.password)
			if result != tt.expected {
				t.Errorf("isValidPassword(%q) = %v, want %v", tt.password, result, tt.expected)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid Email", "test@example.com", true},
		{"Valid Email with subdomain", "user@mail.example.com", true},
		{"Invalid - No @", "invalid.email", false},
		{"Invalid - No domain", "test@", false},
		{"Invalid - No username", "@example.com", false},
		{"Invalid - Multiple @", "test@@example.com", false},
		{"Empty Email", "", false},
		{"Too Long Email", "verylongemailaddressthatexceedsthemaximumlengthallowed@verylongdomainnamethatisnotvalid.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"Valid Username", "validuser", true},
		{"Valid with numbers", "user123", true},
		{"Valid with underscore", "user_name", true},
		{"Too Short", "ab", false},
		{"Too Long", "thisusernameiswaytoolongtobevalid", false},
		{"Invalid characters", "user@name", false},
		{"Empty Username", "", false},
		{"Only spaces", "   ", false},
		{"Special characters", "user-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidUsername(tt.username)
			if result != tt.expected {
				t.Errorf("isValidUsername(%q) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}
