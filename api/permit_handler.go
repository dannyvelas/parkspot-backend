package api

import (
	"encoding/json"
	"errors"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/util"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
)

type permitHandler struct {
	permitService   app.PermitService
	residentService app.ResidentService
	carService      app.CarService
}

func newPermitHandler(permitService app.PermitService, residentService app.ResidentService, carService app.CarService) permitHandler {
	return permitHandler{
		permitService:   permitService,
		residentService: residentService,
		carService:      carService,
	}
}

func (h permitHandler) get(permitFilter models.PermitFilter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		reversed := util.ToBool(r.URL.Query().Get("reversed"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("error getting access payload: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if accessPayload.Role == models.ResidentRole {
			residentID = accessPayload.ID
		}

		permitsWithMetadata, err := h.permitService.GetAll(permitFilter, limit, page, reversed, search, residentID)
		if err != nil {
			log.Error().Msgf("error getting permits with metadata: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, permitsWithMetadata)
	}
}

func (h permitHandler) getOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := util.ToPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		permit, err := h.permitService.GetOne(id)
		if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("error getting one permit in permit service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, permit)
	}
}

func (h permitHandler) create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newPermitReq models.Permit
		if err := json.NewDecoder(r.Body).Decode(&newPermitReq); err != nil {
			respondError(w, newErrMalformed("NewPermitReq"))
			return
		}

		if err := newPermitReq.ValidateCreation(); err != nil {
			respondError(w, newErrBadRequest(err.Error()))
			return
		}

		ctx := r.Context()
		accessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("permit_router.createPermit: error getting access payload: %v", err)
			respondInternalError(w)
			return
		}

		if accessPayload.Role == models.ResidentRole && newPermitReq.ExceptionReason != "" {
			message := "Residents cannot request parking permits with exceptions"
			respondError(w, newErrBadRequest(message))
			return
		}

		// error out if resident DNE
		existingResident, err := h.residentService.GetOne(newPermitReq.ResidentID)
		if err != nil && !errors.Is(err, app.ErrNotFound) { // unexpected error
			log.Error().Msgf("error getting one from residentService: %v", err)
			respondInternalError(w)
			return
		} else if errors.Is(err, app.ErrNotFound) { // resident does not exist
			message := "Users must have a registered account to request a guest" +
				" parking permit. Please create their account before requesting their permit."
			respondError(w, newErrBadRequest(message))
			return
		}

		// TODO: perhaps in ValidateCreation or perhaps above:
		// check that the car for which we are creating the permit
		// already exists. do this by checking the carID

		err = h.permitService.ValidateCreation(newPermitReq, existingResident)
		var createPermitErr app.CreatePermitError
		if errors.As(err, &createPermitErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("error validating permit creation in permitservice: %v", err)
			respondInternalError(w)
			return
		}

		createdPermit, err := h.permitService.Create(newPermitReq, existingResident)
		if err != nil {
			log.Error().Msgf("error creating permit in permitservice: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, createdPermit)
	}
}

func (h permitHandler) deleteOne() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := util.ToPosInt(chi.URLParam(r, "id"))
		if id == 0 {
			respondError(w, newErrBadRequest("id parameter cannot be empty"))
			return
		}

		err := h.permitService.Delete(id)
		if errors.Is(err, app.ErrNotFound) {
			respondError(w, newErrNotFound("permit"))
			return
		} else if err != nil {
			log.Error().Msgf("error deleting permit in permit service: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, message{"Successfully deleted permit"})
	}
}
