package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/services"
	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
)

// AuthHandler gère les requêtes d'authentification
type AuthHandler struct {
	authService services.AuthService
	config      *config.AppConfig
	logger      *zap.Logger
	metrics     *prometheus.Registry
}

// NewAuthHandler crée un nouveau handler d'authentification
func NewAuthHandler(
	authService services.AuthService,
	config *config.AppConfig,
	logger *zap.Logger,
	metrics *prometheus.Registry,
) (*AuthHandler, error) {
	return &AuthHandler{
		authService: authService,
		config:      config,
		logger:      logger,
		metrics:     metrics,
	}, nil
}

// RegisterRequest requête d'inscription
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest requête de connexion
type LoginRequest struct {
	Login    string `json:"login" binding:"required"` // Username ou email
	Password string `json:"password" binding:"required"`
}

// RefreshRequest requête de rafraîchissement de token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse réponse d'authentification
type AuthResponse struct {
	User         *entities.UserPublic `json:"user"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	ExpiresIn    int64                `json:"expires_in"`
}

// Register inscription d'un nouvel utilisateur
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erreur validation inscription", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Données d'inscription invalides",
			"details": err.Error(),
		})
		return
	}

	// Création de l'utilisateur via le service
	user, err := h.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Error("Erreur inscription utilisateur",
			zap.String("username", req.Username),
			zap.String("email", req.Email),
			zap.Error(err))

		status := http.StatusInternalServerError
		message := "Erreur lors de l'inscription"

		// Gestion des erreurs spécifiques
		switch err.Error() {
		case "username already exists":
			status = http.StatusConflict
			message = "Ce nom d'utilisateur est déjà utilisé"
		case "email already exists":
			status = http.StatusConflict
			message = "Cette adresse email est déjà utilisée"
		case "invalid username":
			status = http.StatusBadRequest
			message = "Nom d'utilisateur invalide"
		case "invalid email":
			status = http.StatusBadRequest
			message = "Adresse email invalide"
		case "weak password":
			status = http.StatusBadRequest
			message = "Mot de passe trop faible"
		}

		c.JSON(status, gin.H{"error": message})
		return
	}

	// Génération des tokens
	accessToken, refreshToken, err := h.authService.GenerateTokens(c.Request.Context(), user)
	if err != nil {
		h.logger.Error("Erreur génération tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la génération des tokens",
		})
		return
	}

	h.logger.Info("Utilisateur inscrit avec succès",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Inscription réussie",
		"data": AuthResponse{
			User:         user.ToPublic(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(h.config.JWT.AccessTTL.Seconds()),
		},
	})
}

// Login connexion d'un utilisateur
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erreur validation connexion", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Données de connexion invalides",
			"details": err.Error(),
		})
		return
	}

	// Authentification via le service
	user, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		h.logger.Error("Erreur connexion utilisateur",
			zap.String("login", req.Login),
			zap.Error(err))

		status := http.StatusUnauthorized
		message := "Identifiants invalides"

		switch err.Error() {
		case "user not found":
			message = "Utilisateur non trouvé"
		case "invalid password":
			message = "Mot de passe incorrect"
		case "user not active":
			status = http.StatusForbidden
			message = "Compte désactivé"
		case "user suspended":
			status = http.StatusForbidden
			message = "Compte suspendu"
		case "user banned":
			status = http.StatusForbidden
			message = "Compte banni"
		}

		c.JSON(status, gin.H{"error": message})
		return
	}

	// Génération des tokens
	accessToken, refreshToken, err := h.authService.GenerateTokens(c.Request.Context(), user)
	if err != nil {
		h.logger.Error("Erreur génération tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la génération des tokens",
		})
		return
	}

	h.logger.Info("Utilisateur connecté avec succès",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username))

	c.JSON(http.StatusOK, gin.H{
		"message": "Connexion réussie",
		"data": AuthResponse{
			User:         user.ToPublic(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(h.config.JWT.AccessTTL.Seconds()),
		},
	})
}

// RefreshToken rafraîchissement du token d'accès
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erreur validation refresh token", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token de rafraîchissement invalide",
		})
		return
	}

	// Rafraîchissement via le service
	user, accessToken, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Erreur rafraîchissement token", zap.Error(err))

		status := http.StatusUnauthorized
		message := "Token de rafraîchissement invalide ou expiré"

		switch err.Error() {
		case "token not found":
			message = "Token non trouvé"
		case "token expired":
			message = "Token expiré"
		case "user not found":
			message = "Utilisateur non trouvé"
		case "user not active":
			status = http.StatusForbidden
			message = "Compte désactivé"
		}

		c.JSON(status, gin.H{"error": message})
		return
	}

	h.logger.Info("Token rafraîchi avec succès", zap.Int64("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Token rafraîchi avec succès",
		"data": AuthResponse{
			User:         user.ToPublic(),
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    int64(h.config.JWT.AccessTTL.Seconds()),
		},
	})
}

// Logout déconnexion d'un utilisateur
func (h *AuthHandler) Logout(c *gin.Context) {
	// Récupération de l'utilisateur depuis le contexte (middleware auth)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	userIDInt64, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	// Révocation de tous les tokens via le service
	if err := h.authService.Logout(c.Request.Context(), userIDInt64); err != nil {
		h.logger.Error("Erreur déconnexion utilisateur",
			zap.Int64("user_id", userIDInt64),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la déconnexion",
		})
		return
	}

	h.logger.Info("Utilisateur déconnecté avec succès", zap.Int64("user_id", userIDInt64))

	c.JSON(http.StatusOK, gin.H{
		"message": "Déconnexion réussie",
	})
}

// Me informations de l'utilisateur connecté
func (h *AuthHandler) Me(c *gin.Context) {
	// Récupération de l'utilisateur depuis le contexte
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non authentifié"})
		return
	}

	userIDInt64, err := strconv.ParseInt(userID.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	// Récupération des informations utilisateur
	user, err := h.authService.GetUserByID(c.Request.Context(), userIDInt64)
	if err != nil {
		h.logger.Error("Erreur récupération utilisateur",
			zap.Int64("user_id", userIDInt64),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur lors de la récupération des informations utilisateur",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user.ToPublic(),
	})
}

// ChangePassword changement de mot de passe
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Données invalides",
			"details": err.Error(),
		})
		return
	}

	// Récupération de l'utilisateur
	userID, _ := c.Get("user_id")
	userIDInt64, _ := strconv.ParseInt(userID.(string), 10, 64)

	// Changement de mot de passe via le service
	if err := h.authService.ChangePassword(c.Request.Context(), userIDInt64, req.CurrentPassword, req.NewPassword); err != nil {
		h.logger.Error("Erreur changement mot de passe",
			zap.Int64("user_id", userIDInt64),
			zap.Error(err))

		status := http.StatusBadRequest
		message := "Erreur lors du changement de mot de passe"

		switch err.Error() {
		case "invalid current password":
			status = http.StatusUnauthorized
			message = "Mot de passe actuel incorrect"
		case "weak password":
			message = "Nouveau mot de passe trop faible"
		}

		c.JSON(status, gin.H{"error": message})
		return
	}

	h.logger.Info("Mot de passe changé avec succès", zap.Int64("user_id", userIDInt64))

	c.JSON(http.StatusOK, gin.H{
		"message": "Mot de passe changé avec succès",
	})
}
