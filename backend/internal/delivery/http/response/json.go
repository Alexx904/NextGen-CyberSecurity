package response

import (
	"encoding/json"
	"net/http"
)

// APIError modella il dettaglio di un errore restituito dal server
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Envelope rappresenta la struttura standard di risposta JSON API v1
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *APIError   `json:"error"`
}

// JSON invia una risposta HTTP di successo formattata secondo lo standard della piattaforma
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Envelope{
		Success: true,
		Data:    data,
		Error:   nil,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// Error invia una risposta HTTP di errore strutturata con codice di errore e messaggio
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Envelope{
		Success: false,
		Data:    nil,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

