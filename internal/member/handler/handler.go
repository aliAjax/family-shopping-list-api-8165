package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"family-shopping-list-api/internal/member/service"
	"family-shopping-list-api/pkg/response"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	members, err := h.service.List(r.Context(), listID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "list_members_failed", err.Error())
		return
	}
	response.OK(w, members)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	listID := pathListID(w, r)
	if listID == 0 {
		return
	}
	memberID, err := strconv.ParseUint(r.PathValue("member_id"), 10, 64)
	if err != nil || memberID == 0 {
		response.Error(w, http.StatusBadRequest, "bad_request", "无效的成员 ID")
		return
	}
	if err := h.service.Remove(r.Context(), listID, memberID); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrCannotRemoveOwner) {
			status = http.StatusForbidden
		}
		response.Error(w, status, "remove_member_failed", err.Error())
		return
	}
	response.OK(w, map[string]bool{"removed": true})
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
