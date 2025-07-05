package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

// MagicLinkService service pour l'authentification par liens magiques
type MagicLinkService struct {
	db          *database.DB
	emailSender EmailSender
	baseURL     string
	logger      *slog.Logger
}

// EmailSender interface pour l'envoi d'emails
type EmailSender interface {
	SendMagicLink(email, token, loginURL string) error
}

// NewMagicLinkService crée une nouvelle instance du service Magic Link
func NewMagicLinkService(db *database.DB, emailSender EmailSender, baseURL string, logger *slog.Logger) *MagicLinkService {
	return &MagicLinkService{
		db:          db,
		emailSender: emailSender,
		baseURL:     baseURL,
		logger:      logger,
	}
}

// MagicLinkRequest structure de requête Magic Link
type MagicLinkRequest struct {
	Email       string `json:"email" binding:"required,email"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// MagicLinkValidation structure de validation Magic Link
type MagicLinkValidation struct {
	Token       string `json:"token" binding:"required"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// MagicLink structure interne du Magic Link
type MagicLink struct {
	ID          int        `db:"id"`
	UserID      int64      `db:"user_id"`
	Email       string     `db:"email"`
	Token       string     `db:"token"`
	RedirectURL string     `db:"redirect_url"`
	ExpiresAt   time.Time  `db:"expires_at"`
	UsedAt      *time.Time `db:"used_at"`
	CreatedAt   time.Time  `db:"created_at"`
	IPAddress   string     `db:"ip_address"`
	UserAgent   string     `db:"user_agent"`
}

// SendMagicLink génère et envoie un lien magique
func (s *MagicLinkService) SendMagicLink(email, redirectURL, ipAddress, userAgent string) error {
	// Normaliser l'email
	email = normalizeEmail(email)

	// Vérifier si l'utilisateur existe
	user, err := s.getUserByEmail(email)
	if err != nil {
		// Pour la sécurité, on ne révèle pas si l'email existe ou non
		return nil // Succès apparent même si l'utilisateur n'existe pas
	}

	// Nettoyer les anciens tokens expirés
	s.cleanupExpiredTokens(email)

	// Vérifier les limites de rate limiting
	if s.isRateLimited(email) {
		return fmt.Errorf("trop de tentatives récentes. Veuillez patienter avant de réessayer")
	}

	// Générer un token sécurisé
	token, err := s.generateSecureToken()
	if err != nil {
		return fmt.Errorf("erreur génération token: %w", err)
	}

	// Valider et nettoyer l'URL de redirection
	cleanRedirectURL := s.validateRedirectURL(redirectURL)

	// Sauvegarder le Magic Link
	expiresAt := time.Now().Add(15 * time.Minute) // Expire dans 15 minutes

	_, err = s.db.Exec(`
		INSERT INTO magic_links (user_id, email, token, redirect_url, expires_at, created_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, email, token, cleanRedirectURL, expiresAt, time.Now(), ipAddress, userAgent)

	if err != nil {
		return fmt.Errorf("erreur sauvegarde Magic Link: %w", err)
	}

	// Construire l'URL de connexion
	loginURL := s.buildLoginURL(token, cleanRedirectURL)

	// Envoyer l'email
	err = s.emailSender.SendMagicLink(email, token, loginURL)
	if err != nil {
		return fmt.Errorf("erreur envoi email: %w", err)
	}

	return nil
}

// ValidateMagicLink valide un token Magic Link et connecte l'utilisateur
func (s *MagicLinkService) ValidateMagicLink(token, ipAddress string) (*LoginResponse, error) {
	// Récupérer le Magic Link
	magicLink, err := s.getMagicLinkByToken(token)
	if err != nil {
		return nil, fmt.Errorf("token Magic Link invalide ou expiré")
	}

	// Vérifier l'expiration
	if time.Now().After(magicLink.ExpiresAt) {
		return nil, fmt.Errorf("Magic Link expiré")
	}

	// Vérifier qu'il n'a pas déjà été utilisé
	if magicLink.UsedAt != nil {
		return nil, fmt.Errorf("Magic Link déjà utilisé")
	}

	// Récupérer l'utilisateur
	user, err := s.getUserByID(magicLink.UserID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}

	// Vérifier que le compte est actif
	if user.Role == "deleted" || user.Role == "banned" {
		return nil, fmt.Errorf("compte inactif")
	}

	// Marquer le Magic Link comme utilisé
	now := time.Now()
	_, err = s.db.Exec(`
		UPDATE magic_links 
		SET used_at = $1, used_ip_address = $2 
		WHERE token = $3
	`, now, ipAddress, token)

	if err != nil {
		return nil, fmt.Errorf("erreur mise à jour Magic Link: %w", err)
	}

	// Générer les tokens d'authentification
	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Username, user.Role, "your-jwt-secret") // TODO: Injecter le secret
	if err != nil {
		return nil, fmt.Errorf("erreur génération tokens: %w", err)
	}

	// Mettre à jour la dernière connexion
	if _, err := s.db.Exec("UPDATE users SET updated_at = NOW() WHERE id = $1", user.ID); err != nil {
		s.logger.Error("Failed to update user timestamp", "error", err)
	}

	// Enregistrer l'événement de connexion pour audit
	s.recordLoginEvent(int(user.ID), "magic_link", ipAddress, true)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    expiresIn,
	}, nil
}

// GetMagicLinkStatus vérifie le statut d'un Magic Link
func (s *MagicLinkService) GetMagicLinkStatus(token string) (*MagicLinkStatus, error) {
	magicLink, err := s.getMagicLinkByToken(token)
	if err != nil {
		return &MagicLinkStatus{
			Valid:   false,
			Message: "Token invalide",
		}, nil
	}

	status := &MagicLinkStatus{
		Valid:       true,
		Expired:     time.Now().After(magicLink.ExpiresAt),
		Used:        magicLink.UsedAt != nil,
		ExpiresAt:   magicLink.ExpiresAt,
		RedirectURL: magicLink.RedirectURL,
	}

	if status.Expired {
		status.Message = "Token expiré"
		status.Valid = false
	} else if status.Used {
		status.Message = "Token déjà utilisé"
		status.Valid = false
	} else {
		status.Message = "Token valide"
	}

	return status, nil
}

// MagicLinkStatus statut d'un Magic Link
type MagicLinkStatus struct {
	Valid       bool      `json:"valid"`
	Expired     bool      `json:"expired"`
	Used        bool      `json:"used"`
	ExpiresAt   time.Time `json:"expires_at"`
	RedirectURL string    `json:"redirect_url"`
	Message     string    `json:"message"`
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

// getUserByEmail récupère un utilisateur par email
func (s *MagicLinkService) getUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, role, created_at, updated_at
		FROM users WHERE email = $1 AND role != 'deleted'
	`, email).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	return &user, err
}

// getUserByID récupère un utilisateur par ID
func (s *MagicLinkService) getUserByID(userID int64) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	return &user, err
}

