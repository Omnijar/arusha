package util

import (
	"regexp"
)

// IsValidEmail checks the given email string for possible errors.
func IsValidEmail(email string) bool {
	match, _ := regexp.MatchString("(^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$)", email)
	return match
}
