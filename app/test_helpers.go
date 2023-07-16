package app

import (
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
)

// test_ prefix should help differentiate global package variables that are used for tests
// with global package variables that are not used for tests

var (
	// this is the default test resident. this resident has limited parking days
	test_resident = models.Resident{
		ID:        "B1234567",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email@example.com",
		Password:  "notapassword"}
	// this is a test resident with unlimited parking days
	test_residentUnlimDays = models.Resident{
		ID:        "B7654321",
		FirstName: "Daniel",
		LastName:  "Velasquez",
		Phone:     "1234567890",
		Email:     "email2@example.com",
		Password:  "notapassword",
		UnlimDays: util.ToPtr(true)}
	// this is the default test car. the associated resident is test_resident
	test_car = models.NewCar(
		"d1e0affb-14e7-4e9f-b8a3-70be7d49d063",
		test_resident.ID,
		"lp1",
		"color",
		"make",
		"model",
		0)
)