// getMagicLinkByToken récupère un Magic Link par token
func (s *MagicLinkService) getMagicLinkByToken(token string) (*MagicLink, error) {
	var link MagicLink
	err := s.db.QueryRow(`
		SELECT id, user_id, email, token, redirect_url, expires_at, used_at, created_at, ip_address, user_agent
		FROM magic_links WHERE token = $1
	`, token).Scan(
		&link.ID, &link.UserID, &link.Email, &link.Token,
		&link.RedirectURL, &link.ExpiresAt, &link.UsedAt,
		&link.CreatedAt, &link.IPAddress, &link.UserAgent,
	)
	return &link, err
}

// generateSecureToken génère un token sécurisé
func (s *MagicLinkService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

// buildLoginURL construit l'URL de connexion
func (s *MagicLinkService) buildLoginURL(token, redirectURL string) string {
	baseURL := s.baseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	loginURL := fmt.Sprintf("%s/api/v1/auth/magic-link/verify?token=%s", baseURL, url.QueryEscape(token))

	if redirectURL != "" {
		loginURL += "&redirect_url=" + url.QueryEscape(redirectURL)
	}

	return loginURL
}

// validateRedirectURL valide et nettoie l'URL de redirection
func (s *MagicLinkService) validateRedirectURL(redirectURL string) string {
	if redirectURL == "" {
		return ""
	}

	// Parse l'URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return ""
	}

	// Autoriser seulement certains domaines (whitelist)
	allowedHosts := []string{
		"localhost",
		"127.0.0.1",
		"veza.dev",
		"app.veza.dev",
	}

	allowed := false
	for _, host := range allowedHosts {
		if parsedURL.Host == host || parsedURL.Host == host+":3000" || parsedURL.Host == host+":5173" {
			allowed = true
			break
		}
	}

	if !allowed {
		return ""
	}

	return redirectURL
}

// isRateLimited vérifie les limites de taux
func (s *MagicLinkService) isRateLimited(email string) bool {
	var count int
	// Maximum 3 Magic Links par heure
	if err := s.db.QueryRow(`
		SELECT COUNT(*) FROM magic_links 
		WHERE email = $1 AND created_at > $2
	`, email, time.Now().Add(-1*time.Hour)).Scan(&count); err != nil {
		s.logger.Error("Failed to count recent attempts", "error", err)
		return false
	}

	return count >= 3
}

// cleanupExpiredTokens nettoie les tokens expirés
func (s *MagicLinkService) cleanupExpiredTokens(email string) {
	// Nettoyer les anciennes tentatives
	if _, err := s.db.Exec(`
		DELETE FROM magic_link_attempts 
		WHERE created_at < NOW() - INTERVAL '24 hours'
	`); err != nil {
		s.logger.Error("Failed to clean old attempts", "error", err)
	}

	// Nettoyer les tokens expirés
	if _, err := s.db.Exec(`
		DELETE FROM magic_link_tokens 
		WHERE expires_at < NOW()
	`); err != nil {
		s.logger.Error("Failed to clean expired tokens", "error", err)
	}
}

// recordLoginEvent enregistre un événement de connexion
func (s *MagicLinkService) recordLoginEvent(userID int, method, ipAddress string, success bool) {
	s.db.Exec(`
		INSERT INTO login_events (user_id, method, ip_address, success, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, method, ipAddress, success, time.Now())
}

// normalizeEmail normalise une adresse email
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
