package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
)

type PermitService struct {
	permitRepo   storage.PermitRepo
	residentRepo storage.ResidentRepo
	carRepo      storage.CarRepo
}

func NewPermitService(permitRepo storage.PermitRepo, residentRepo storage.ResidentRepo, carRepo storage.CarRepo) PermitService {
	return PermitService{
		permitRepo:   permitRepo,
		residentRepo: residentRepo,
		carRepo:      carRepo,
	}
}

func (s PermitService) GetAll(permitFilter models.PermitFilter, limit, page int, reversed bool, search string, residentID string) (models.ListWithMetadata[models.Permit], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allPermits, err := s.permitRepo.Get(permitFilter, residentID, boundedLimit, offset, reversed, search)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting permits from permit repo: %v", err)
	}

	totalAmount, err := s.permitRepo.GetCount(permitFilter, residentID)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting total amount from permit repo: %v", err)
	}

	return models.NewListWithMetadata(allPermits, totalAmount), nil
}

func (s PermitService) GetOne(id int) (models.Permit, error) {
	permit, err := s.permitRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Permit{}, ErrNotFound
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	return permit, nil
}

func (s PermitService) Create(desiredPermit models.CreatePermit, desiredCar models.CreateCar) (models.Permit, error) {
	// error out if resident DNE
	existingResident, err := s.residentRepo.GetOne(desiredPermit.ResidentID)
	if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
		return models.Permit{}, fmt.Errorf("error getting one from residentRepo: %v", err)
	} else if errors.Is(err, storage.ErrNoRows) { // resident does not exist
		return models.Permit{}, ErrNoResident
	}

	// check if car exists
	existingCar, err := s.carRepo.GetByLicensePlate(desiredCar.LicensePlate)
	if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
		return models.Permit{}, fmt.Errorf("error getting by licensePlate in carRepo: %v", err)
	} else if errors.Is(err, storage.ErrNoRows) {
		// no-op: if car DNE, this is valid and acceptable
	}

	err = s.validateCreation(desiredPermit, existingResident, existingCar)
	if err != nil {
		return models.Permit{}, err
	}

	var carToUse models.Car
	// if car exists and has all fields, use it for this permit
	if existingCar != nil && existingCar.Make != "" && existingCar.Model != "" {
		carToUse = *existingCar
	} else { // otherwise, create it or update it
		carToUse, err = s.upsertCar(existingCar, desiredCar)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error upserting car when creating permit: %v", err)
		}
	}

	permitLength := s.getAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
	affectsDays := desiredPermit.ExceptionReason == "" && !existingResident.UnlimDays
	if affectsDays {
		err = s.residentRepo.AddToAmtParkingDaysUsed(existingResident.ID, permitLength)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error adding to amt parking days used in residentRepo: %v", err)
		}

		err = s.carRepo.AddToAmtParkingDaysUsed(carToUse.ID, permitLength)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error adding to amt parking days used in carRepo: %v", err)
		}
	}

	permitID, err := s.permitRepo.Create(
		desiredPermit.ResidentID,
		carToUse.ID,
		desiredPermit.StartDate,
		desiredPermit.EndDate,
		affectsDays,
		desiredPermit.ExceptionReason,
	)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error create new permit in permitRepo: %v", err)
	}

	newPermit, err := s.permitRepo.GetOne(permitID)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error getting permit after having created it in permitRepo: %v", err)
	}

	return newPermit, nil
}

func (s PermitService) validateCreation(desiredPermit models.CreatePermit, existingResident models.Resident, existingCar *models.Car) error {
	// error out if car exists and has active permits during dates requested
	if existingCar != nil { // car exists
		activePermitsDuring, err := s.permitRepo.GetActiveOfCarDuring(
			existingCar.ID, desiredPermit.StartDate, desiredPermit.EndDate)
		if err != nil {
			return fmt.Errorf("error getting active of car during dates in permitRepo: %v", err)
		} else if len(activePermitsDuring) != 0 {
			return ErrCarActivePermit
		}
	}

	// if this is an exception, there are no more checks to be performed. so return no errors
	if desiredPermit.ExceptionReason != "" {
		return nil
	}

	permitLength := s.getAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
	if permitLength > config.MaxPermitLength {
		return ErrPermitTooLong
	}

	activePermitsDuring, err := s.permitRepo.GetActiveOfResidentDuring(
		existingResident.ID, desiredPermit.StartDate, desiredPermit.EndDate)
	if err != nil {
		return fmt.Errorf("error getting active of resident during dates in permitRepo: %v", err)
	} else if len(activePermitsDuring) >= 2 {
		return ErrResidentTwoActivePermits
	}

	if !existingResident.UnlimDays {
		if existingResident.AmtParkingDaysUsed >= config.MaxParkingDays {
			return errEntityDaysTooLong("resident", existingResident.AmtParkingDaysUsed)
		} else if existingResident.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
			return errPermitPlusEntityDaysTooLong("resident", existingResident.AmtParkingDaysUsed)
		}

		if existingCar != nil && existingCar.AmtParkingDaysUsed >= config.MaxParkingDays {
			return errEntityDaysTooLong("car", existingCar.AmtParkingDaysUsed)
		} else if existingCar.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
			return errPermitPlusEntityDaysTooLong("car", existingCar.AmtParkingDaysUsed)
		}
	}

	return nil
}

func (s PermitService) Delete(id int) error {
	permit, err := s.permitRepo.GetOne(id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	err = s.permitRepo.Delete(id)
	if errors.Is(err, storage.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	permitLength := int(permit.EndDate.Sub(permit.StartDate).Hours() / 24)
	if permit.AffectsDays {
		err = s.residentRepo.AddToAmtParkingDaysUsed(permit.ResidentID, -permitLength)
		if err != nil {
			return fmt.Errorf("error adding to amtParkingDaysUsed in residentRepo: %v", err)
		}

		err = s.carRepo.AddToAmtParkingDaysUsed(permit.Car.ID, -permitLength)
		if err != nil {
			return fmt.Errorf("error adding to amtParkingDaysUsed in carRepo: %v", err)
		}
	}

	return nil
}

// helpers
func (s PermitService) getAmtDays(startDate, endDate int64) int {
	const amtSecondsInADay = 86400
	return int((endDate - startDate) / amtSecondsInADay)
}

func (s PermitService) upsertCar(existingCar *models.Car, desiredCar models.CreateCar) (models.Car, error) {
	// car exits but missing fields
	if existingCar != nil {
		err := s.carRepo.Update(existingCar.ID, desiredCar.Color, desiredCar.Make, desiredCar.Model)
		if err != nil {
			return models.Car{}, fmt.Errorf("error updating car which is missing fields: %v", err)
		}
		newCar := models.NewCar(existingCar.ID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, existingCar.AmtParkingDaysUsed)
		return newCar, nil
	}

	// car DNE
	carID, err := s.carRepo.Create(desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("error creating car with carRepo: %v", err)
	}

	newCar := models.NewCar(carID, desiredCar.LicensePlate, desiredCar.Color, desiredCar.Make, desiredCar.Model, 0)
	return newCar, nil
}
