package remoteclients

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrBadRoute = errors.New("route not found")
)

func encodeHttpError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case ErrBadRequest:
		w.WriteHeader(http.StatusBadRequest)
	case ErrBadRoute:
		w.WriteHeader(http.StatusNotFound)
	default:
	w.WriteHeader(http.StatusBadRequest)
	}
	e := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func encodeHTTPResponse(_ context.Context, code int, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(code)
	if response == nil {
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
