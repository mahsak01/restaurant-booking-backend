package utils

import (
	"regexp"
	"strings"
)

// ValidatePhoneNumber validates phone number format
// Supports formats like: +989123456789, 09123456789, 9123456789, +989338467840
func ValidatePhoneNumber(phone string) bool {
	// Remove spaces, dashes, and other non-digit characters except +
	cleanedPhone := strings.ReplaceAll(phone, " ", "")
	cleanedPhone = strings.ReplaceAll(cleanedPhone, "-", "")
	cleanedPhone = strings.ReplaceAll(cleanedPhone, "(", "")
	cleanedPhone = strings.ReplaceAll(cleanedPhone, ")", "")

	// Check if phone number is between 10 and 15 characters (including +)
	if len(cleanedPhone) < 10 || len(cleanedPhone) > 15 {
		return false
	}

	// Pattern: optional +, then 1-9, then 9-14 digits
	// This supports: +989123456789, 09123456789, 9123456789
	phoneRegex := regexp.MustCompile(`^(\+?[1-9]\d{9,14}|0\d{9,10})$`)

	return phoneRegex.MatchString(cleanedPhone)
}
