package utils

import (
	"regexp"
)

// ValidatePhoneNumber validates phone number format
// Supports formats like: +989123456789, 09123456789, 9123456789
func ValidatePhoneNumber(phone string) bool {
	// Remove spaces and dashes
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	cleanedPhone := regexp.MustCompile(`[\s-]`).ReplaceAllString(phone, "")
	
	// Check if phone number is between 10 and 15 digits
	if len(cleanedPhone) < 10 || len(cleanedPhone) > 15 {
		return false
	}
	
	return phoneRegex.MatchString(cleanedPhone)
}

