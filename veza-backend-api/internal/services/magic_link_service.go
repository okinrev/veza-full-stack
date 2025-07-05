package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// Stubs pour les services manquants
type EmailService interface {
	SendMagicLink(ctx context.Context, email string, link *MagicLink) error
}

type AuditService interface {
	LogSecurityEvent(ctx context.Context, event *AuditEvent) error
}

type AuditEvent struct {
	UserID   *int64                 `json:"user_id,omitempty"`
	Action   string                 `json:"action"`
	Resource string                 `json:"resource"`
	Details  map[string]interface{} `json:"details"`
	Success  bool                   `json:"success"`
}

// MagicLinkService service pour l'authentification par magic links
type MagicLinkService interface {
	// Génération de magic links
	GenerateMagicLink(ctx context.Context, email string, purpose MagicLinkPurpose) (*MagicLink, error)

	// Validation et utilisation
	ValidateMagicLink(ctx context.Context, token string) (*MagicLinkValidation, error)
	ConsumeMagicLink(ctx context.Context, token string) (*entities.User, error)

	// Gestion
	RevokeMagicLink(ctx context.Context, token string) error
	RevokeAllUserMagicLinks(ctx context.Context, userID int64) error

	// Monitoring
	GetMagicLinkStats(ctx context.Context) (*MagicLinkStats, error)
	CleanupExpiredLinks(ctx context.Context) (int, error)
}

// MagicLinkPurpose types de magic links
type MagicLinkPurpose string

const (
	PurposeLogin          MagicLinkPurpose = "login"
	PurposeRegistration   MagicLinkPurpose = "registration"
	PurposePasswordReset  MagicLinkPurpose = "password_reset"
	PurposeEmailVerify    MagicLinkPurpose = "email_verify"
	PurposeAccountRecover MagicLinkPurpose = "account_recover"
)

// MagicLink représente un magic link
type MagicLink struct {
	Token     string           `json:"token"`
	Email     string           `json:"email"`
	UserID    *int64           `json:"user_id,omitempty"`
	Purpose   MagicLinkPurpose `json:"purpose"`
	ExpiresAt time.Time        `json:"expires_at"`
	CreatedAt time.Time        `json:"created_at"`
	IPAddress string           `json:"ip_address"`
	UserAgent string           `json:"user_agent"`
	URL       string           `json:"url"`
}

// MagicLinkValidation résultat de validation d'un magic link
type MagicLinkValidation struct {
	Valid      bool             `json:"valid"`
	Email      string           `json:"email"`
	UserID     *int64           `json:"user_id,omitempty"`
	Purpose    MagicLinkPurpose `json:"purpose"`
	ExpiresAt  time.Time        `json:"expires_at"`
	ConsumedAt *time.Time       `json:"consumed_at,omitempty"`
	IPAddress  string           `json:"ip_address"`
}

// MagicLinkStats statistiques des magic links
type MagicLinkStats struct {
	ActiveLinks    int64 `json:"active_links"`
	TotalGenerated int64 `json:"total_generated"`
	TotalConsumed  int64 `json:"total_consumed"`
	TotalExpired   int64 `json:"total_expired"`
	TotalRevoked   int64 `json:"total_revoked"`

	// Stats par purpose
	StatsByPurpose map[MagicLinkPurpose]int64 `json:"stats_by_purpose"`

	// Performance
	AvgGenerationTimeMs int64 `json:"avg_generation_time_ms"`
	AvgValidationTimeMs int64 `json:"avg_validation_time_ms"`
}

// MagicLinkConfig configuration du service
type MagicLinkConfig struct {
	DefaultTTL       time.Duration                      `json:"default_ttl"`
	TTLByPurpose     map[MagicLinkPurpose]time.Duration `json:"ttl_by_purpose"`
	TokenLength      int                                `json:"token_length"`
	BaseURL          string                             `json:"base_url"`
	MaxActivePerUser int                                `json:"max_active_per_user"`
	RateLimitPerHour int                                `json:"rate_limit_per_hour"`
	SecureTransport  bool                               `json:"secure_transport"`
	HashingAlgorithm string                             `json:"hashing_algorithm"`
}

