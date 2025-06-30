package services

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
	"github.com/okinrev/veza-web-app/internal/infrastructure/jwt"
)

// AuthService interface pour les services d'authentification
type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*entities.User, error)
	Login(ctx context.Context, login, password string) (*entities.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entities.User, string, string, error)
	GenerateTokens(ctx context.Context, user *entities.User) (string, string, error)
	Logout(ctx context.Context, userID int64) error
	GetUserByID(ctx context.Context, userID int64) (*entities.User, error)
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
	ValidateToken(ctx context.Context, token string) (*entities.UserSession, error)
}

// authService implémentation du service d'authentification
type authService struct {
	userRepo     repositories.UserRepository
	cacheService CacheService
	jwtConfig    config.JWTConfig
	logger       *zap.Logger
	jwtService   *jwt.JWTService
	hashCost     int
}

// NewAuthService crée une nouvelle instance du service d'authentification
func NewAuthService(
	userRepo repositories.UserRepository,
	cacheService CacheService,
	jwtConfig config.JWTConfig,
	logger *zap.Logger,
	jwtService *jwt.JWTService,
) (AuthService, error) {
	return &authService{
		userRepo:     userRepo,
		cacheService: cacheService,
		jwtConfig:    jwtConfig,
		logger:       logger,
		jwtService:   jwtService,
		hashCost:     12, // bcrypt cost élevé pour la sécurité
	}, nil
}

// Register inscrit un nouvel utilisateur
func (s *authService) Register(ctx context.Context, username, email, password string) (*entities.User, error) {
	// Vérifier si l'utilisateur existe déjà
	exists, err := s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Erreur vérification username", zap.Error(err))
		return nil, errors.New("erreur interne du serveur")
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	exists, err = s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Erreur vérification email", zap.Error(err))
		return nil, errors.New("erreur interne du serveur")
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Créer l'utilisateur
	user, err := entities.NewUser(username, email, password)
	if err != nil {
		return nil, err
	}

	// Hasher le mot de passe
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		s.logger.Error("Erreur hashage mot de passe", zap.Error(err))
		return nil, errors.New("erreur interne du serveur")
	}
	user.Password = hashedPassword

	// Sauvegarder en base
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("Erreur création utilisateur", zap.Error(err))
		return nil, errors.New("erreur lors de la création de l'utilisateur")
	}

	s.logger.Info("Utilisateur créé avec succès",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username))

	return user, nil
}

// Login authentifie un utilisateur
func (s *authService) Login(ctx context.Context, login, password string) (*entities.User, error) {
	// Chercher l'utilisateur par username ou email
	var user *entities.User
	var err error

	// Essayer par username d'abord
	user, err = s.userRepo.GetByUsername(ctx, login)
	if err != nil || user == nil {
		// Essayer par email
		user, err = s.userRepo.GetByEmail(ctx, login)
		if err != nil || user == nil {
			return nil, errors.New("user not found")
		}
	}

	// Vérifier que l'utilisateur peut se connecter
	if !user.IsAllowedToLogin() {
		switch user.Status {
		case entities.StatusSuspended:
			return nil, errors.New("user suspended")
		case entities.StatusBanned:
			return nil, errors.New("user banned")
		default:
			return nil, errors.New("user not active")
		}
	}

	// Vérifier le mot de passe
	if !s.verifyPassword(password, user.Password) {
		return nil, errors.New("invalid password")
	}

	// Mettre à jour la dernière connexion
	user.UpdateLastLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Warn("Erreur mise à jour dernière connexion", zap.Error(err))
	}

	return user, nil
}

// GenerateTokens génère les tokens JWT
func (s *authService) GenerateTokens(ctx context.Context, user *entities.User) (string, string, error) {
	// Générer l'access token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	// Générer le refresh token
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	// Sauvegarder le refresh token
	expiresAt := time.Now().Add(s.jwtConfig.RefreshTTL).Unix()
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		s.logger.Error("Erreur sauvegarde refresh token", zap.Error(err))
		return "", "", errors.New("erreur génération tokens")
	}

	return accessToken, refreshToken, nil
}

