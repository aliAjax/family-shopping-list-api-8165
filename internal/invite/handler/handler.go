package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"family-shopping-list-api/internal/invite/service"
	"family-shopping-list-api/internal/middleware"
	"family-shopping-list-api/pkg/response"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	var input service.CreateInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	invite, err := h.service.Create(r.Context(), userID, listID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "create_invite_failed", err.Error())
		return
	}
	response.Created(w, invite)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	invites, err := h.service.ListByList(r.Context(), listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "list_invites_failed", err.Error())
		return
	}
	response.OK(w, invites)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var input service.JoinInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	input.Code = strings.TrimSpace(input.Code)
	invite, err := h.service.Join(r.Context(), userID, input.Code)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, service.ErrExpired), errors.Is(err, service.ErrInactive):
			status = http.StatusGone
		case errors.Is(err, service.ErrLimitReached):
			status = http.StatusConflict
		}
		response.Error(w, status, "join_list_failed", err.Error())
		return
	}
	response.OK(w, invite)
}

func pathListID(w http.ResponseWriter, r *http.Request) uint64 {
	id, err := strconv.ParseUint(r.PathValue("list_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "无效的清单 ID")
		return 0
	}
	return id
}

func decode(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("请求体不是合法 JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("请求体只能包含一个 JSON 对象")
	}
	return nil
}
