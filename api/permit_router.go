package api

import (
	"encoding/json"
	"errors"
	"fmt"
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
		var createPermitReq createPermitReq
		if err := json.NewDecoder(r.Body).Decode(&createPermitReq); err != nil {
			respondError(w, errBadRequest)
			return
		}

		// change request body to a validated `models.CreatePermit` datatype
		createPermit, err := createPermitReq.toModels()
		if err != nil {
			respondErrorWith(w, errBadRequest, err.Error())
			return
		}

		{ // error out if car exists and has active permits during dates requested
			existingCar, err := carRepo.GetByLicensePlate(createPermit.CreateCar.LicensePlate)
			if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
				log.Error().Msgf("permit_router: Error querying carRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			}

			if err == nil { // car exists
				activePermitsDuring, err := permitRepo.GetActiveOfCarDuring(
					existingCar.Id, createPermit.StartDate, createPermit.EndDate)
				if err != nil {
					log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
					respondError(w, errInternalServerError)
					return
				} else if len(activePermitsDuring) != 0 {
					message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
						"because this car has at least one active permit during that time.",
						createPermit.StartDate.Format(dateFormat),
						createPermit.EndDate.Format(dateFormat))
					respondErrorWith(w, errBadRequest, message)
					return
				}
			}
		}

		// error out if resident DNE
		existingResident, err := residentRepo.GetOne(createPermit.ResidentId)
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

		if createPermit.ExceptionReason == nil { // if not exception
			permitLength := int(createPermit.EndDate.Sub(createPermit.StartDate).Hours() / 24)

			if permitLength > maxPermitLength {
				message := fmt.Sprintf("Error: Requests cannot be longer than %d days, unless there is an exception."+
					" If this resident wants their guest to park for more than %d days, they can request"+
					" %d days of parking and apply for another request once that one expires.", maxPermitLength, maxPermitLength, maxPermitLength)
				respondErrorWith(w, errBadRequest, message)
				return
			}

			activePermitsDuring, err := permitRepo.GetActiveOfResidentDuring(
				existingResident.Id, createPermit.StartDate, createPermit.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
				respondError(w, errInternalServerError)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
					"because this resident has at least one active permit during that time.",
					createPermit.StartDate.Format(dateFormat),
					createPermit.EndDate.Format(dateFormat))
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
				}

				if existingResident.AmtParkingDaysUsed+permitLength > 20 {
					message := fmt.Sprintf("Error: This request would exceed the resident's yearly guest parking pass limit of %d days."+
						"\nThis resident has given out parking permits for a total of %d days."+
						"\nThis resident can give out max %d more day(s) before reaching their limit."+
						"\nThis resident can give only give more if they have unlimited days or if their requested permites are"+
						" exceptions", maxParkingDays, existingResident.AmtParkingDaysUsed, maxParkingDays-existingResident.AmtParkingDaysUsed)
					respondErrorWith(w, errBadRequest, message)
					return
				}

				// TODO: check licensePlate stay
			}
		}

		respondJSON(w, 200, createPermit)
	}
}
