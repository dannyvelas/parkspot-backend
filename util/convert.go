package util

import (
	"regexp"
	"strconv"
)

func ToPosInt(value string) int {
	if len(value) > 0 && value[0] == '-' {
		value = value[1:]
	}

	if parsed, err := strconv.Atoi(value); err != nil {
		return 0
	} else {
		return parsed
	}
}

func IsUUIDV4(id string) bool {
	re := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return re.MatchString(id)
}

func ToBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parsed
}
