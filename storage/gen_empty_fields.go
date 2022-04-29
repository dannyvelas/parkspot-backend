package storage

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"reflect"
)

type modelType interface {
	models.Car | models.Permit | models.NewCarArgs | models.NewPermitArgs
}

// returns a map, where each key is the name of a field
// each value is a T which has that field zeroed-out
func genEmptyFields[T modelType](allFields T) map[string]T {
	tsEmptyFields := map[string]T{}

	for _, field := range reflect.VisibleFields(reflect.TypeOf(allFields)) {
		tEmptyField := allFields
		tEmptyFieldV := reflect.ValueOf(&tEmptyField).Elem()
		fieldV := tEmptyFieldV.FieldByName(field.Name)
		fieldV.SetString("")
		tsEmptyFields[field.Name] = tEmptyField
	}

	return tsEmptyFields
}
