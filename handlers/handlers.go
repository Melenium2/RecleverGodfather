package handlers

import (
	"RecleverGodfather/grandlog"
	"RecleverGodfather/grandlog/loggerepo"
	"encoding/json"
	"errors"
	"net/http"
)

func Log(logger grandlog.GrandLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := &loggerepo.SingleLog{}
		if err := json.NewDecoder(r.Body).Decode(log); err != nil {
			logger.Log("[Error]", err)
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if log.MessageType == "" {
			logger.Log("[Error]", "empty request")
			writeError(w, http.StatusBadRequest, errors.New("empty request"))
			return
		}
		if err := logger.LogObject(r.Context(), log); err != nil {
			logger.Log("[Error]", err)
			writeError(w, http.StatusBadRequest, err)
			return
		}

		writeResponse(w, http.StatusOK, "Recorded")
	}
}

func writeResponse(w http.ResponseWriter,code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			writeError(w, http.StatusInternalServerError, err)
		}
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeResponse(w, code, map[string]interface{} {
		"error": err.Error(),
	})
}