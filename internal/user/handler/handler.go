package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"family-shopping-list-api/internal/middleware"
	"family-shopping-list-api/internal/user/service"
	"family-shopping-list-api/pkg/response"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	result, err := h.service.Register(r.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrUsernameTaken) {
			status = http.StatusConflict
		}
		response.Error(w, status, "register_failed", err.Error())
		return
	}
	response.Created(w, result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := decode(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		response.Error(w, status, "login_failed", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "缺少用户身份")
		return
	}
	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "get_user_failed", err.Error())
		return
	}
	response.OK(w, user)
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
