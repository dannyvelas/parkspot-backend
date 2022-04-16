package typesafe

import (
	"reflect"
)

func ZeroValFields(aStruct any) []string {
	structV := reflect.ValueOf(aStruct)
	structT := reflect.TypeOf(aStruct)

	zeroValFields := []string{}
	for i := 0; i < structV.NumField(); i++ {
		fieldV := structV.Field(i)
		fieldT := structT.Field(i)
		if fieldV.Kind() != reflect.Bool && fieldV.IsZero() {
			zeroValFields = append(zeroValFields, fieldT.Name)
		}
	}
	return zeroValFields
}
