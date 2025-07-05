package services

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
	"github.com/okinrev/veza-web-app/internal/core/domain/repositories"
	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
)

// AuthService service d'authentification enterprise
type AuthService interface {
	// Authentification standard
	Login(ctx context.Context, email, password string) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, sessionToken string) error
	LogoutAll(ctx context.Context, userID int64) error

	// OAuth2
	GetOAuthURL(ctx context.Context, provider string, state string) (string, error)
	HandleOAuthCallback(ctx context.Context, provider, code, state string) (*AuthResponse, error)

	// 2FA
	Enable2FA(ctx context.Context, userID int64) (*TwoFactorSetup, error)
	Verify2FA(ctx context.Context, userID int64, token string) error
	Disable2FA(ctx context.Context, userID int64, password string) error

	// Gestion des mots de passe
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	ResetPasswordRequest(ctx context.Context, email string) error
	ResetPasswordConfirm(ctx context.Context, token, newPassword string) error

	// Sessions et sécurité
	ValidateSession(ctx context.Context, sessionToken string) (*entities.User, error)
	GetActiveSessions(ctx context.Context, userID int64) ([]*repositories.UserSession, error)
	RevokeSession(ctx context.Context, sessionID int64) error
}

// authService implémentation du service d'authentification
type authService struct {
	userRepo     repositories.UserRepository
	emailService EmailService
	oauthService *OAuthService
	logger       *zap.Logger
	config       *config.Config

	// OAuth2 configs
	googleOAuth  *oauth2.Config
	githubOAuth  *oauth2.Config
	discordOAuth *oauth2.Config

	// JWT
	jwtSecret     []byte
	jwtExpiry     time.Duration
	refreshExpiry time.Duration
}

// NewAuthService crée une nouvelle instance du service d'authentification
func NewAuthService(
	userRepo repositories.UserRepository,
	emailService EmailService,
	logger *zap.Logger,
	cfg *config.Config,
) AuthService {
	service := &authService{
		userRepo:      userRepo,
		emailService:  emailService,
		oauthService:  NewOAuthService(logger),
		logger:        logger,
		config:        cfg,
		jwtSecret:     []byte(cfg.JWT.Secret),
		jwtExpiry:     cfg.JWT.Expiry,
		refreshExpiry: cfg.JWT.RefreshExpiry,
	}

	// Configuration OAuth2
	service.setupOAuthConfigs()

	return service
}

// setupOAuthConfigs configure les clients OAuth2
func (s *authService) setupOAuthConfigs() {
	// Google OAuth2
	s.googleOAuth = &oauth2.Config{
		ClientID:     s.config.OAuth.Google.ClientID,
		ClientSecret: s.config.OAuth.Google.ClientSecret,
		RedirectURL:  s.config.OAuth.Google.RedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	// GitHub OAuth2
	s.githubOAuth = &oauth2.Config{
		ClientID:     s.config.OAuth.GitHub.ClientID,
		ClientSecret: s.config.OAuth.GitHub.ClientSecret,
		RedirectURL:  s.config.OAuth.GitHub.RedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Discord OAuth2
	s.discordOAuth = &oauth2.Config{
		ClientID:     s.config.OAuth.Discord.ClientID,
		ClientSecret: s.config.OAuth.Discord.ClientSecret,
		RedirectURL:  s.config.OAuth.Discord.RedirectURL,
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}

// Login authentifie un utilisateur avec email/mot de passe
func (s *authService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("Tentative de connexion avec email invalide", zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Vérifier si le compte est actif
	if !user.IsActive() {
		return nil, ErrAccountInactive
	}

	// Vérifier si le compte est verrouillé
	if user.IsLocked() {
		return nil, ErrAccountLocked
	}

	// Vérifier le mot de passe
	if !user.CheckPassword(password) {
		// Incrémenter les tentatives échouées
		if err := s.userRepo.UpdateLoginAttempts(ctx, user.ID, user.LoginAttempts+1, nil); err != nil {
			s.logger.Error("Erreur incrémentation tentatives échouées", zap.Error(err))
		}

		// Log de sécurité
		s.auditLog(ctx, user.ID, "login_failed", "invalid_password", nil)
		return nil, ErrInvalidCredentials
	}

	// Réinitialiser les tentatives échouées
	if err := s.userRepo.ResetLoginAttempts(ctx, user.ID); err != nil {
		s.logger.Error("Erreur reset tentatives échouées", zap.Error(err))
	}

	// Vérifier si 2FA est activé
	if user.TwoFactorEnabled {
		// Retourner une réponse indiquant que 2FA est requis
		return &AuthResponse{
			RequiresTwoFactor: true,
			UserID:            user.ID,
		}, nil
	}

	// Créer la session
	return s.createSession(ctx, user, "", "password")
}

// Register crée un nouveau compte utilisateur
func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Vérifier si l'email existe déjà
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("vérification email existant: %w", err)
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Vérifier si le username existe déjà
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("vérification username existant: %w", err)
	}

	if existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	}

	// Créer l'utilisateur
	user := &entities.User{
		UUID:          uuid.New().String(),
		Username:      req.Username,
		Email:         req.Email,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		DisplayName:   req.DisplayName,
		Role:          entities.RoleUser,
		Status:        entities.StatusActive,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Hasher le mot de passe
	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("hashage mot de passe: %w", err)
	}

	// Créer l'utilisateur en base
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("création utilisateur: %w", err)
	}

	// Log d'audit
	s.auditLog(ctx, user.ID, "user_registered", "registration", map[string]interface{}{
		"email":    user.Email,
		"username": user.Username,
	})

	// Créer la session
	return s.createSession(ctx, user, "", "registration")
}