// RefreshToken rafraîchit les tokens
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*entities.User, string, string, error) {
	// Vérifier le refresh token
	tokenData, err := s.userRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil || tokenData == nil {
		return nil, "", "", errors.New("token not found")
	}

	// Vérifier l'expiration
	if time.Now().Unix() > tokenData.ExpiresAt {
		// Supprimer le token expiré
		s.userRepo.RevokeRefreshToken(ctx, refreshToken)
		return nil, "", "", errors.New("token expired")
	}

	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByID(ctx, tokenData.UserID)
	if err != nil || user == nil {
		return nil, "", "", errors.New("user not found")
	}

	// Vérifier que l'utilisateur peut se connecter
	if !user.IsAllowedToLogin() {
		return nil, "", "", errors.New("user not active")
	}

	// Révoquer l'ancien token si rotation activée
	if s.jwtConfig.RefreshRotation {
		if err := s.userRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
			s.logger.Warn("Erreur révocation ancien token", zap.Error(err))
		}
	}

	// Générer de nouveaux tokens
	newAccessToken, newRefreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, newAccessToken, newRefreshToken, nil
}

// Logout déconnecte un utilisateur
func (s *authService) Logout(ctx context.Context, userID int64) error {
	// Révoquer tous les tokens de l'utilisateur
	if err := s.userRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		s.logger.Error("Erreur révocation tokens", zap.Int64("user_id", userID), zap.Error(err))
		return errors.New("erreur lors de la déconnexion")
	}

	// Nettoyer le cache si disponible
	if s.cacheService != nil {
		cacheKey := s.getUserCacheKey(userID)
		s.cacheService.Delete(ctx, cacheKey)
	}

	return nil
}

// GetUserByID récupère un utilisateur par son ID
func (s *authService) GetUserByID(ctx context.Context, userID int64) (*entities.User, error) {
	// Essayer le cache d'abord
	if s.cacheService != nil {
		cacheKey := s.getUserCacheKey(userID)
		if cached, err := s.cacheService.Get(ctx, cacheKey); err == nil && cached != nil {
			if user, ok := cached.(*entities.User); ok {
				return user, nil
			}
		}
	}

	// Récupérer depuis la base
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Mettre en cache
	if s.cacheService != nil && user != nil {
		cacheKey := s.getUserCacheKey(userID)
		s.cacheService.Set(ctx, cacheKey, user, time.Hour)
	}

	return user, nil
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *authService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Vérifier le mot de passe actuel
	if !s.verifyPassword(currentPassword, user.Password) {
		return errors.New("invalid current password")
	}

	// Valider le nouveau mot de passe
	tempUser := &entities.User{Password: newPassword}
	if err := tempUser.ValidatePassword(); err != nil {
		return errors.New("weak password")
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		s.logger.Error("Erreur hashage nouveau mot de passe", zap.Error(err))
		return errors.New("erreur interne du serveur")
	}

	// Mettre à jour
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.Error("Erreur mise à jour mot de passe", zap.Error(err))
		return errors.New("erreur lors de la mise à jour")
	}

	// Révoquer tous les tokens existants pour forcer une nouvelle connexion
	s.userRepo.RevokeAllUserTokens(ctx, userID)

	// Nettoyer le cache
	if s.cacheService != nil {
		cacheKey := s.getUserCacheKey(userID)
		s.cacheService.Delete(ctx, cacheKey)
	}

	return nil
}

// ValidateToken valide un token JWT
func (s *authService) ValidateToken(ctx context.Context, token string) (*entities.UserSession, error) {
	// TODO: Implémenter la validation JWT
	// Cette méthode sera complétée avec la logique JWT complète
	return nil, errors.New("not implemented")
}

// Méthodes utilitaires privées

func (s *authService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.hashCost)
	return string(bytes), err
}

func (s *authService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *authService) generateAccessToken(user *entities.User) (string, error) {
	// TODO: Implémenter la génération JWT access token
	return "mock_access_token", nil
}

func (s *authService) generateRefreshToken(user *entities.User) (string, error) {
	// TODO: Implémenter la génération JWT refresh token
	return "mock_refresh_token", nil
}

func (s *authService) getUserCacheKey(userID int64) string {
	return "user:" + string(rune(userID))
}
