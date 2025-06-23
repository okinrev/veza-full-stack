package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response représente une réponse JSON standardisée
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// JSON envoie une réponse JSON avec le statut HTTP spécifié
func JSON(c *gin.Context, status int, data interface{}, message string) {
	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
		Message: message,
	}

	fmt.Printf("📤 Envoi de la réponse: status=%d, success=%v, message=%s\n", status, response.Success, message)
	if data != nil {
		fmt.Printf("📦 Données: %+v\n", data)
	}

	// Afficher la réponse JSON complète
	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("📄 Réponse JSON:\n%s\n", string(jsonBytes))

	c.JSON(status, response)
}

// Success envoie une réponse JSON de succès
func Success(c *gin.Context, data interface{}, message string) {
	JSON(c, http.StatusOK, data, message)
}

// Error envoie une réponse JSON d'erreur
func Error(c *gin.Context, status int, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}

	fmt.Printf("❌ Envoi de l'erreur: status=%d, message=%s\n", status, message)

	// Afficher la réponse JSON complète
	jsonBytes, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("📄 Réponse JSON:\n%s\n", string(jsonBytes))

	c.JSON(status, response)
}

// BadRequest envoie une réponse 400 Bad Request
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized envoie une réponse 401 Unauthorized
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden envoie une réponse 403 Forbidden
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound envoie une réponse 404 Not Found
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError envoie une réponse 500 Internal Server Error
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// ValidationError envoie une réponse 422 Unprocessable Entity
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message)
}

// TooManyRequests envoie une réponse 429 Too Many Requests
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, message)
}