// RefreshToken rafraîchit les tokens JWT
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Récupérer la session par refresh token
	session, err := s.userRepo.GetSession(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("récupération session: %w", err)
	}

	if session == nil {
		return nil, ErrInvalidToken
	}

	// Vérifier si la session est encore valide
	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Récupérer l'utilisateur
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil || !user.IsActive() {
		return nil, ErrAccountInactive
	}

	// Générer nouveaux tokens
	accessToken, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("génération JWT: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("génération refresh token: %w", err)
	}

	// Mettre à jour la session
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(s.refreshExpiry)
	session.LastActivity = time.Now()

	if err := s.userRepo.UpdateSession(ctx, session.SessionToken, session.LastActivity); err != nil {
		return nil, fmt.Errorf("mise à jour session: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtExpiry.Seconds()),
		User:         user,
	}, nil
}

// Logout déconnecte un utilisateur
func (s *authService) Logout(ctx context.Context, sessionToken string) error {
	if err := s.userRepo.InvalidateSession(ctx, sessionToken); err != nil {
		return fmt.Errorf("invalidation session: %w", err)
	}

	return nil
}

// LogoutAll déconnecte un utilisateur de toutes ses sessions
func (s *authService) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.userRepo.InvalidateAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("invalidation toutes sessions: %w", err)
	}

	// Log d'audit
	s.auditLog(ctx, userID, "logout_all", "security", nil)

	return nil
}

// GetOAuthURL génère l'URL d'authentification OAuth2
func (s *authService) GetOAuthURL(ctx context.Context, provider, state string) (string, error) {
	var config *oauth2.Config

	switch provider {
	case "google":
		config = s.googleOAuth
	case "github":
		config = s.githubOAuth
	case "discord":
		config = s.discordOAuth
	default:
		return "", ErrUnsupportedProvider
	}

	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, nil
}

// HandleOAuthCallback traite le callback OAuth2
func (s *authService) HandleOAuthCallback(ctx context.Context, provider, code, state string) (*AuthResponse, error) {
	var config *oauth2.Config

	switch provider {
	case "google":
		config = s.googleOAuth
	case "github":
		config = s.githubOAuth
	case "discord":
		config = s.discordOAuth
	default:
		return nil, ErrUnsupportedProvider
	}

	// Échanger le code contre un token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("échange code OAuth2: %w", err)
	}

	// Récupérer les informations utilisateur
	userInfo, err := s.fetchOAuthUserInfo(ctx, provider, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("récupération info utilisateur OAuth2: %w", err)
	}

	// Chercher un utilisateur existant par email
	user, err := s.userRepo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("recherche utilisateur: %w", err)
	}

	if user == nil {
		// Créer un nouvel utilisateur
		user = &entities.User{
			UUID:          uuid.New().String(),
			Username:      userInfo.Username,
			Email:         userInfo.Email,
			FirstName:     userInfo.FirstName,
			LastName:      userInfo.LastName,
			DisplayName:   userInfo.DisplayName,
			Avatar:        userInfo.Avatar,
			Role:          entities.UserRoleUser,
			Status:        entities.UserStatusActive,
			EmailVerified: true, // OAuth implique email vérifié
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := s.userRepo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("création utilisateur OAuth2: %w", err)
		}

		// Log d'audit
		s.auditLog(ctx, user.ID, "oauth_registration", provider, map[string]interface{}{
			"email":    user.Email,
			"provider": provider,
		})
	} else {
		// Log d'audit
		s.auditLog(ctx, user.ID, "oauth_login", provider, nil)
	}

	// Créer la session
	return s.createSession(ctx, user, "", fmt.Sprintf("oauth_%s", provider))
}

