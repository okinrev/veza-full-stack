// internal/common/context.go
package common

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	// Handle different types that might be stored
	switch v := userID.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return id, true
		}
	}

	return 0, false
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

// SetUserIDInContext sets user ID in Gin context
func SetUserIDInContext(c *gin.Context, userID int64) {
	c.Set(UserIDKey, userID)
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

// RequireOwnership middleware to check if user owns a resource
func RequireOwnership(getOwnerIDFunc func(*gin.Context) (int64, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		ownerID, err := getOwnerIDFunc(c)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid resource"})
			c.Abort()
			return
		}

		if userID != ownerID {
			c.JSON(403, gin.H{"error": "Forbidden: You don't own this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserIDFromParam extracts user ID from URL parameter
func GetUserIDFromParam(c *gin.Context, paramName string) (int64, error) {
	param := c.Param(paramName)
	if param == "" {
		return 0, errors.New("missing user ID parameter")
	}

	userID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	return userID, nil
}

// GetUserIDFromQuery extracts user ID from query parameter
func GetUserIDFromQuery(c *gin.Context, queryName string) (int64, error) {
	param := c.Query(queryName)
	if param == "" {
		return 0, errors.New("missing user ID query parameter")
	}

	userID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	return userID, nil
}
