package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.InternalServerError(c, "Registration failed")
		return
	}

	response.Success(c, map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}, "User registered successfully")
}

func (h *Handler) Login(c *gin.Context) {
	// Log des headers
	fmt.Println("📨 Headers de la requête:")
	for k, v := range c.Request.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}

	// Log du corps de la requête
	body, _ := c.GetRawData()
	fmt.Printf("📦 Corps de la requête: %s\n", string(body))
	// Restaurer le corps pour qu'il puisse être lu à nouveau
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ Erreur de validation de la requête: %v\n", err)
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	fmt.Printf("📥 Tentative de connexion pour l'email: %s\n", req.Email)
	fmt.Printf("📦 Contenu de la requête: %+v\n", req)

	loginResp, err := h.service.Login(req)
	if err != nil {
		fmt.Printf("❌ Erreur de connexion: %v\n", err)
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	fmt.Printf("✅ Connexion réussie pour l'utilisateur: %+v\n", loginResp.User)
	fmt.Printf("🔑 Tokens générés: access_token=%s, refresh_token=%s\n", loginResp.AccessToken, loginResp.RefreshToken)

	// Afficher la réponse complète avant de l'envoyer
	fmt.Printf("📤 Réponse complète:\n")
	fmt.Printf("  User: %+v\n", loginResp.User)
	fmt.Printf("  AccessToken: %s\n", loginResp.AccessToken)
	fmt.Printf("  RefreshToken: %s\n", loginResp.RefreshToken)
	fmt.Printf("  ExpiresIn: %d\n", loginResp.ExpiresIn)

	// Vérifier que les tokens ne sont pas vides
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		fmt.Printf("❌ ERREUR: Les tokens sont vides!\n")
		response.Error(c, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	response.Success(c, loginResp, "Login successful")
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	tokenResp, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Invalid refresh token")
		return
	}

	response.Success(c, tokenResp, "Token refreshed")
}

func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	err := h.service.Logout(req.RefreshToken)
	if err != nil {
		response.InternalServerError(c, "Logout failed")
		return
	}

	response.Success(c, nil, "Logged out successfully")
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, err := h.service.GetMe(userID)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user, "User profile retrieved")
}
