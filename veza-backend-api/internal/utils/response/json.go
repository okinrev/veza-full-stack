//file: internal/utils/response/json.go

package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// SuccessJSON envoie une réponse JSON de succès
func SuccessJSON(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ErrorJSON envoie une réponse JSON d'erreur
func ErrorJSON(w http.ResponseWriter, message string, statusCode int) {
	response := APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// PaginatedJSON envoie une réponse paginée
func PaginatedJSON(w http.ResponseWriter, data interface{}, meta *Meta, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APIResponse{
		Success: true,
		Data:    data,
		Message: message,
		Meta:    meta,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ValidationErrorJSON envoie une réponse JSON d'erreur de validation
func ValidationErrorJSON(w http.ResponseWriter, errors []ValidationError) {
	response := APIResponse{
		Success: false,
		Message: "Validation failed",
		Data:    errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Fallback to plain text if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
