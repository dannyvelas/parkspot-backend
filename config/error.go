package config

import (
	"fmt"
)

type ConfigError interface {
	Error() string
	ErrorUsingDefault(interface{}) string
}

type NotFoundError struct {
	variableName string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("No config value found for %s", e.variableName)
}

func (e NotFoundError) ErrorUsingDefault(val interface{}) string {
	return fmt.Sprintf("%s. Using default of: %v", e.Error(), val)
}

type ConversionError struct {
	variableName    string
	desiredTypeName string
}

func (e ConversionError) Error() string {
	return fmt.Sprintf("%s could not be converted to type %s", e.variableName, e.desiredTypeName)
}

func (e ConversionError) ErrorUsingDefault(val interface{}) string {
	return fmt.Sprintf("%s. Using default of: %v", e.Error(), val)
}
