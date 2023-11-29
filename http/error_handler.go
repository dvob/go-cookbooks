package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// errorHandler responds with error an error
func errorHandler(w http.ResponseWriter, r *http.Request, code int, errorMessage any) {
	var message string
	if errorMessage == nil {
		message = http.StatusText(code)
	} else {
		message = fmt.Sprint(errorMessage)
	}

	requestID := getRequestID(r.Context()).String()

	contentType := r.Header.Get("Accept")
	switch contentType {
	case "application/json":
		errJSON := struct {
			RequestID string `json:"request_id,omitempty"`
			Error     any    `json:"error"`
		}{
			RequestID: requestID,
			Error:     message,
		}

		out, err := json.Marshal(errJSON)
		if err != nil {
			slog.LogAttrs(r.Context(), slog.LevelError, "failed to format error", slog.Any("err", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(code)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
		return
	default:
		if requestID != "" {
			message += " (request ID: " + requestID + ")"
		}
		http.Error(w, message, code)
		return
	}
}
