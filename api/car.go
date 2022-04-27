package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
)

type createCarReq struct {
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func (createCarReq createCarReq) emptyFields() error {
	emptyFields := []string{}

	if createCarReq.LicensePlate == "" {
		emptyFields = append(emptyFields, "LicensePlate")
	}
	if createCarReq.Color == "" {
		emptyFields = append(emptyFields, "Color")
	}
	if createCarReq.Make == "" {
		emptyFields = append(emptyFields, "Make")
	}
	if createCarReq.Model == "" {
		emptyFields = append(emptyFields, "Model")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createCarReq createCarReq) invalidFields() error {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(createCarReq.LicensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}

	if len(createCarReq.LicensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

	if !regexp.MustCompile("^[A-Za-z]+$").MatchString(createCarReq.Color) {
		errors = append(errors, "color must be one word only letters")
	}

	makeAndModelRe := regexp.MustCompile("^[A-Za-z0-9 -]+$")
	if !makeAndModelRe.MatchString(createCarReq.Make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}

	if !makeAndModelRe.MatchString(createCarReq.Model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createCarReq createCarReq) toModels() (models.CreateCar, error) {
	if err := createCarReq.emptyFields(); err != nil {
		return models.CreateCar{}, err
	}

	if err := createCarReq.invalidFields(); err != nil {
		return models.CreateCar{}, err
	}

	return models.NewCreateCar(
		createCarReq.LicensePlate,
		createCarReq.Color,
		createCarReq.Make,
		createCarReq.Model,
	), nil
}