// magicLinkService implémentation du service
type magicLinkService struct {
	redis        *redis.Client
	userRepo     repositories.UserRepository
	emailService EmailService
	auditService AuditService
	logger       *zap.Logger
	config       *MagicLinkConfig
}

// NewMagicLinkService crée une nouvelle instance du service
func NewMagicLinkService(
	redisClient *redis.Client,
	userRepo repositories.UserRepository,
	emailService EmailService,
	auditService AuditService,
	logger *zap.Logger,
) MagicLinkService {
	config := &MagicLinkConfig{
		DefaultTTL:       15 * time.Minute,
		TokenLength:      32,
		MaxActivePerUser: 3,
		RateLimitPerHour: 5,
		SecureTransport:  true,
		HashingAlgorithm: "sha256",
		TTLByPurpose: map[MagicLinkPurpose]time.Duration{
			PurposeLogin:          15 * time.Minute,
			PurposeRegistration:   1 * time.Hour,
			PurposePasswordReset:  30 * time.Minute,
			PurposeEmailVerify:    24 * time.Hour,
			PurposeAccountRecover: 1 * time.Hour,
		},
	}

	return &magicLinkService{
		redis:        redisClient,
		userRepo:     userRepo,
		emailService: emailService,
		auditService: auditService,
		logger:       logger,
		config:       config,
	}
}

// GenerateMagicLink génère un nouveau magic link
func (m *magicLinkService) GenerateMagicLink(ctx context.Context, email string, purpose MagicLinkPurpose) (*MagicLink, error) {
	start := time.Now()
	defer func() {
		m.logger.Debug("Magic link generation completed",
			zap.String("email", email),
			zap.String("purpose", string(purpose)),
			zap.Duration("duration", time.Since(start)))
	}()

	// Vérifier le rate limiting
	if err := m.checkRateLimit(ctx, email); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Vérifier l'utilisateur
	user, err := m.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Pour certains purposes, l'utilisateur doit exister
	if purpose == PurposeLogin || purpose == PurposePasswordReset {
		if user == nil {
			return nil, fmt.Errorf("user not found")
		}
	}

	// Nettoyer les anciens links de l'utilisateur
	if user != nil {
		if err := m.cleanupUserOldLinks(ctx, user.ID, purpose); err != nil {
			m.logger.Warn("Failed to cleanup old links", zap.Error(err))
		}
	}

	// Générer le token
	token, err := m.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Déterminer TTL
	ttl := m.getTTLForPurpose(purpose)

	// Créer le magic link
	magicLink := &MagicLink{
		Token:     token,
		Email:     email,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
		URL:       m.buildMagicLinkURL(token),
	}

	if user != nil {
		magicLink.UserID = &user.ID
	}

	// Stocker en Redis
	if err := m.storeMagicLink(ctx, magicLink); err != nil {
		return nil, fmt.Errorf("failed to store magic link: %w", err)
	}

	// Envoyer l'email
	if err := m.sendMagicLinkEmail(ctx, magicLink); err != nil {
		m.logger.Error("Failed to send magic link email", zap.Error(err))
		// Ne pas échouer complètement si l'email ne part pas
	}

	// Audit log
	if err := m.auditService.LogSecurityEvent(ctx, &AuditEvent{
		UserID:   magicLink.UserID,
		Action:   "magic_link_generated",
		Resource: "auth",
		Details: map[string]interface{}{
			"email":   email,
			"purpose": string(purpose),
		},
		Success: true,
	}); err != nil {
		m.logger.Error("Failed to log security event", zap.Error(err))
	}

	// Incrémenter les statistiques
	m.incrementStats(ctx, "generated", purpose)

	return magicLink, nil
}

