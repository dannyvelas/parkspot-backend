package app

import (
	"errors"
	"fmt"
	"github.com/dannyvelas/parkspot-backend/config"
	"github.com/dannyvelas/parkspot-backend/errs"
	"github.com/dannyvelas/parkspot-backend/models"
	"github.com/dannyvelas/parkspot-backend/models/validator"
	"github.com/dannyvelas/parkspot-backend/storage"
	"github.com/dannyvelas/parkspot-backend/storage/selectopts"
	"github.com/dannyvelas/parkspot-backend/util"
	"strings"
)

type PermitService struct {
	permitRepo   storage.PermitRepo
	residentRepo storage.ResidentRepo
	carService   CarService
}

func NewPermitService(permitRepo storage.PermitRepo, residentRepo storage.ResidentRepo, carService CarService) PermitService {
	return PermitService{
		permitRepo:   permitRepo,
		residentRepo: residentRepo,
		carService:   carService,
	}
}

func (s PermitService) GetAll(status models.Status, limit, page int, reversed bool, search, residentID string) (models.ListWithMetadata[models.Permit], error) {
	boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

	allPermits, err := s.permitRepo.SelectWhere(models.Permit{ResidentID: residentID},
		selectopts.WithStatus(status),
		selectopts.WithLimitAndOffset(boundedLimit, offset),
		selectopts.WithReversed(reversed),
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting permits from permit repo: %v", err)
	}

	totalAmount, err := s.permitRepo.SelectCountWhere(models.Permit{ResidentID: residentID},
		selectopts.WithStatus(status),
		selectopts.WithSearch(search),
	)
	if err != nil {
		return models.ListWithMetadata[models.Permit]{}, fmt.Errorf("error getting total amount from permit repo: %v", err)
	}

	return models.NewListWithMetadata(allPermits, totalAmount), nil
}

func (s PermitService) GetOne(id int) (models.Permit, error) {
	if id == 0 {
		return models.Permit{}, errs.MissingIDField
	}

	return s.permitRepo.GetOne(id)
}

func (s PermitService) Delete(id int) error {
	if id == 0 {
		return errs.MissingIDField
	}

	permit, err := s.permitRepo.GetOne(id)
	if err != nil {
		return err
	}

	if err = s.permitRepo.Delete(id); err != nil {
		return err
	}

	permitLength := int(permit.EndDate.Sub(permit.StartDate).Hours() / 24)
	if permit.AffectsDays {
		err = s.residentRepo.AddToAmtParkingDaysUsed(permit.ResidentID, -permitLength)
		if err != nil {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in residentRepo: %v", err)
		}

		err = s.carService.carRepo.AddToAmtParkingDaysUsed(permit.CarID, -permitLength)
		if err != nil && !errors.Is(err, errs.NotFound) {
			return fmt.Errorf("error subtracting amtParkingDaysUsed in carRepo: %v", err)
		}
		// purposely not returning error if error.Is(err, errs.NotFound)
		// its possible that the car with id of permit.CarID was deleted and no longer exists.
		// unlike a resident deletion, a car deletion does not cascade delete its permits
	}

	return nil
}

func (s PermitService) Create(desiredPermit models.Permit) (models.Permit, error) {
	if err := s.validateDates(desiredPermit); err != nil {
		return models.Permit{}, err
	}

	permitLength := util.GetAmtDays(desiredPermit.StartDate, desiredPermit.EndDate)
	resident, err := s.getAndValidateResident(desiredPermit, permitLength)
	if err != nil {
		return models.Permit{}, err
	}

	// if carID != "", this permit will be created for a pre-existing car
	if desiredPermit.CarID != "" {
		// find and validate that this car follows policy for creating a permit
		car, err := s.getAndValidateCar(desiredPermit, *resident.UnlimDays, permitLength)
		if err != nil {
			return models.Permit{}, err
		}

		// get a snapshot of car and save it into the permit
		desiredPermit.LicensePlate = car.LicensePlate
		desiredPermit.Color = car.Color
		desiredPermit.Make = car.Make
		desiredPermit.Model = car.Model
	} else {
		// otherwise, we will create a new car for this permit
		desiredCar := models.Car{ResidentID: desiredPermit.ResidentID, LicensePlate: desiredPermit.LicensePlate, Color: desiredPermit.Color, Make: desiredPermit.Make, Model: desiredPermit.Model}
		createdCar, err := s.carService.Create(desiredCar)
		if err != nil {
			return models.Permit{}, fmt.Errorf("error creating car: %w", err)
		}

		// record the carID that was used to create this car
		desiredPermit.CarID = createdCar.ID
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
	{
		car := models.Car{LicensePlate: updatedFields.LicensePlate, Color: updatedFields.Color, Make: updatedFields.Make, Model: updatedFields.Model}
		if err := validator.EditCar.Run(car); err != nil {
			return models.Permit{}, err
		}
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

	residents, err := s.residentRepo.SelectWhere(models.Resident{ID: desiredPermit.ResidentID})
	if err != nil {
		return models.Resident{}, fmt.Errorf("error getting one from resident repo: %v", err)
	} else if len(residents) == 0 {
		return models.Resident{}, errs.ResidentForPermitDNE
	}
	resident := residents[0]

	// if this is an exception, there are no more checks to be performed. so return no errors
	if desiredPermit.ExceptionReason != "" {
		return resident, nil
	}

	if permitLength > config.MaxPermitLength {
		return models.Resident{}, errs.PermitTooLong
	}

	residentActivePermitsDuring, err := s.permitRepo.SelectWhere(models.Permit{ResidentID: resident.ID},
		selectopts.WithDateIntersect(desiredPermit.StartDate, desiredPermit.EndDate),
	)
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

	existingCar, err := s.carService.GetOne(desiredPermit.CarID)
	if errors.Is(err, errs.NotFound) {
		return models.Car{}, errs.CarForPermitDNE
	} else if err != nil {
		return models.Car{}, fmt.Errorf("error getting one from carRepo: %v", err)
	}

	// we found the car: error out if it has active permits during dates requested
	carActivePermitsDuring, err := s.permitRepo.SelectWhere(models.Permit{CarID: existingCar.ID},
		selectopts.WithDateIntersect(desiredPermit.StartDate, desiredPermit.EndDate),
	)
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
	var errors []string

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

		err = s.carService.carRepo.AddToAmtParkingDaysUsed(desiredPermit.CarID, permitLength)
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
