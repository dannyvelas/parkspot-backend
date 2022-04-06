package config

import (
	"fmt"
)

type NotFoundError struct {
	variableName string
	defaultValue any
}

func (e NotFoundError) Error() string {
	err := fmt.Sprintf("No config value found for %s", e.variableName)
	return fmt.Sprintf("%s. Using default of: %v", err, e.defaultValue)
}

func newNotFoundError(variableName string, devaultValue any) NotFoundError {
	return NotFoundError{variableName, devaultValue}
}

type ConversionError struct {
	variableName    string
	desiredTypeName string
	defaultValue    any
}

func (e ConversionError) Error() string {
	err := fmt.Sprintf("%s could not be converted to type %s", e.variableName, e.desiredTypeName)
	return fmt.Sprintf("%s. Using default of: %v", err, e.defaultValue)
}

func newConversionError(variableName string, desiredTypeName string, defaultValue any) ConversionError {
	return ConversionError{variableName, desiredTypeName, defaultValue}
}
