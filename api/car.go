package api

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"regexp"
	"strings"
)

type createCarReq struct {
	licensePlate string `json:"licensePlate"`
	color        string `json:"color"`
	make         string `json:"make"`
	model        string `json:"model"`
}

func (createCarReq createCarReq) emptyFields() error {
	emptyFields := []string{}

	if createCarReq.licensePlate == "" {
		emptyFields = append(emptyFields, "LicensePlate")
	}
	if createCarReq.color == "" {
		emptyFields = append(emptyFields, "Color")
	}
	if createCarReq.make == "" {
		emptyFields = append(emptyFields, "Make")
	}
	if createCarReq.model == "" {
		emptyFields = append(emptyFields, "Model")
	}

	if len(emptyFields) > 0 {
		return fmt.Errorf("%w: %v", errEmptyFields, strings.Join(emptyFields, ", "))
	}

	return nil
}

func (createCarReq createCarReq) invalidFields() error {
	errors := []string{}

	if !regexp.MustCompile("^[A-Za-z0-9]+$").MatchString(createCarReq.licensePlate) {
		errors = append(errors, "licensePlate can only be letters or numbers")
	}

	if len(createCarReq.licensePlate) > 8 {
		errors = append(errors, "licensePlate can be maximum 8 characters")
	}

	if !regexp.MustCompile("^[A-Za-z]+$").MatchString(createCarReq.color) {
		errors = append(errors, "color must be one word only letters")
	}

	makeAndModelRe := regexp.MustCompile("^[A-Za-z0-9 -]+$")
	if !makeAndModelRe.MatchString(createCarReq.make) {
		errors = append(errors, "make can only have spaces, letters, numbers, and dashes")
	}

	if !makeAndModelRe.MatchString(createCarReq.model) {
		errors = append(errors, "model can only have spaces, letters, numbers, and dashes")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %v", errInvalidFields, strings.Join(errors, ". "))
	}

	return nil
}

func (createCarReq createCarReq) validate() error {
	if err := createCarReq.emptyFields(); err != nil {
		return err
	}

	if err := createCarReq.invalidFields(); err != nil {
		return err
	}

	return nil
}

func (createCarReq createCarReq) toNewCarArgs() models.NewCarArgs {
	return models.NewNewCarArgs(
		createCarReq.licensePlate,
		createCarReq.color,
		createCarReq.make,
		createCarReq.model,
	)
}
