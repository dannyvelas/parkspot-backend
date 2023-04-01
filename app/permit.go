package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/util"
	"strings"
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

func (s PermitService) GetAll(permitFilter models.PermitFilter, limit, page int, reversed bool, search, residentID string) (models.ListWithMetadata[models.Permit], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allPermits, err := s.permitRepo.Get(permitFilter, residentID, boundedLimit, offset, reversed, search)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting permits from permit repo: %v", err)
	}

	totalAmount, err := s.permitRepo.GetCount(permitFilter, residentID, search)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting total amount from permit repo: %v", err)
	}

	return models.NewListWithMetadata(allPermits, totalAmount), nil
}

func (s PermitService) GetOne(id int) (models.Permit, error) {
	return s.permitRepo.GetOne(id)
}

func (s PermitService) Delete(permitID int) error {
	permit, err := s.permitRepo.GetOne(permitID)
	if err != nil {
		return err
	}

	err = s.permitRepo.Delete(permitID)
	if err != nil {
		return err
	}

	permitLength := int(permit.EndDate.Sub(permit.StartDate).Hours() / 24)
	if permit.AffectsDays {
		err = s.residentRepo.AddToAmtParkingDaysUsed(permit.ResidentID, -permitLength)
		if err != nil {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in residentRepo: %v", err)
		}

		err = s.carRepo.AddToAmtParkingDaysUsed(permit.CarID, -permitLength)
		if err != nil && !errors.Is(err, errs.NotFound) {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in carRepo: %v", err)
		}
		// purposely not returning error if error.Is(err, errs.NotFound)
		// its possible that the car with id of permit.CarID was deleted and no longer exists.
		// unlike a resident deletion, a car deletion does not cascade delete its permits
	}

	return nil
}

func (s PermitService) ValidateAndCreate(desiredPermit models.Permit) (models.Permit, error) {
	permitLength := util.GetAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)

	resident, err := s.getAndValidateResident(desiredPermit, permitLength)
	if err != nil {
		return models.Permit{}, err
	}

	// if carID != "", this permit will be created for a pre-existing car
	// find and validate that this car follows policy for creating a permit
	if desiredPermit.CarID != "" {
		car, err := s.getAndValidateCar(desiredPermit, *resident.UnlimDays, permitLength)
		if err != nil {
			return models.Permit{}, err
		}

		// get a snapshot of car and save it into the permit
		desiredPermit.LicensePlate = car.LicensePlate
		desiredPermit.Color = car.Color
		desiredPermit.Make = car.Make
		desiredPermit.Model = car.Model
	} else if err := models.NewCarFieldValidator(false).Validate(desiredPermit.LicensePlate, desiredPermit.Color, desiredPermit.Make, desiredPermit.Model); err != nil {
		return models.Permit{}, err
	}

	if err := s.validateDates(desiredPermit); err != nil {
		return models.Permit{}, err
	}

	desiredPermit.AffectsDays = desiredPermit.ExceptionReason == "" && !*resident.UnlimDays
	createdPermit, err := s.create(desiredPermit)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error creating permit in permitservice: %v", err)
	}

	return createdPermit, nil
}

func (s PermitService) Update(updatedFields models.Permit) (models.Permit, error) {
	if updatedFields.ID == 0 {
		return models.Permit{}, errs.MissingIDField
	}
	if updatedFields.LicensePlate == "" && updatedFields.Color == "" && updatedFields.Make == "" && updatedFields.Model == "" {
		return models.Permit{}, errs.AllEditFieldsEmpty("licensePlate, color, make, model")
	}
	if err := models.NewCarFieldValidator(true).Validate(updatedFields.LicensePlate, updatedFields.Color, updatedFields.Make, updatedFields.Model); err != nil {
		return models.Permit{}, err
	}

	if err := s.permitRepo.Update(updatedFields); err != nil {
		return models.Permit{}, fmt.Errorf("error updating permit from permitRepo: %w", err)
	}

	permit, err := s.permitRepo.GetOne(updatedFields.ID)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error getting permit from permitRepo: %w", err)
	}

	return permit, nil
}

