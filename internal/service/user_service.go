package service

import (
	"errors"
	"regexp"
)

func ValidatePassword(password string) error {
	var (
		minLen    = 8
		hasLetter = `[A-Za-z]`
		hasNumber = `[0-9]`
		hasSpecial= `[!@#~$%^&*()+|_]`
	)

	if len(password) < minLen {
		return errors.New("password must be at least 8 characters long")
	}
	if !regexp.MustCompile(hasLetter).MatchString(password) {
		return errors.New("password must include at least one letter")
	}
	if !regexp.MustCompile(hasNumber).MatchString(password) {
		return errors.New("password must include at least one number")
	}
	if !regexp.MustCompile(hasSpecial).MatchString(password) {
		return errors.New("password must include at least one special character")
	}
	return nil
}

func ValidatePhoneNumber(phone string) bool {
	// Contact number validation: +countrycode + digits
	match, _ := regexp.MatchString(`^\+[1-9]\d{1,14}$`, phone)
	return match
}
