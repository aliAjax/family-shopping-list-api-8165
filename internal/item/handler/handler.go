package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"family-shopping-list-api/internal/item/service"
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
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	var input service.CreateInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	item, err := h.service.Create(r.Context(), listID, userID, input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "create_item_failed", err.Error())
		return
	}
	response.Created(w, item)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	items, err := h.service.List(context.Background(), listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "list_items_failed", err.Error())
		return
	}
	response.OK(w, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	itemID := pathItemID(w, r)
	if itemID == 0 {
		return
	}
	item, err := h.service.Get(r.Context(), listID, itemID)
	if err != nil {
		handleItemError(w, err)
		return
	}
	response.OK(w, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	itemID := pathItemID(w, r)
	if itemID == 0 {
		return
	}
	userID, _ := middleware.UserIDFromContext(r.Context())
	var input service.UpdateInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	item, err := h.service.Update(r.Context(), listID, itemID, userID, input)
	if err != nil {
		handleItemError(w, err)
		return
	}
	response.OK(w, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	itemID := pathItemID(w, r)
	if itemID == 0 {
		return
	}
	if err := h.service.Delete(r.Context(), listID, itemID); err != nil {
		handleItemError(w, err)
		return
	}
	response.OK(w, map[string]bool{"deleted": true})
}

func pathListID(w http.ResponseWriter, r *http.Request) uint64 {
	id, err := strconv.ParseUint(r.PathValue("list_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "无效的清单 ID")
		return 0
	}
	return id
}

func pathItemID(w http.ResponseWriter, r *http.Request) uint64 {
	id, err := strconv.ParseUint(r.PathValue("item_id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "无效的商品 ID")
		return 0
	}
	return id
}

func handleItemError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotFound) {
		response.Error(w, http.StatusNotFound, "item_not_found", "商品不存在")
		return
	}
	response.Error(w, http.StatusBadRequest, "item_operation_failed", err.Error())
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
