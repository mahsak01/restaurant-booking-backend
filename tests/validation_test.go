package tests

import (
	"restaurant-booking-backend/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePhoneNumber(t *testing.T) {
	t.Run("Valid phone numbers", func(t *testing.T) {
		validPhones := []string{
			"09123456789",
			"+989123456789",
			"9123456789",
			"00989123456789",
		}

		for _, phone := range validPhones {
			assert.True(t, utils.ValidatePhoneNumber(phone), "Phone %s should be valid", phone)
		}
	})

	t.Run("Invalid phone numbers", func(t *testing.T) {
		invalidPhones := []string{
			"123",
			"invalid",
			"",
			"12345678901234567890", // Too long
			"abc123456789",
		}

		for _, phone := range invalidPhones {
			assert.False(t, utils.ValidatePhoneNumber(phone), "Phone %s should be invalid", phone)
		}
	})
}