// ValidateMagicLink valide un magic link sans le consommer
func (m *magicLinkService) ValidateMagicLink(ctx context.Context, token string) (*MagicLinkValidation, error) {
	start := time.Now()
	defer func() {
		m.logger.Debug("Magic link validation completed",
			zap.String("token", m.hashToken(token)),
			zap.Duration("duration", time.Since(start)))
	}()

	// Récupérer le magic link
	magicLink, err := m.getMagicLink(ctx, token)
	if err != nil {
		return &MagicLinkValidation{Valid: false}, nil
	}

	// Vérifier l'expiration
	if time.Now().After(magicLink.ExpiresAt) {
		m.incrementStats(ctx, "expired", magicLink.Purpose)
		return &MagicLinkValidation{Valid: false}, nil
	}

	// Vérifier si déjà consommé
	consumed, err := m.isTokenConsumed(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to check consumption status: %w", err)
	}

	validation := &MagicLinkValidation{
		Valid:     !consumed,
		Email:     magicLink.Email,
		UserID:    magicLink.UserID,
		Purpose:   magicLink.Purpose,
		ExpiresAt: magicLink.ExpiresAt,
		IPAddress: magicLink.IPAddress,
	}

	if consumed {
		consumedAt, _ := m.getConsumptionTime(ctx, token)
		validation.ConsumedAt = consumedAt
	}

	// Logger l'événement de sécurité
	if err := m.auditService.LogSecurityEvent(ctx, &AuditEvent{
		UserID:   magicLink.UserID,
		Action:   "magic_link_verified",
		Resource: "auth",
		Details: map[string]interface{}{
			"token": token,
		},
		Success: true,
	}); err != nil {
		m.logger.Error("Failed to log security event", zap.Error(err))
	}

	return validation, nil
}

