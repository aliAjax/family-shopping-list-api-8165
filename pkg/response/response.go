package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, map[string]any{"data": data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, map[string]any{"data": data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorEnvelope{Error: ErrorBody{Code: code, Message: message}})
}
