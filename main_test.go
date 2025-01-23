package main

import "testing"

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user123@domain.co", true},
		{"invalid-email.com", false},
		{"@domain.com", false},
		{"user@.com", false},
		{"user@domain", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isValidEmail(tc.email)
		if result != tc.expected {
			t.Errorf("isValidEmail(%s) = %v; expected %v", tc.email, result, tc.expected)
		}
	}
}
