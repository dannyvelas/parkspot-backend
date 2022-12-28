package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/errs"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/dannyvelas/lasvistas_api/util"
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
		return models.Permit{}, errs.NotFound("permit")
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	return permit, nil
}

func (s PermitService) Delete(permitID int) error {
	permit, err := s.permitRepo.GetOne(permitID)
	if errors.Is(err, storage.ErrNoRows) {
		return errs.NotFound("permit")
	} else if err != nil {
		return fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	err = s.permitRepo.Delete(permitID)
	if errors.Is(err, storage.ErrNoRows) {
		return errs.NotFound("permit")
	} else if err != nil {
		return fmt.Errorf("error getting permit from permit repo: %v", err)
	}

	permitLength := int(permit.EndDate.Sub(permit.StartDate).Hours() / 24)
	if permit.AffectsDays {
		err = s.residentRepo.AddToAmtParkingDaysUsed(permit.ResidentID, -permitLength)
		if err != nil {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in residentRepo: %v", err)
		}

		err = s.carRepo.AddToAmtParkingDaysUsed(permit.CarID, -permitLength)
		if err != nil && !errors.Is(err, storage.ErrNoRows) {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in carRepo: %v", err)
		}
		// purposely not returning error if error.Is(err, storage.ErrNoRows)
		// its possible that the car with id of permit.CarID was deleted and no longer exists.
		// unlike a resident deletion, a car deletion does not cascade delete its permits
	}

	return nil
}

func (s PermitService) ValidateAndCreate(desiredPermit models.Permit) (models.Permit, error) {
	// error out if resident DNE
	existingResident, err := s.residentRepo.GetOne(desiredPermit.ResidentID)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Permit{}, errs.ResidentForPermitDNE
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("error getting one from residnet repo: %v", err)
	}

	// error out if car DNE
	existingCar, err := s.carRepo.GetOne(desiredPermit.CarID)
	if errors.Is(err, storage.ErrNoRows) {
		return models.Permit{}, errs.CarForPermitDNE
	} else if err != nil {
		return models.Permit{}, fmt.Errorf("error getting one from carRepo: %v", err)
	}

	if apiErr := s.validateCreation(desiredPermit, existingResident, existingCar); apiErr != nil {
		return models.Permit{}, apiErr
	}

	// get a snapshot of the car being used and save it into the permit
	desiredPermit.LicensePlate = existingCar.LicensePlate
	desiredPermit.Color = existingCar.Color
	desiredPermit.Make = existingCar.Make
	desiredPermit.Model = existingCar.Model
	desiredPermit.AffectsDays = desiredPermit.ExceptionReason == "" && !*existingResident.UnlimDays

	createdPermit, err := s.create(desiredPermit)
	if err != nil {
		return models.Permit{}, fmt.Errorf("error creating permit in permitservice: %v", err)
	}

	return createdPermit, nil
}

func (s PermitService) validateCreation(desiredPermit models.Permit, existingResident models.Resident, existingCar models.Car) error {
	if apiErr := desiredPermit.ValidateCreation(); apiErr != nil {
		return apiErr
	}

	// error out if car has active permits during dates requested
	carActivePermitsDuring, err := s.permitRepo.GetActiveOfCarDuring(
		existingCar.ID, desiredPermit.StartDate, desiredPermit.EndDate)
	if err != nil {
		return fmt.Errorf("error getting active of car during dates in permitRepo: %v", err)
	} else if len(carActivePermitsDuring) != 0 {
		return errs.CarActivePermit
	}

	// if this is an exception, there are no more checks to be performed. so return no errors
	if desiredPermit.ExceptionReason != "" {
		return nil
	}

	permitLength := util.GetAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
	if permitLength > config.MaxPermitLength {
		return errs.PermitTooLong
	}

	residentActivePermitsDuring, err := s.permitRepo.GetActiveOfResidentDuring(
		existingResident.ID, desiredPermit.StartDate, desiredPermit.EndDate)
	if err != nil {
		return fmt.Errorf("error getting active of resident during dates in permitRepo: %v", err)
	} else if len(residentActivePermitsDuring) >= 2 {
		return errs.ResidentTwoActivePermits
	}

	if existingResident.UnlimDays == nil || existingResident.AmtParkingDaysUsed == nil {
		return fmt.Errorf("data type error in permit service validate create. unlimDays or amtParkingDaysUsed is nil")
	}

	if !*existingResident.UnlimDays {
		if *existingResident.AmtParkingDaysUsed >= config.MaxParkingDays {
			return errs.EntityDaysTooLong("resident", *existingResident.AmtParkingDaysUsed)
		} else if *existingResident.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
			return errs.PermitPlusEntityDaysTooLong("resident", *existingResident.AmtParkingDaysUsed)
		}

		if existingCar.AmtParkingDaysUsed >= config.MaxParkingDays {
			return errs.EntityDaysTooLong("car", existingCar.AmtParkingDaysUsed)
		} else if existingCar.AmtParkingDaysUsed+permitLength > config.MaxParkingDays {
			return errs.PermitPlusEntityDaysTooLong("car", existingCar.AmtParkingDaysUsed)
		}
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
