package main

import (
	"github.com/pjebs/jsonerror"
	"net/http"
)

func SetCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "*")
}

// NewError is a helper function to create a Response with an error
func NewError(code int, error, message string) *Response {
	return &Response{
		Successful: false,
		Error:      jsonerror.New(code, error, message).Render(),
	}
}

func InvalidJSONError(w http.ResponseWriter, err error) {
	response := NewError(1, "Invalid JSON request", err.Error())
	RenderJSONResponse(w, http.StatusBadRequest, response)
}