// Enable2FA active l'authentification à deux facteurs
func (s *authService) Enable2FA(ctx context.Context, userID int64) (*TwoFactorSetup, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	if user.TwoFactorEnabled {
		return nil, ErrTwoFactorAlreadyEnabled
	}

	// Générer une clé secrète
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Veza",
		AccountName: user.Email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("génération clé TOTP: %w", err)
	}

	// Sauvegarder la clé secrète (temporairement, jusqu'à confirmation)
	if err := s.userRepo.SetTwoFactorSecret(ctx, userID, key.Secret()); err != nil {
		return nil, fmt.Errorf("sauvegarde secret 2FA: %w", err)
	}

	// Générer des codes de récupération
	recoveryCodes, err := s.generateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("génération codes récupération: %w", err)
	}

	if err := s.userRepo.SetRecoveryCodes(ctx, userID, recoveryCodes); err != nil {
		return nil, fmt.Errorf("sauvegarde codes récupération: %w", err)
	}

	return &TwoFactorSetup{
		Secret:        key.Secret(),
		QRCode:        key.String(),
		RecoveryCodes: recoveryCodes,
	}, nil
}

// Verify2FA vérifie un code 2FA
func (s *authService) Verify2FA(ctx context.Context, userID int64, token string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	// Récupérer le secret 2FA
	secret, err := s.userRepo.GetTwoFactorSecret(ctx, userID)
	if err != nil {
		return fmt.Errorf("récupération secret 2FA: %w", err)
	}

	// Vérifier le token TOTP
	valid := totp.Validate(token, secret)
	if !valid {
		// Vérifier si c'est un code de récupération
		if err := s.verifyRecoveryCode(ctx, userID, token); err != nil {
			return ErrInvalidTwoFactorCode
		}
	}

	// Activer 2FA si ce n'est pas encore fait
	if !user.TwoFactorEnabled {
		if err := s.userRepo.EnableTwoFactor(ctx, userID); err != nil {
			return fmt.Errorf("activation 2FA: %w", err)
		}

		// Log d'audit
		s.auditLog(ctx, userID, "2fa_enabled", "security", nil)
	}

	return nil
}

// Disable2FA désactive l'authentification à deux facteurs
func (s *authService) Disable2FA(ctx context.Context, userID int64, password string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	// Vérifier le mot de passe
	if !user.ValidatePassword(password) {
		return ErrInvalidPassword
	}

	// Désactiver 2FA
	if err := s.userRepo.DisableTwoFactor(ctx, userID); err != nil {
		return fmt.Errorf("désactivation 2FA: %w", err)
	}

	// Log d'audit
	s.auditLog(ctx, userID, "2fa_disabled", "security", nil)

	return nil
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *authService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	// Vérifier l'ancien mot de passe
	if !user.ValidatePassword(oldPassword) {
		return ErrInvalidPassword
	}

	// Hasher le nouveau mot de passe
	if err := user.HashPassword(newPassword); err != nil {
		return fmt.Errorf("hashage nouveau mot de passe: %w", err)
	}

	// Mettre à jour en base
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("mise à jour mot de passe: %w", err)
	}

	// Log d'audit
	s.auditLog(ctx, userID, "password_changed", "security", nil)

	return nil
}

