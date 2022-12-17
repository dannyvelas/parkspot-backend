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

func getPermits(permitRepo storage.PermitRepo, permitFilter models.PermitFilter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		reversed := toBool(r.URL.Query().Get("reversed"))
		search := r.URL.Query().Get("search")
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		ctx := r.Context()
		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("visitor_router.getVisitorsOfResident: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if AccessPayload.Role == models.ResidentRole {
			residentID = AccessPayload.Id
		}

		allPermits, err := permitRepo.Get(permitFilter, residentID, boundedLimit, offset, reversed, search)
		if err != nil {
			log.Error().Msgf("permit_router.getPermits: Error getting permits: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := permitRepo.GetCount(permitFilter, residentID)
		if err != nil {
			log.Error().Msgf("permit_router.getPermits: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(allPermits, totalAmount)

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func getOnePermit(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := toPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		permit, err := permitRepo.GetOne(id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("permit_router.getOne: Error getting permit: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, permit)
	}
}

func createPermit(permitRepo storage.PermitRepo, residentRepo storage.ResidentRepo, carRepo storage.CarRepo, dateFormat string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newPermitReq newPermitReq
		if err := json.NewDecoder(r.Body).Decode(&newPermitReq); err != nil {
			respondError(w, newErrMalformed("NewPermitReq"))
			return
		}

		if err := newPermitReq.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		ctx := r.Context()
		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("permit_router.createPermit: error getting access payload: %v", err)
			respondInternalError(w)
			return
		}

		if AccessPayload.Role == models.ResidentRole && newPermitReq.ExceptionReason != "" {
			message := "Residents cannot request parking permits with exceptions"
			respondError(w, newErrBadRequest(message))
			return
		}

		// check if car exists
		existingCar, err := carRepo.GetByLicensePlate(newPermitReq.Car.LicensePlate)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router.createPermit: Error querying carRepo: %v", err)
			respondInternalError(w)
			return
		}

		// error out if car exists and has active permits during dates requested
		if existingCar != (models.Car{}) { // car exists
			activePermitsDuring, err := permitRepo.GetActiveOfCarDuring(
				existingCar.Id, newPermitReq.StartDate, newPermitReq.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.createPermit: Error querying permitRepo: %v", err)
				respondInternalError(w)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during these dates" +
					" because this car has at least one active permit during that time.")
				respondError(w, newErrBadRequest(message))
				return
			}
		}

		// error out if resident DNE
		existingResident, err := residentRepo.GetOne(newPermitReq.ResidentId)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router.createPermit: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, storage.ErrNoRows) { // resident does not exist
			message := "Users must have a registered account to request a guest parking permit." +
				" Please create their account before requesting their permit."
			respondError(w, newErrBadRequest(message))
			return
		}

		permitLength := int(newPermitReq.EndDate.Sub(newPermitReq.StartDate).Hours() / 24)
		if newPermitReq.ExceptionReason == "" { // if not exception
			if permitLength > maxPermitLength {
				message := fmt.Sprintf("Error: Requests cannot be longer than %d days, unless there is an exception."+
					" If this resident wants their guest to park for more than %d days, they can request"+
					" %d days of parking and apply for another request once that one expires.", maxPermitLength, maxPermitLength, maxPermitLength)
				respondError(w, newErrBadRequest(message))
				return
			}

			activePermitsDuring, err := permitRepo.GetActiveOfResidentDuring(
				existingResident.Id, newPermitReq.StartDate, newPermitReq.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.createPermit: Error querying permitRepo: %v", err)
				respondInternalError(w)
				return
			} else if len(activePermitsDuring) >= 2 {
				message := fmt.Sprintf("Cannot create a permit during these dates" +
					" because this resident has at least two active permits during that time.")
				respondError(w, newErrBadRequest(message))
				return
			}

			if !existingResident.UnlimDays {
				if existingResident.AmtParkingDaysUsed >= maxParkingDays {
					message := fmt.Sprintf("Error: This resident has given out parking passes that have lasted a combined total of"+
						" %d days or more."+
						"\nResidents are allowed maximum %d days of parking passes, unless there is an exception."+
						"\nThis resident must wait until next year to give out new parking passes.", maxParkingDays, maxParkingDays)
					respondError(w, newErrBadRequest(message))
					return
				} else if existingResident.AmtParkingDaysUsed+permitLength > maxParkingDays {
					message := fmt.Sprintf("Error: This request would exceed the resident's yearly guest parking pass limit of %d days."+
						"\nThis resident has given out parking permits for a total of %d days."+
						"\nThis resident can give out max %d more day(s) before reaching their limit."+
						"\nThis resident can only give more permits if they have unlimited days or if their requested permites are"+
						" exceptions", maxParkingDays, existingResident.AmtParkingDaysUsed, maxParkingDays-existingResident.AmtParkingDaysUsed)
					respondError(w, newErrBadRequest(message))
					return
				}

				if existingCar.AmtParkingDaysUsed >= maxParkingDays {
					message := fmt.Sprintf("Error: This car has had a combined total of %d parking days or more."+
						"\nEach car is allowed maximum %d days of parking, unless there is an exception."+
						"\nThis car must wait until next year to get a new parking permit.", maxParkingDays, maxParkingDays)
					respondError(w, newErrBadRequest(message))
					return
				} else if existingCar.AmtParkingDaysUsed+permitLength > maxParkingDays {
					message := fmt.Sprintf("Error: This request would exceed this car's yearly parking permit limit of %d days."+
						"\nThis car has received parking permits for a total of %d days."+
						"\nThis car can receive %d more day(s) before reaching its limit."+
						"\nThis resident can give only give more permits if they have unlimited days or if their requested permites are"+
						" exceptions", maxParkingDays, existingCar.AmtParkingDaysUsed, maxParkingDays-existingCar.AmtParkingDaysUsed)
					respondError(w, newErrBadRequest(message))
					return
				}
			}
		}

		// checks successful: proceed to create permit

		permitCar, err := getOrCreateCar(carRepo, existingCar, newPermitReq.Car)
		if err != nil {
			log.Error().Msgf("permit_router.createPermit: %v", err)
			respondInternalError(w)
		}

		affectsDays := newPermitReq.ExceptionReason == "" && !existingResident.UnlimDays
		if affectsDays {
			err = residentRepo.AddToAmtParkingDaysUsed(existingResident.Id, permitLength)
			if err != nil {
				log.Error().Msgf("permit_router.createPermit: Error querying residentRepo: %v", err)
				respondInternalError(w)
				return
			}

			err = carRepo.AddToAmtParkingDaysUsed(permitCar.Id, permitLength)
			if err != nil {
				log.Error().Msgf("permit_router.createPermit: Error querying carRepo: %v", err)
				respondInternalError(w)
				return
			}
		}

		permitId, err := permitRepo.Create(
			newPermitReq.ResidentId,
			permitCar.Id,
			newPermitReq.StartDate.Unix(),
			newPermitReq.EndDate.Unix(),
			affectsDays,
			newPermitReq.ExceptionReason,
		)
		if err != nil {
			log.Error().Msgf("permit_router.createPermit: Error querying permitRepo: %v", err)
			respondInternalError(w)
			return
		}

		newPermit, err := permitRepo.GetOne(permitId)
		if err != nil {
			log.Error().Msgf("permit_router.createPermit: Error getting permit: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, newPermit)
	}
}

func deletePermit(permitRepo storage.PermitRepo, residentRepo storage.ResidentRepo, carRepo storage.CarRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := toPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		permit, err := permitRepo.GetOne(id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("permit_router.deletePermit: Error getting permit: %v", err)
			respondInternalError(w)
			return
		}

		err = permitRepo.Delete(permit.Id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("permit_router.deletePermit: %v", err)
			respondInternalError(w)
			return
		}

		permitLength := int(permit.EndDate.Sub(permit.StartDate).Hours() / 24)
		if permit.AffectsDays {
			err = residentRepo.AddToAmtParkingDaysUsed(permit.ResidentId, -permitLength)
			if err != nil {
				log.Error().Msgf("permit_router.deletePermit: Error querying residentRepo: %v", err)
				respondInternalError(w)
				return
			}

			err = carRepo.AddToAmtParkingDaysUsed(permit.Car.Id, -permitLength)
			if err != nil {
				log.Error().Msgf("permit_router.deletePermit: Error querying carRepo: %v", err)
				respondInternalError(w)
				return
			}
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted permit"})
	}
}

// helpers
func getOrCreateCar(
	carRepo storage.CarRepo,
	existingCar models.Car,
	newCarReq newCarReq,
) (models.Car, error) {
	// car exists and has all fields
	if existingCar != (models.Car{}) && existingCar.Make != "" && existingCar.Model != "" {
		return existingCar, nil
	}

	// car exits but missing fields
	if existingCar != (models.Car{}) {
		err := carRepo.Update(existingCar.Id, newCarReq.Color, newCarReq.Make, newCarReq.Model)
		if err != nil {
			return models.Car{}, fmt.Errorf("Error updating car: %v", err)
		}
		return models.NewCar(existingCar.Id,
			newCarReq.LicensePlate,
			newCarReq.Color,
			newCarReq.Make,
			newCarReq.Model,
			existingCar.AmtParkingDaysUsed), nil
	}

	// car DNE
	carId, err := carRepo.Create(newCarReq.LicensePlate, newCarReq.Color, newCarReq.Make, newCarReq.Model)
	if err != nil {
		return models.Car{}, fmt.Errorf("Error creating car: %v", err)
	}

	return models.NewCar(
		carId,
		newCarReq.LicensePlate,
		newCarReq.Color,
		newCarReq.Make,
		newCarReq.Model,
		0), nil
}
