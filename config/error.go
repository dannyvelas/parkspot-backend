package config

import (
	"fmt"
)

type configError interface {
	Error() string
	ErrorUsingDefault(interface{}) string
}

type notFoundError struct {
	variableName string
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("No config value found for %s", e.variableName)
}

func (e notFoundError) ErrorUsingDefault(val interface{}) string {
	return fmt.Sprintf("%s. Using default of: %v", e, val)
}

type conversionError struct {
	variableName    string
	desiredTypeName string
}

func (e conversionError) Error() string {
	return fmt.Sprintf("%s could not be converted to type %s", e.variableName, e.desiredTypeName)
}

func (e conversionError) ErrorUsingDefault(val interface{}) string {
	return fmt.Sprintf("%s. Using default of: %v", e, val)
}

type invalidError struct {
	variableName string
	reason       string
}

func (e invalidError) Error() string {
	return fmt.Sprintf("%s is invalid because it is %s", e.variableName, e.reason)
}

func (e invalidError) ErrorUsingDefault(val interface{}) string {
	return fmt.Sprintf("%s. Using default of: %v", e, val)
}