// ResetPasswordRequest initie une réinitialisation de mot de passe
func (s *authService) ResetPasswordRequest(ctx context.Context, email string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		// Ne pas révéler si l'email existe ou non
		return nil
	}

	// Générer un token de réinitialisation
	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := s.userRepo.CreatePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return fmt.Errorf("création token reset: %w", err)
	}

	// Envoyer l'email de réinitialisation
	if err := s.emailService.SendPasswordReset(ctx, user.Email, token); err != nil {
		return fmt.Errorf("envoi email reset: %w", err)
	}

	// Log d'audit
	s.auditLog(ctx, user.ID, "password_reset_requested", "security", nil)

	return nil
}

// ResetPasswordConfirm confirme la réinitialisation de mot de passe
func (s *authService) ResetPasswordConfirm(ctx context.Context, token, newPassword string) error {
	// Récupérer l'utilisateur par token
	user, err := s.userRepo.GetUserByPasswordResetToken(ctx, token)
	if err != nil {
		return fmt.Errorf("récupération utilisateur par token: %w", err)
	}

	if user == nil {
		return ErrInvalidToken
	}

	// Hasher le nouveau mot de passe
	if err := user.HashPassword(newPassword); err != nil {
		return fmt.Errorf("hashage nouveau mot de passe: %w", err)
	}

	// Mettre à jour le mot de passe et supprimer le token
	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("mise à jour mot de passe: %w", err)
	}

	// TODO: Implémenter DeletePasswordResetToken avec le bon paramètre
	// if err := s.userRepo.DeletePasswordResetToken(ctx, token); err != nil {
	// 	s.logger.Warn("Erreur suppression token reset", zap.Error(err))
	// }

	// Invalider toutes les sessions
	if err := s.userRepo.InvalidateAllUserSessions(ctx, user.ID); err != nil {
		s.logger.Warn("Erreur invalidation sessions", zap.Error(err))
	}

	// Log d'audit
	s.auditLog(ctx, user.ID, "password_reset_completed", "security", nil)

	return nil
}

// ValidateSession valide une session utilisateur
func (s *authService) ValidateSession(ctx context.Context, sessionToken string) (*entities.User, error) {
	session, err := s.userRepo.GetSession(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("récupération session: %w", err)
	}

	if session == nil {
		return nil, ErrInvalidToken
	}

	// Vérifier l'expiration
	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Mettre à jour l'activité
	if err := s.userRepo.UpdateSession(ctx, sessionToken, time.Now()); err != nil {
		s.logger.Warn("Erreur mise à jour activité session", zap.Error(err))
	}

	return session.User, nil
}

// GetActiveSessions récupère les sessions actives d'un utilisateur
func (s *authService) GetActiveSessions(ctx context.Context, userID int64) ([]*repositories.UserSession, error) {
	return s.userRepo.GetUserSessions(ctx, userID)
}

// RevokeSession révoque une session spécifique
func (s *authService) RevokeSession(ctx context.Context, sessionID int64) error {
	// TODO: Implémenter la révocation par ID de session
	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// createSession crée une nouvelle session utilisateur
func (s *authService) createSession(ctx context.Context, user *entities.User, deviceInfo, loginMethod string) (*AuthResponse, error) {
	// Générer les tokens
	accessToken, err := s.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("génération JWT: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("génération refresh token: %w", err)
	}

	sessionToken := uuid.New().String()

	// Créer la session
	session := &repositories.UserSession{
		UserID:       user.ID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		DeviceInfo:   deviceInfo,
		IsActive:     true,
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
		CreatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("création session: %w", err)
	}

	// Mettre à jour la dernière connexion de l'utilisateur
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, ""); err != nil {
		s.logger.Warn("Erreur mise à jour dernière connexion", zap.Error(err))
	}

	// Log d'audit
	s.auditLog(ctx, user.ID, "login_success", loginMethod, map[string]interface{}{
		"session_id": session.ID,
		"method":     loginMethod,
	})

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtExpiry.Seconds()),
		User:         user,
		SessionToken: sessionToken,
	}, nil
}

