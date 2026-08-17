package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"family-shopping-list-api/internal/list/service"
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
	var input service.CreateInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	list, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "create_list_failed", err.Error())
		return
	}
	response.Created(w, list)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := pathID(w, r)
	if id == 0 {
		return
	}
	list, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleListError(w, err)
		return
	}
	response.OK(w, list)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	lists, err := h.service.ListByUser(context.Background(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "list_lists_failed", err.Error())
		return
	}
	response.OK(w, lists)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathID(w, r)
	if id == 0 {
		return
	}
	var input service.UpdateInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	list, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		handleListError(w, err)
		return
	}
	response.OK(w, list)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathID(w, r)
	if id == 0 {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		handleListError(w, err)
		return
	}
	response.OK(w, map[string]bool{"deleted": true})
}

func pathID(w http.ResponseWriter, r *http.Request) uint64 {
	id, err := strconv.ParseUint(r.PathValue("list_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "无效的清单 ID")
		return 0
	}
	return id
}

func handleListError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "list_not_found", "清单不存在")
		return
	}
	response.Error(w, http.StatusInternalServerError, "list_operation_failed", err.Error())
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
