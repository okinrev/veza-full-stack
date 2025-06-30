package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
	"github.com/okinrev/veza-web-app/internal/infrastructure/jwt"
)

// Phase2Handler gère les endpoints d'authentification Phase 2 complets
type Phase2Handler struct {
	userRepo   repositories.UserRepository
	cacheRepo  repositories.CacheRepository
	jwtService *jwt.JWTService
	logger     *zap.Logger
}

// NewPhase2Handler crée un nouveau handler Phase 2
func NewPhase2Handler(
	userRepo repositories.UserRepository,
	cacheRepo repositories.CacheRepository,
	jwtService *jwt.JWTService,
	logger *zap.Logger,
) *Phase2Handler {
	return &Phase2Handler{
		userRepo:   userRepo,
		cacheRepo:  cacheRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// Note: RegisterRequest et LoginRequest sont définis dans handler.go

// RefreshRequest représente une demande de renouvellement de token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse représente la réponse d'authentification complète
type AuthResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	User         *entities.PublicUser `json:"user"`
	ExpiresAt    time.Time            `json:"expires_at"`
	TokenType    string               `json:"token_type"`
}

// RegisterComplete gère l'inscription complète avec DB
func (h *Phase2Handler) RegisterComplete(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"phase":   "Phase 2 - Complete Registration",
		})
		return
	}

	ctx := c.Request.Context()

	// 1. Validation avancée
	if err := h.validateRegisterRequest(&req); err != nil {
		h.logger.Warn("Registration validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// 2. Vérifier unicité
	if exists, err := h.checkUserExists(ctx, req.Email, req.Username); err != nil {
		h.logger.Error("Error checking user existence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate user data",
		})
		return
	} else if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists with this email or username",
		})
		return
	}

	// 3. Créer l'utilisateur
	user, err := h.createUser(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user account",
		})
		return
	}

	// 4. Générer tokens et réponse
	authResponse, err := h.generateAuthResponse(ctx, user)
	if err != nil {
		h.logger.Error("Failed to generate auth response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to complete registration",
		})
		return
	}

	h.logger.Info("User registered successfully",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration completed successfully",
		"data":    authResponse,
		"phase":   "Phase 2 - Complete",
	})
}

// LoginComplete gère la connexion complète avec DB
func (h *Phase2Handler) LoginComplete(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
			"phase":   "Phase 2 - Complete Login",
		})
		return
	}

	ctx := c.Request.Context()

	// 1. Authentifier l'utilisateur
	user, err := h.authenticateUser(ctx, &req)
	if err != nil {
		h.logger.Warn("Authentication failed", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 2. Mettre à jour dernière connexion
	user.UpdateLastLogin()
	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.Warn("Failed to update last login", zap.Error(err))
	}

	// 3. Générer tokens et réponse
	authResponse, err := h.generateAuthResponse(ctx, user)
	if err != nil {
		h.logger.Error("Failed to generate auth response for login", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to complete login",
		})
		return
	}

	h.logger.Info("User logged in successfully",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    authResponse,
		"phase":   "Phase 2 - Complete",
	})
}

// RefreshComplete renouvelle les tokens avec validation complète
func (h *Phase2Handler) RefreshComplete(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// 1. Valider le refresh token
	session, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Invalid refresh token", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	// 2. Vérifier en base de données
	storedToken, err := h.userRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil || storedToken == nil {
		h.logger.Warn("Refresh token not found in database", zap.Int64("user_id", session.UserID))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Refresh token not found or expired",
		})
		return
	}

	// 3. Récupérer et valider l'utilisateur
	user, err := h.userRepo.GetByID(ctx, session.UserID)
	if err != nil || user == nil {
		h.logger.Warn("User not found for refresh", zap.Int64("user_id", session.UserID))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		return
	}

	if !user.IsAllowedToLogin() {
		h.logger.Warn("Refresh attempt for disabled account", zap.Int64("user_id", user.ID))
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Account is disabled",
		})
		return
	}

	// 4. Révoquer l'ancien token
	if err := h.userRepo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		h.logger.Warn("Failed to revoke old refresh token", zap.Error(err))
	}

	// 5. Générer nouveaux tokens
	authResponse, err := h.generateAuthResponse(ctx, user)
	if err != nil {
		h.logger.Error("Failed to generate new tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh tokens",
		})
		return
	}

	h.logger.Info("Tokens refreshed successfully", zap.Int64("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Tokens refreshed successfully",
		"data":    authResponse,
	})
}

// LogoutComplete déconnecte complètement l'utilisateur
func (h *Phase2Handler) LogoutComplete(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	ctx := c.Request.Context()

	// Récupérer le refresh token (optionnel)
	refreshToken := h.extractRefreshToken(c)

	// Révoquer le token spécifique si fourni
	if refreshToken != "" {
		if err := h.userRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
			h.logger.Warn("Failed to revoke refresh token", zap.Error(err))
		}
	}

	// Supprimer la session du cache
	sessionKey := fmt.Sprintf("user_session:%d", userID)
	if err := h.cacheRepo.Delete(ctx, sessionKey); err != nil {
		h.logger.Warn("Failed to delete session from cache", zap.Error(err))
	}

	h.logger.Info("User logged out", zap.Int64("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
		"phase":   "Phase 2 - Complete",
	})
}