// ConsumeMagicLink valide et consomme un magic link
func (m *magicLinkService) ConsumeMagicLink(ctx context.Context, token string) (*entities.User, error) {
	start := time.Now()
	defer func() {
		m.logger.Debug("Magic link consumption completed",
			zap.String("token", m.hashToken(token)),
			zap.Duration("duration", time.Since(start)))
	}()

	// Valider d'abord
	validation, err := m.ValidateMagicLink(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if !validation.Valid {
		return nil, fmt.Errorf("invalid or expired magic link")
	}

	// Marquer comme consommé atomiquement
	consumed, err := m.markAsConsumed(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to mark as consumed: %w", err)
	}

	if !consumed {
		return nil, fmt.Errorf("magic link already consumed")
	}

	// Récupérer l'utilisateur
	user, err := m.userRepo.GetByID(ctx, *validation.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Vérifier que l'email correspond
	if user.Email != validation.Email {
		// Log l'incohérence mais ne pas échouer
		m.logger.Warn("Email mismatch in magic link consumption",
			zap.String("link_email", validation.Email),
			zap.String("user_email", user.Email))
	}

	// Audit log
	if err := m.auditService.LogSecurityEvent(ctx, &AuditEvent{
		UserID:   validation.UserID,
		Action:   "magic_link_consumed",
		Resource: "auth",
		Details: map[string]interface{}{
			"email":   validation.Email,
			"purpose": validation.Purpose,
		},
		Success: true,
	}); err != nil {
		m.logger.Error("Failed to log security event", zap.Error(err))
	}

	// Incrémenter les statistiques
	m.incrementStats(ctx, "consumed", validation.Purpose)

	// Logger l'événement de sécurité
	if err := m.auditService.LogSecurityEvent(ctx, &AuditEvent{
		UserID:   &user.ID,
		Action:   "magic_link_expired",
		Resource: "auth",
		Details: map[string]interface{}{
			"token": token,
		},
		Success: true,
	}); err != nil {
		m.logger.Error("Failed to log security event", zap.Error(err))
	}

	return user, nil
}

// RevokeMagicLink révoque un magic link spécifique
func (m *magicLinkService) RevokeMagicLink(ctx context.Context, token string) error {
	hashedToken := m.hashToken(token)

	// Supprimer de Redis
	err := m.redis.Del(ctx, fmt.Sprintf("magic_link:%s", hashedToken)).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke magic link: %w", err)
	}

	// Ajouter à la liste des révoqués
	err = m.redis.Set(ctx, fmt.Sprintf("magic_link_revoked:%s", hashedToken), time.Now().Unix(), 24*time.Hour).Err()
	if err != nil {
		m.logger.Warn("Failed to mark as revoked", zap.Error(err))
	}

	m.incrementStats(ctx, "revoked", "")

	return nil
}

// RevokeAllUserMagicLinks révoque tous les magic links d'un utilisateur
func (m *magicLinkService) RevokeAllUserMagicLinks(ctx context.Context, userID int64) error {
	// Récupérer tous les tokens de l'utilisateur
	pattern := fmt.Sprintf("magic_link_user:%d:*", userID)
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Supprimer tous les tokens
	if len(keys) > 0 {
		err = m.redis.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete user tokens: %w", err)
		}
	}

	// Audit log
	if err := m.auditService.LogSecurityEvent(ctx, &AuditEvent{
		UserID:   &userID,
		Action:   "magic_links_revoked_all",
		Resource: "auth",
		Details: map[string]interface{}{
			"revoked_count": len(keys),
		},
		Success: true,
	}); err != nil {
		m.logger.Error("Failed to log security event", zap.Error(err))
	}

	return nil
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

// generateSecureToken génère un token sécurisé
func (m *magicLinkService) generateSecureToken() (string, error) {
	bytes := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken hash un token pour le stockage
func (m *magicLinkService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// storeMagicLink stocke un magic link en Redis
func (m *magicLinkService) storeMagicLink(ctx context.Context, link *MagicLink) error {
	hashedToken := m.hashToken(link.Token)

	// Stocker le magic link
	key := fmt.Sprintf("magic_link:%s", hashedToken)
	ttl := time.Until(link.ExpiresAt)

	linkData := map[string]interface{}{
		"email":      link.Email,
		"purpose":    link.Purpose,
		"expires_at": link.ExpiresAt.Unix(),
		"created_at": link.CreatedAt.Unix(),
		"ip_address": link.IPAddress,
		"user_agent": link.UserAgent,
	}

	if link.UserID != nil {
		linkData["user_id"] = *link.UserID
	}

	err := m.redis.HMSet(ctx, key, linkData).Err()
	if err != nil {
		return err
	}

	err = m.redis.Expire(ctx, key, ttl).Err()
	if err != nil {
		return err
	}

	// Indexer par utilisateur si applicable
	if link.UserID != nil {
		userKey := fmt.Sprintf("magic_link_user:%d:%s", *link.UserID, hashedToken)
		err = m.redis.Set(ctx, userKey, hashedToken, ttl).Err()
		if err != nil {
			m.logger.Warn("Failed to index by user", zap.Error(err))
		}
	}

	return nil
}

// getMagicLink récupère un magic link depuis Redis
func (m *magicLinkService) getMagicLink(ctx context.Context, token string) (*MagicLink, error) {
	hashedToken := m.hashToken(token)
	key := fmt.Sprintf("magic_link:%s", hashedToken)

	data, err := m.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("magic link not found")
	}

	link := &MagicLink{
		Token:   token,
		Email:   data["email"],
		Purpose: MagicLinkPurpose(data["purpose"]),
	}

	// TODO: Parse user ID if exists
	// if userIDStr, exists := data["user_id"]; exists {
	// 	// Parse user ID
	// 	// Implementation needed
	// }

	// Parse timestamps
	// Implementation needed

	return link, nil
}

// buildMagicLinkURL construit l'URL du magic link
func (m *magicLinkService) buildMagicLinkURL(token string) string {
	return fmt.Sprintf("%s/auth/magic-link?token=%s", m.config.BaseURL, token)
}

// sendMagicLinkEmail envoie l'email avec le magic link
func (m *magicLinkService) sendMagicLinkEmail(ctx context.Context, link *MagicLink) error {
	// Implementation avec le service email
	return m.emailService.SendMagicLink(ctx, link.Email, link)
}

// checkRateLimit vérifie le rate limiting
func (m *magicLinkService) checkRateLimit(ctx context.Context, email string) error {
	key := fmt.Sprintf("magic_link_rate:%s", email)
	count, err := m.redis.Incr(ctx, key).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		m.redis.Expire(ctx, key, time.Hour)
	}

	if count > int64(m.config.RateLimitPerHour) {
		return fmt.Errorf("rate limit exceeded: %d requests in the last hour", count)
	}

	return nil
}

// getTTLForPurpose retourne le TTL pour un purpose donné
func (m *magicLinkService) getTTLForPurpose(purpose MagicLinkPurpose) time.Duration {
	if ttl, exists := m.config.TTLByPurpose[purpose]; exists {
		return ttl
	}
	return m.config.DefaultTTL
}

// cleanupUserOldLinks nettoie les anciens liens d'un utilisateur
func (m *magicLinkService) cleanupUserOldLinks(ctx context.Context, userID int64, purpose MagicLinkPurpose) error {
	pattern := fmt.Sprintf("magic_link_user:%d:*", userID)
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	// Limiter le nombre de liens actifs
	if len(keys) >= m.config.MaxActivePerUser {
		// Supprimer les plus anciens
		oldestKeys := keys[:len(keys)-m.config.MaxActivePerUser+1]
		if len(oldestKeys) > 0 {
			m.redis.Del(ctx, oldestKeys...)
		}
	}

	return nil
}

// isTokenConsumed vérifie si un token a été consommé
func (m *magicLinkService) isTokenConsumed(ctx context.Context, token string) (bool, error) {
	hashedToken := m.hashToken(token)
	key := fmt.Sprintf("magic_link_consumed:%s", hashedToken)

	exists, err := m.redis.Exists(ctx, key).Result()
	return exists > 0, err
}

// markAsConsumed marque un token comme consommé
func (m *magicLinkService) markAsConsumed(ctx context.Context, token string) (bool, error) {
	hashedToken := m.hashToken(token)
	key := fmt.Sprintf("magic_link_consumed:%s", hashedToken)

	// Utiliser SET NX pour éviter les double consommations
	result, err := m.redis.SetNX(ctx, key, time.Now().Unix(), 24*time.Hour).Result()
	return result, err
}

// getConsumptionTime récupère le timestamp de consommation
func (m *magicLinkService) getConsumptionTime(ctx context.Context, token string) (*time.Time, error) {
	hashedToken := m.hashToken(token)
	key := fmt.Sprintf("magic_link_consumed:%s", hashedToken)

	timestamp, err := m.redis.Get(ctx, key).Int64()
	if err != nil {
		return nil, err
	}

	consumedAt := time.Unix(timestamp, 0)
	return &consumedAt, nil
}

// incrementStats incrémente les statistiques
func (m *magicLinkService) incrementStats(ctx context.Context, action string, purpose MagicLinkPurpose) {
	// Statistiques générales
	m.redis.Incr(ctx, fmt.Sprintf("magic_link_stats:total_%s", action))

	// Statistiques par purpose
	if purpose != "" {
		m.redis.Incr(ctx, fmt.Sprintf("magic_link_stats:%s_%s", purpose, action))
	}

	// Statistiques quotidiennes
	today := time.Now().Format("2006-01-02")
	m.redis.Incr(ctx, fmt.Sprintf("magic_link_stats:daily:%s:%s", today, action))
	m.redis.Expire(ctx, fmt.Sprintf("magic_link_stats:daily:%s:%s", today, action), 7*24*time.Hour)
}

// GetMagicLinkStats retourne les statistiques
func (m *magicLinkService) GetMagicLinkStats(ctx context.Context) (*MagicLinkStats, error) {
	stats := &MagicLinkStats{
		StatsByPurpose: make(map[MagicLinkPurpose]int64),
	}

	// Récupérer les statistiques principales
	stats.TotalGenerated, _ = m.redis.Get(ctx, "magic_link_stats:total_generated").Int64()
	stats.TotalConsumed, _ = m.redis.Get(ctx, "magic_link_stats:total_consumed").Int64()
	stats.TotalExpired, _ = m.redis.Get(ctx, "magic_link_stats:total_expired").Int64()
	stats.TotalRevoked, _ = m.redis.Get(ctx, "magic_link_stats:total_revoked").Int64()

	// Calculer les liens actifs
	activePattern := "magic_link:*"
	activeKeys, _ := m.redis.Keys(ctx, activePattern).Result()
	stats.ActiveLinks = int64(len(activeKeys))

	return stats, nil
}

// CleanupExpiredLinks nettoie les liens expirés
func (m *magicLinkService) CleanupExpiredLinks(ctx context.Context) (int, error) {
	pattern := "magic_link:*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	cleaned := 0
	for _, key := range keys {
		ttl, err := m.redis.TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		if ttl < 0 { // Expiré
			m.redis.Del(ctx, key)
			cleaned++
		}
	}

	return cleaned, nil
}

// EmailMagicLink structure pour l'email
type EmailMagicLink struct {
	Email     string    `json:"email"`
	URL       string    `json:"url"`
	Purpose   string    `json:"purpose"`
	ExpiresAt time.Time `json:"expires_at"`
}
