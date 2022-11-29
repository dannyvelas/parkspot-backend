package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func getAllResidents(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := toPosInt(r.URL.Query().Get("limit"))
		page := toPosInt(r.URL.Query().Get("page"))
		search := r.URL.Query().Get("search")
		boundedLimit, offset := getBoundedLimitAndOffset(limit, page)

		allResidents, err := residentRepo.GetAll(boundedLimit, offset, search)
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		totalAmount, err := residentRepo.GetAllTotalAmount()
		if err != nil {
			log.Error().Msgf("resident_router.getAll: Error getting total amount: %v", err)
			respondInternalError(w)
			return
		}

		residentsWithMetadata := newListWithMetadata(allResidents, totalAmount)

		respondJSON(w, http.StatusOK, residentsWithMetadata)
	}
}

func getOneResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		resident, err := residentRepo.GetOne(id)
		if err != nil && !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("resident_router.getOne: Error getting resident: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("resident"))
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func editResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if err := models.IsResidentId(id); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		var editResidentReq editResidentReq
		if err := json.NewDecoder(r.Body).Decode(&editResidentReq); err != nil {
			respondError(w, newErrMalformed("EditResidentReq"))
			return
		}

		if err := editResidentReq.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		desiredResident := models.EditResident{
			FirstName:          editResidentReq.FirstName,
			LastName:           editResidentReq.LastName,
			Phone:              editResidentReq.Phone,
			Email:              editResidentReq.Email,
			Password:           editResidentReq.Password,
			UnlimDays:          editResidentReq.UnlimDays,
			AmtParkingDaysUsed: editResidentReq.AmtParkingDaysUsed,
		}

		err := residentRepo.Update(id, desiredResident)
		if err != nil {
			log.Error().Msgf("resident_router.editResident: Error updating resident: %v", err)
			respondInternalError(w)
			return
		}

		resident, err := residentRepo.GetOne(id)
		if err != nil {
			log.Error().Msgf("resident_router.editResident: Error getting resident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, resident)
	}
}

func deleteResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		resident, err := residentRepo.GetOne(id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("resident"))
			return
		} else if err != nil {
			log.Error().Msgf("resident_router.deleteResident: Error getting resident: %v", err)
			respondInternalError(w)
			return
		}

		err = residentRepo.Delete(resident.Id)
		if errors.Is(err, storage.ErrNoRows) {
			respondError(w, newErrNotFound("resident"))
			return
		} else if err != nil {
			log.Error().Msgf("resident_router.deleteResident: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted resident"})
	}
}

func createResident(residentRepo storage.ResidentRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload newResidentReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondError(w, newErrMalformed("NewResidentReq"))
			return
		}

		if err := payload.validate(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		if _, err := residentRepo.GetOne(payload.ResidentId); err == nil {
			respondError(w, newErrBadRequest("Resident with this id already exists. Please delete the old account if necessary."))
			return
		} else if !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("auth_router.createResident: error getting resident by id: %v", err)
			respondInternalError(w)
			return
		}

		if _, err := residentRepo.GetOneByEmail(payload.Email); err == nil {
			respondError(w, newErrBadRequest("Resident with this email already exists. Please delete the old account or use a different email."))
			return
		} else if !errors.Is(err, storage.ErrNoRows) {
			log.Error().Msgf("auth_router.createResident error getting resident by email: %v", err)
			respondInternalError(w)
			return
		}

		hashBytes, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Msg("auth_router.createResident: error generating hash:" + err.Error())
			respondInternalError(w)
			return
		}
		hashString := string(hashBytes)

		resident := models.NewResident(payload.ResidentId,
			payload.FirstName,
			payload.LastName,
			payload.Phone,
			payload.Email,
			hashString,
			payload.UnlimDays,
			0, 0)

		err = residentRepo.Create(resident)
		if err != nil {
			log.Error().Msgf("auth_router.createResident: Error querying residentRepo: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Resident successfully created."})
	}
}
