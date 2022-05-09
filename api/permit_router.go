package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PermitRouter(permitRepo storage.PermitRepo, carRepo storage.CarRepo, residentRepo storage.ResidentRepo, dateFormat string) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/active", getActive(permitRepo))
		r.Get("/all", getAll(permitRepo))
		r.Post("/create", create(permitRepo, carRepo, residentRepo, dateFormat))
	}
}

func getActive(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		activePermits, err := permitRepo.GetActive(boundedSize, offset)
		if err != nil {
			log.Error().Msgf("permit_router.GetActive: Error querying permitRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, activePermits)
	}
}

func getAll(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		size := toUint(r.URL.Query().Get("size"))
		page := toUint(r.URL.Query().Get("page"))
		boundedSize, offset := getBoundedSizeAndOffset(size, page)

		allPermits, err := permitRepo.GetAll(boundedSize, offset)
		if err != nil {
			log.Error().Msgf("permit_router.getAll: Error querying permitRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, allPermits)
	}
}

func create(permitRepo storage.PermitRepo, carRepo storage.CarRepo, residentRepo storage.ResidentRepo, dateFormat string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newPermitReq newPermitReq
		if err := json.NewDecoder(r.Body).Decode(&newPermitReq); err != nil {
			respondError(w, errBadRequest)
			return
		}

		if err := newPermitReq.validate(); err != nil {
			respondErrorWith(w, errBadRequest, err.Error())
			return
		}

		// check if car exists
		existingCar, err := carRepo.GetByLicensePlate(newPermitReq.NewCarReq.LicensePlate)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		// error out if car exists and has active permits during dates requested
		if existingCar != (models.Car{}) { // car exists
			activePermitsDuring, err := permitRepo.GetActiveOfCarDuring(
				existingCar.Id, newPermitReq.StartDate, newPermitReq.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
					"because this car has at least one active permit during that time.",
					newPermitReq.StartDate.Format(dateFormat),
					newPermitReq.EndDate.Format(dateFormat))
				respondErrorWith(w, errBadRequest, message)
				return
			}
		}

		// error out if resident DNE
		existingResident, err := residentRepo.GetOne(newPermitReq.ResidentId)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router: Error querying residentRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		} else if errors.Is(err, storage.ErrNoRows) { // resident does not exist
			message := "Users must have a registered account to request a guest parking permit." +
				" Please create their account before requesting their permit."
			respondErrorWith(w, errBadRequest, message)
			return
		}

		permitLength := int(newPermitReq.EndDate.Sub(newPermitReq.StartDate).Hours() / 24)
		if newPermitReq.ExceptionReason == "" { // if not exception
			if permitLength > maxPermitLength {
				message := fmt.Sprintf("Error: Requests cannot be longer than %d days, unless there is an exception."+
					" If this resident wants their guest to park for more than %d days, they can request"+
					" %d days of parking and apply for another request once that one expires.", maxPermitLength, maxPermitLength, maxPermitLength)
				respondErrorWith(w, errBadRequest, message)
				return
			}

			activePermitsDuring, err := permitRepo.GetActiveOfResidentDuring(
				existingResident.Id, newPermitReq.StartDate, newPermitReq.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
					"because this resident has at least one active permit during that time.",
					newPermitReq.StartDate.Format(dateFormat),
					newPermitReq.EndDate.Format(dateFormat))
				respondErrorWith(w, errBadRequest, message)
				return
			}

			if !existingResident.UnlimDays {
				if existingResident.AmtParkingDaysUsed >= maxParkingDays {
					message := fmt.Sprintf("Error: This resident has given out parking passes that have lasted a combined total of"+
						" %d days or more."+
						"\nResidents are allowed maximum %d days of parking passes, unless there is an exception."+
						"\nThis resident must wait until next year to give out new parking passes.", maxParkingDays, maxParkingDays)
					respondErrorWith(w, errBadRequest, message)
					return
				} else if existingResident.AmtParkingDaysUsed+permitLength > maxParkingDays {
					message := fmt.Sprintf("Error: This request would exceed the resident's yearly guest parking pass limit of %d days."+
						"\nThis resident has given out parking permits for a total of %d days."+
						"\nThis resident can give out max %d more day(s) before reaching their limit."+
						"\nThis resident only give more permits if they have unlimited days or if their requested permites are"+
						" exceptions", maxParkingDays, existingResident.AmtParkingDaysUsed, maxParkingDays-existingResident.AmtParkingDaysUsed)
					respondErrorWith(w, errBadRequest, message)
					return
				}

				if existingCar.AmtParkingDaysUsed >= maxParkingDays {
					message := fmt.Sprintf("Error: This car has had a combined total of %d parking days or more."+
						"\nEach car is allowed maximum %d days of parking, unless there is an exception."+
						"\nThis car must wait until next year to get a new parking permit.", maxParkingDays, maxParkingDays)
					respondErrorWith(w, errBadRequest, message)
					return
				} else if existingCar.AmtParkingDaysUsed+permitLength > maxParkingDays {
					message := fmt.Sprintf("Error: This request would exceed this car's yearly parking permit limit of %d days."+
						"\nThis car has received parking permits for a total of %d days."+
						"\nThis car can receive %d more day(s) before reaching its limit."+
						"\nThis resident can give only give more permits if they have unlimited days or if their requested permites are"+
						" exceptions", maxParkingDays, existingCar.AmtParkingDaysUsed, maxParkingDays-existingCar.AmtParkingDaysUsed)
					respondErrorWith(w, errBadRequest, message)
					return
				}
			}
		}

		// checks successful: proceed to create permit

		// get or create car
		var permitCar models.Car
		if existingCar != (models.Car{}) {
			permitCar = existingCar
		} else {
			newCarArgs := newPermitReq.NewCarReq.toNewCarArgs()
			carId, err := carRepo.Create(newCarArgs)
			if err != nil {
				log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			}

			permitCar = newCarArgs.ToCar(carId)
		}

		err = residentRepo.AddToAmtParkingDaysUsed(existingResident.Id, permitLength)
		if err != nil {
			log.Error().Msgf("permit_router: Error querying residentRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		err = carRepo.AddToAmtParkingDaysUsed(permitCar.Id, permitLength)
		if err != nil {
			log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		newPermitArgs := newPermitReq.toNewPermitArgs(permitCar.Id)
		permitId, err := permitRepo.Create(newPermitArgs)
		if err != nil {
			log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
			respondError(w, errInternalServerError)
			return
		}

		newPermit := newPermitArgs.ToPermit(permitId, permitCar)

		respondJSON(w, 200, newPermit)
	}
}
