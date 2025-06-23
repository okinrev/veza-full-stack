// internal/common/context.go
package common

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext récupère l'ID de l'utilisateur depuis le contexte
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// Gestion des différents types possibles
	switch v := userID.(type) {
	case int:
		return v, true
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

// GetUsernameFromContext récupère le nom d'utilisateur depuis le contexte
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetUserRoleFromContext récupère le rôle de l'utilisateur depuis le contexte
func GetUserRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return role.(string), true
}

// GetRequestIDFromContext récupère l'ID de la requête depuis le contexte
func GetRequestIDFromContext(c *gin.Context) (string, bool) {
	requestID, exists := c.Get("request_id")
	if !exists {
		return "", false
	}
	return requestID.(string), true
}

// SetUserIDInContext définit l'ID de l'utilisateur dans le contexte
func SetUserIDInContext(c *gin.Context, userID int) {
	c.Set("user_id", userID)
}

// SetUsernameInContext définit le nom d'utilisateur dans le contexte
func SetUsernameInContext(c *gin.Context, username string) {
	c.Set("username", username)
}

// SetUserRoleInContext définit le rôle de l'utilisateur dans le contexte
func SetUserRoleInContext(c *gin.Context, role string) {
	c.Set("user_role", role)
}

// SetRequestIDInContext définit l'ID de la requête dans le contexte
func SetRequestIDInContext(c *gin.Context, requestID string) {
	c.Set("request_id", requestID)
}

// RequireOwnership middleware checks if user owns the resource
func RequireOwnership(getOwnerIDFunc func(*gin.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User not authenticated",
			})
			c.Abort()
			return
		}

		ownerID, err := getOwnerIDFunc(c)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Resource not found",
			})
			c.Abort()
			return
		}

		if userID != ownerID {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