// generateJWT génère un token JWT
func (s *authService) generateJWT(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  string(user.Role),
		"exp":   time.Now().Add(s.jwtExpiry).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "veza-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateRefreshToken génère un refresh token sécurisé
func (s *authService) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

// generateRecoveryCodes génère des codes de récupération 2FA
func (s *authService) generateRecoveryCodes() ([]string, error) {
	codes := make([]string, 8)
	for i := range codes {
		bytes := make([]byte, 6)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		codes[i] = base32.StdEncoding.EncodeToString(bytes)[:8]
	}
	return codes, nil
}

// auditLog enregistre un événement d'audit
func (s *authService) auditLog(ctx context.Context, userID int64, action, resource string, details map[string]interface{}) {
	log := &repositories.UserAuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   fmt.Sprintf("%v", details),
		Success:   true,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreateAuditLog(ctx, log); err != nil {
		s.logger.Error("Erreur création audit log", zap.Error(err))
	}
}

// fetchOAuthUserInfo récupère les informations utilisateur OAuth2
func (s *authService) fetchOAuthUserInfo(ctx context.Context, provider, accessToken string) (*OAuthUserInfo, error) {
	userInfo, err := s.oauthService.FetchOAuthUserInfo(ctx, provider, accessToken)
	if err != nil {
		s.logger.Error("Erreur récupération infos OAuth2",
			zap.String("provider", provider),
			zap.Error(err))
		return nil, fmt.Errorf("échec récupération infos %s: %w", provider, err)
	}

	s.logger.Info("Infos utilisateur OAuth2 récupérées",
		zap.String("provider", provider),
		zap.String("email", userInfo.Email),
		zap.String("username", userInfo.Username))

	return userInfo, nil
}

// verifyRecoveryCode vérifie un code de récupération 2FA
func (s *authService) verifyRecoveryCode(ctx context.Context, userID int64, code string) error {
	// TODO: Implémenter la vérification des codes de récupération
	return ErrInvalidTwoFactorCode
}

// ============================================================================
// TYPES ET STRUCTURES
// ============================================================================

// AuthResponse réponse d'authentification
type AuthResponse struct {
	AccessToken       string         `json:"access_token"`
	RefreshToken      string         `json:"refresh_token"`
	ExpiresIn         int64          `json:"expires_in"`
	TokenType         string         `json:"token_type"`
	User              *entities.User `json:"user"`
	SessionToken      string         `json:"session_token,omitempty"`
	RequiresTwoFactor bool           `json:"requires_two_factor,omitempty"`
	UserID            int64          `json:"user_id,omitempty"`
}

// RegisterRequest requête d'inscription
type RegisterRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=30"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	FirstName   string `json:"first_name" validate:"required,min=1,max=50"`
	LastName    string `json:"last_name" validate:"required,min=1,max=50"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

// Validate valide la requête d'inscription
func (r *RegisterRequest) Validate() error {
	if len(r.Username) < 3 {
		return ErrInvalidUsername
	}
	if len(r.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// TwoFactorSetup configuration 2FA
type TwoFactorSetup struct {
	Secret        string   `json:"secret"`
	QRCode        string   `json:"qr_code"`
	RecoveryCodes []string `json:"recovery_codes"`
}

// OAuthUserInfo informations utilisateur OAuth2
type OAuthUserInfo struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
}

// EmailService interface pour l'envoi d'emails
type EmailService interface {
	SendPasswordReset(ctx context.Context, email, token string) error
	SendVerificationEmail(ctx context.Context, email, token string) error
}

// Erreurs du service d'authentification
var (
	ErrInvalidCredentials      = fmt.Errorf("identifiants invalides")
	ErrAccountInactive         = fmt.Errorf("compte inactif")
	ErrAccountLocked           = fmt.Errorf("compte verrouillé")
	ErrEmailAlreadyExists      = fmt.Errorf("email déjà utilisé")
	ErrUsernameAlreadyExists   = fmt.Errorf("nom d'utilisateur déjà utilisé")
	ErrInvalidToken            = fmt.Errorf("token invalide")
	ErrTokenExpired            = fmt.Errorf("token expiré")
	ErrUnsupportedProvider     = fmt.Errorf("fournisseur OAuth2 non supporté")
	ErrUserNotFound            = fmt.Errorf("utilisateur non trouvé")
	ErrTwoFactorAlreadyEnabled = fmt.Errorf("2FA déjà activé")
	ErrInvalidTwoFactorCode    = fmt.Errorf("code 2FA invalide")
	ErrInvalidPassword         = fmt.Errorf("mot de passe invalide")
	ErrInvalidUsername         = fmt.Errorf("nom d'utilisateur invalide")
	ErrPasswordTooShort        = fmt.Errorf("mot de passe trop court")
)
