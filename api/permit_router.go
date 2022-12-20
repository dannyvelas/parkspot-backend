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

type PermitHandler struct {
	permitService app.PermitService
}

func NewPermitHandler(permitService app.PermitService) PermitHandler {
	return PermitHandler{
		permitService: permitService,
	}
}

func (h PermitHandler) Get(permitFilter models.PermitFilter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := util.ToPosInt(r.URL.Query().Get("limit"))
		page := util.ToPosInt(r.URL.Query().Get("page"))
		reversed := util.ToBool(r.URL.Query().Get("reversed"))
		search := r.URL.Query().Get("search")

		ctx := r.Context()
		AccessPayload, err := ctxGetAccessPayload(ctx)
		if err != nil {
			log.Error().Msgf("error getting access payload: %v", err)
			respondInternalError(w)
			return
		}

		residentID := ""
		if AccessPayload.Role == models.ResidentRole {
			residentID = AccessPayload.ID
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

func (h PermitHandler) GetOne() http.HandlerFunc {
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

func (h PermitHandler) Create() http.HandlerFunc {
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

		desiredPermit := models.CreatePermit{
			ResidentID:      newPermitReq.ResidentID,
			StartDate:       newPermitReq.StartDate.Unix(),
			EndDate:         newPermitReq.EndDate.Unix(),
			ExceptionReason: newPermitReq.ExceptionReason,
		}
		desiredCar := models.CreateCar(newPermitReq.Car)

		createdPermit, err := h.permitService.Create(desiredPermit, desiredCar)
		var createPermitErr app.CreatePermitError
		if errors.As(err, &createPermitErr) {
			respondError(w, newErrBadRequest(err.Error()))
			return
		} else if err != nil {
			log.Error().Msgf("error creating permit in permitservice: %v", err)
			respondInternalError(w)
			return
		}

		respondJSON(w, http.StatusOK, createdPermit)
	}
}

func (h PermitHandler) Delete() http.HandlerFunc {
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