// helpers
func (s PermitService) getAndValidateResident(desiredPermit models.Permit, permitLength int) (models.Resident, error) {
	if err := models.IsResidentID(desiredPermit.ResidentID); err != nil {
		return models.Resident{}, errs.InvalidResID
	}

	resident, err := s.residentRepo.GetOne(desiredPermit.ResidentID)
	if errors.Is(err, errs.NotFound) {
		return models.Resident{}, errs.ResidentForPermitDNE
	} else if err != nil {
		return models.Resident{}, fmt.Errorf("error getting one from resident repo: %v", err)
	}

	// if this is an exception, there are no more checks to be performed. so return no errors
	if desiredPermit.ExceptionReason != "" {
		return resident, nil
	}

	if permitLength > config.MaxPermitLength {
		return models.Resident{}, errs.PermitTooLong
	}

	residentActivePermitsDuring, err := s.permitRepo.GetActiveOfResidentDuring(
		resident.ID, desiredPermit.StartDate, desiredPermit.EndDate)
	if err != nil {
		return models.Resident{}, fmt.Errorf("error getting active of resident during dates in permitRepo: %v", err)
	} else if len(residentActivePermitsDuring) >= 2 {
		return models.Resident{}, errs.ResidentTwoActivePermits
	}

	if resident.UnlimDays == nil || resident.AmtParkingDaysUsed == nil {
		return models.Resident{}, fmt.Errorf("data type error in permit service validate create. unlimDays or amtParkingDaysUsed is nil")
	}

	if !*resident.UnlimDays {
		if *resident.AmtParkingDaysUsed >= config.MaxParkingDays {
			return models.Resident{}, errs.EntityDaysTooLong("resident", *resident.AmtParkingDaysUsed)
		} else if *resident.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
			return models.Resident{}, errs.PermitPlusEntityDaysTooLong("resident", *resident.AmtParkingDaysUsed)
		}
	}

	return resident, nil
}

func (s PermitService) getAndValidateCar(desiredPermit models.Permit, unlimDays bool, permitLength int) (models.Car, error) {
	if !util.IsUUIDV4(desiredPermit.CarID) {
		return models.Car{}, errs.InvalidFields("CarID is not a UUID")
	}

	existingCar, err := s.carRepo.GetOne(desiredPermit.CarID)
	if errors.Is(err, errs.NotFound) {
		return models.Car{}, errs.CarForPermitDNE
	} else if err != nil {
		return models.Car{}, fmt.Errorf("error getting one from carRepo: %v", err)
	}

	// we found the car: error out if it has active permits during dates requested
	carActivePermitsDuring, err := s.permitRepo.GetActiveOfCarDuring(
		existingCar.ID, desiredPermit.StartDate, desiredPermit.EndDate)
	if err != nil {
		return models.Car{}, fmt.Errorf("error getting active of car during dates in permitRepo: %v", err)
	} else if len(carActivePermitsDuring) != 0 {
		return models.Car{}, errs.CarActivePermit
	}

	// if this is an exception, or if this resident has unlimited days,
	// there are no more car checks to be performed. so return no errors
	if desiredPermit.ExceptionReason != "" || unlimDays {
		return existingCar, nil
	}

	if *existingCar.AmtParkingDaysUsed >= config.MaxParkingDays {
		return models.Car{}, errs.EntityDaysTooLong("car", *existingCar.AmtParkingDaysUsed)
	} else if *existingCar.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
		return models.Car{}, errs.PermitPlusEntityDaysTooLong("car", *existingCar.AmtParkingDaysUsed)
	}

	return existingCar, nil
}

func (s PermitService) validateDates(desiredPermit models.Permit) error {
	errors := []string{}

	if desiredPermit.StartDate.IsZero() {
		errors = append(errors, "startDate cannot be empty")
	}
	if desiredPermit.EndDate.IsZero() {
		errors = append(errors, "endDate cannot be empty")
	}
	if desiredPermit.StartDate.After(desiredPermit.EndDate) {
		errors = append(errors, "startDate cannot be after endDate")
	}
	if desiredPermit.StartDate.Equal(desiredPermit.EndDate) {
		errors = append(errors, "startDate cannot be equal to endDate")
	}

	if len(errors) > 0 {
		return errs.InvalidFields(strings.Join(errors, ". "))
	}

	return nil
}

func (s PermitService) create(desiredPermit models.Permit) (models.Permit, error) {
	permitLength := util.GetAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
	if desiredPermit.AffectsDays {
		err := s.residentRepo.AddToAmtParkingDaysUsed(desiredPermit.ResidentID, permitLength)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error adding to amt parking days used in residentRepo: %v", err)
		}

		err = s.carRepo.AddToAmtParkingDaysUsed(desiredPermit.CarID, permitLength)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error adding to amt parking days used in carRepo: %v", err)
		}
	}

	permitID, err := s.permitRepo.Create(desiredPermit)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error create new permit in permitRepo: %v", err)
	}

	newPermit, err := s.permitRepo.GetOne(permitID)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error getting permit after having created it in permitRepo: %v", err)
	}

	return newPermit, nil
}
