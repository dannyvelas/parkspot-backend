package config

import "fmt"

type NotFoundError struct {
	variableName string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("No config value found for %s", e.variableName)
}

type ConversionError struct {
	variableName    string
	desiredTypeName string
}

func (e ConversionError) Error() string {
	return fmt.Sprintf("%s could not be converted to type %s", e.variableName, e.desiredTypeName)
}
