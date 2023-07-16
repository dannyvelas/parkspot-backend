package models

import (
	"github.com/dannyvelas/lasvistas_api/util"
)

// test_ prefix should help differentiate variables that are used for tests
// with variables that are not used for tests

var (
	// this is the default test resident. this resident has limited parking days
	Test_resident = Resident{
		ID:        "B1234567",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}
	// this is a test resident with unlimited parking days
	Test_residentUnlimDays = Resident{
		ID:        "B7654321",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email2@example.com",
		Password:  "notapassword",
		UnlimDays: util.ToPtr(true)}
	// this is the default test car. the associated resident is test_resident
	Test_car = NewCar(
		"d1e0affb-14e7-4e9f-b8a3-70be7d49d063",
		Test_resident.ID,
		"lp1",
		"color",
		"make",
		"model",
		0)
)
