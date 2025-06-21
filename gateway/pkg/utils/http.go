package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func WriteJSON(w http.ResponseWriter, payload any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(payload)
}

func DecodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func WriteValidationError(w http.ResponseWriter, err error) error {
	res := ValidationErrorResponse{
		Message: "invalid request",
		Fields:  make(map[string]string),
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			res.Fields[field] = fe.Tag()
		}
	}

	return WriteJSON(w, res, http.StatusBadRequest)
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, message string, code int) error {
	return WriteJSON(w, ErrorResponse{Message: message}, code)
}

type MessageReponse struct {
	Message string `json:"message"`
}

func WriteMessage(w http.ResponseWriter, message string) error {
	return WriteJSON(w, MessageReponse{Message: message}, http.StatusOK)
}