// LogoutAllComplete déconnecte de toutes les sessions
func (h *Phase2Handler) LogoutAllComplete(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	ctx := c.Request.Context()

	// Révoquer tous les tokens
	if err := h.userRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		h.logger.Error("Failed to revoke all tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to logout from all sessions",
		})
		return
	}

	// Supprimer la session du cache
	sessionKey := fmt.Sprintf("user_session:%d", userID)
	if err := h.cacheRepo.Delete(ctx, sessionKey); err != nil {
		h.logger.Warn("Failed to delete session from cache", zap.Error(err))
	}

	h.logger.Info("User logged out from all sessions", zap.Int64("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all sessions successfully",
		"phase":   "Phase 2 - Complete",
	})
}

// ProfileComplete récupère le profil avec cache
func (h *Phase2Handler) ProfileComplete(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	ctx := c.Request.Context()

	// Essayer le cache d'abord
	sessionKey := fmt.Sprintf("user_session:%d", userID)
	var cachedUser entities.User
	fromCache := false

	if err := h.cacheRepo.Get(ctx, sessionKey, &cachedUser); err == nil {
		fromCache = true
		c.JSON(http.StatusOK, gin.H{
			"user":   cachedUser.ToPublic(),
			"cached": fromCache,
			"phase":  "Phase 2 - Complete",
		})
		return
	}

	// Récupérer depuis la DB
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve profile",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Mettre en cache
	if err := h.cacheRepo.Set(ctx, sessionKey, user, 30*time.Minute); err != nil {
		h.logger.Warn("Failed to cache user session", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   user.ToPublic(),
		"cached": fromCache,
		"phase":  "Phase 2 - Complete",
	})
}

// Méthodes utilitaires

func (h *Phase2Handler) checkUserExists(ctx context.Context, email, username string) (bool, error) {
	// Vérifier email
	existingByEmail, err := h.userRepo.GetByEmail(ctx, email)
	if err == nil && existingByEmail != nil {
		return true, nil
	}

	// Vérifier username
	existingByUsername, err := h.userRepo.GetByUsername(ctx, username)
	if err == nil && existingByUsername != nil {
		return true, nil
	}

	return false, nil
}

func (h *Phase2Handler) createUser(ctx context.Context, req *RegisterRequest) (*entities.User, error) {
	// Hacher le mot de passe
	hashedPassword, err := h.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Créer l'entité
	user, err := entities.NewUser(req.Username, req.Email, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity: %w", err)
	}

	// Sauvegarder en base
	err = h.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (h *Phase2Handler) authenticateUser(ctx context.Context, req *LoginRequest) (*entities.User, error) {
	// Récupérer l'utilisateur
	user, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Vérifier statut
	if !user.IsAllowedToLogin() {
		return nil, fmt.Errorf("account disabled")
	}

	// Vérifier mot de passe
	if err := h.verifyPassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (h *Phase2Handler) generateAuthResponse(ctx context.Context, user *entities.User) (*AuthResponse, error) {
	// Créer session JWT
	jwtSession := &jwt.UserSession{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}

	// Générer access token
	accessToken, err := h.jwtService.GenerateAccessToken(jwtSession)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Générer refresh token
	refreshToken, err := h.jwtService.GenerateRefreshToken(jwtSession)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Sauvegarder refresh token
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()
	if err := h.userRepo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Cache la session
	sessionKey := fmt.Sprintf("user_session:%d", user.ID)
	if err := h.cacheRepo.Set(ctx, sessionKey, user, 30*time.Minute); err != nil {
		h.logger.Warn("Failed to cache session", zap.Error(err))
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToPublic(),
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		TokenType:    "Bearer",
	}, nil
}

func (h *Phase2Handler) getUserIDFromContext(c *gin.Context) (int64, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("user not authenticated")
	}

	if userID, ok := userIDStr.(int64); ok {
		return userID, nil
	}

	if userIDString, ok := userIDStr.(string); ok {
		return strconv.ParseInt(userIDString, 10, 64)
	}

	return 0, fmt.Errorf("invalid user ID type")
}

func (h *Phase2Handler) extractRefreshToken(c *gin.Context) string {
	// Depuis l'en-tête Authorization
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if parts := strings.Split(authHeader, " "); len(parts) == 2 && parts[0] == "Refresh" {
			return parts[1]
		}
	}

	// Depuis le body (optionnel)
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		return req.RefreshToken
	}

	return ""
}

func (h *Phase2Handler) validateRegisterRequest(req *RegisterRequest) error {
	tempUser := &entities.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := tempUser.ValidateUsername(); err != nil {
		return err
	}
	if err := tempUser.ValidateEmail(); err != nil {
		return err
	}
	if err := tempUser.ValidatePassword(); err != nil {
		return err
	}

	return nil
}

func (h *Phase2Handler) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (h *Phase2Handler) verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
