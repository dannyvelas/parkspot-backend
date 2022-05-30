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

func getAll(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		reversed := toBool(r.URL.Query().Get("reversed"))

		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		allPermits, err := permitRepo.GetAll(boundedLimit, offset, reversed)
		if err != nil {
			log.Error().Msgf("permit_router.getAll: Error getting permits: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := permitRepo.GetAllTotalAmount()
		if err != nil {
			log.Error().Msgf("permit_router.getAll: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(allPermits, totalAmount)

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func getActive(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		reversed := toBool(r.URL.Query().Get("reversed"))

		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		activePermits, err := permitRepo.GetActive(boundedLimit, offset, reversed)
		if err != nil {
			log.Error().Msgf("permit_router.getActive: Error querying permitRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := permitRepo.GetActiveTotalAmount()
		if err != nil {
			log.Error().Msgf("permit_router.getActive: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(activePermits, totalAmount)

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func getExceptions(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		reversed := toBool(r.URL.Query().Get("reversed"))

		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		exceptionPermits, err := permitRepo.GetExceptions(boundedLimit, offset, reversed)
		if err != nil {
			log.Error().Msgf("permit_router.getExceptions: Error querying permitRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := permitRepo.GetExceptionsTotalAmount()
		if err != nil {
			log.Error().Msgf("permit_router.getExceptions: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(exceptionPermits, totalAmount)

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func getExpired(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		reversed := toBool(r.URL.Query().Get("reversed"))

		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		window := toPosInt(r.URL.Query().Get("window"))
		if window == 0 {
			window = defaultExpirationWindow
		}

		expiredPermits, err := permitRepo.GetExpired(boundedLimit, offset, reversed, int(window))
		if err != nil {
			log.Error().Msgf("permit_router.getExpired: Error querying permitRepo: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(expiredPermits, len(expiredPermits))

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func getOne(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := toPosInt(chi.URLParam(r, "id"))

		permit, err := permitRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("permit_router.getOne: Error getting permit: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, permit)
	}
}

func create(permitRepo storage.PermitRepo, carRepo storage.CarRepo, residentRepo storage.ResidentRepo, dateFormat string) http.HandlerFunc {
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
		user, err := ctxGetUser(ctx)
		if err != nil {
			log.Error().Msgf("permit_router.create: error getting user: %v", err)
			respondInternalError(w)
			return
		}

		if user.Role == ResidentRole && newPermitReq.ExceptionReason != "" {
			message := "Residents cannot request parking permits with exceptions"
			respondError(w, newErrBadRequest(message))
			return
		}

		// check if car exists
		existingCar, err := carRepo.GetByLicensePlate(newPermitReq.NewCarReq.LicensePlate)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router.create: Error querying carRepo: %v", err)
			respondInternalError(w)
			return
		}

		// error out if car exists and has active permits during dates requested
		if existingCar != (models.Car{}) { // car exists
			activePermitsDuring, err := permitRepo.GetActiveOfCarDuring(
				existingCar.Id, newPermitReq.StartDate, newPermitReq.EndDate)
			if err != nil {
				log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
				respondInternalError(w)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
					"because this car has at least one active permit during that time.",
					newPermitReq.StartDate.Format(dateFormat),
					newPermitReq.EndDate.Format(dateFormat))
				respondError(w, newErrBadRequest(message))
				return
			}
		}

		// error out if resident DNE
		existingResident, err := residentRepo.GetOne(newPermitReq.ResidentId)
		if err != nil && !errors.Is(err, storage.ErrNoRows) { // unexpected error
			log.Error().Msgf("permit_router.create: Error querying residentRepo: %v", err)
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
				log.Error().Msgf("permit_router.create: Error querying permitRepo: %v", err)
				respondInternalError(w)
				return
			} else if len(activePermitsDuring) != 0 {
				message := fmt.Sprintf("Cannot create a permit during dates %s and %s, "+
					"because this resident has at least one active permit during that time.",
					newPermitReq.StartDate.Format(dateFormat),
					newPermitReq.EndDate.Format(dateFormat))
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

		// get or create car
		var permitCar models.Car
		if existingCar != (models.Car{}) {
			permitCar = existingCar
		} else {
			newCarArgs := newPermitReq.NewCarReq.toNewCarArgs()
			carId, err := carRepo.Create(newCarArgs)
			if err != nil {
				log.Error().Msgf("permit_router.create: Error querying carRepo: %v", err)
				respondInternalError(w)
				return
			}

			permitCar = newCarArgs.ToCar(carId)
		}

		err = residentRepo.AddToAmtParkingDaysUsed(existingResident.Id, permitLength)
		if err != nil {
			log.Error().Msgf("permit_router.create: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		err = carRepo.AddToAmtParkingDaysUsed(permitCar.Id, permitLength)
		if err != nil {
			log.Error().Msgf("permit_router.create: Error querying carRepo: %v", err)
			respondInternalError(w)
			return
		}

		affectsDays := newPermitReq.ExceptionReason != "" || existingResident.UnlimDays
		newPermitArgs := newPermitReq.toNewPermitArgs(permitCar.Id, affectsDays)
		permitId, err := permitRepo.Create(newPermitArgs)
		if err != nil {
			log.Error().Msgf("permit_router.create: Error querying carRepo: %v", err)
			respondInternalError(w)
			return
		}

		newPermit, err := permitRepo.GetOne(permitId)
		if err != nil {
			log.Error().Msgf("permit_router.create: Error getting permit: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, newPermit)
	}
}

func deletePermit(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := toPosInt(chi.URLParam(r, "id"))

		err := permitRepo.Delete(id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("permit_router.deletePermit: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, emptyResponse{Ok: true})
	}
}

func searchPermits(permitRepo storage.PermitRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		searchStr := r.URL.Query().Get("search")
		if searchStr == "" {
			respondJSON(w, http.StatusOK, []models.Permit{})
			return
		}

		listType := r.URL.Query().Get("listType")
		permitFilter, err := storage.NewPermitFilter(listType)
		if err != nil {
			respondError(w, newErrBadRequest("invalid listType value"))
		}

		permits, err := permitRepo.Search(searchStr, permitFilter)
		if err != nil {
			log.Error().Msgf("permit_router.searchPermits: Error getting permits: %v", err)
			respondInternalError(w)
			return
		}

		permitsWithMetadata := newListWithMetadata(permits, len(permits))

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}
