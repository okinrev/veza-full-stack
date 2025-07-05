package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// SecretRotationService gère la rotation automatique des secrets JWT
type SecretRotationService struct {
	config    *JWTConfig
	logger    *zap.Logger
	mutex     sync.RWMutex
	stopChan  chan struct{}
	isRunning bool
}

// JWTConfig configuration pour la rotation des secrets
type JWTConfig struct {
	RotationEnabled    bool
	RotationInterval   time.Duration
	SecretHistorySize  int
	CurrentSecretIndex int
	SecretHistory      []string
	Issuer             string
	SecretLength       int
}

// NewSecretRotationService crée un nouveau service de rotation de secrets
func NewSecretRotationService(config *JWTConfig, logger *zap.Logger) *SecretRotationService {
	service := &SecretRotationService{
		config:   config,
		logger:   logger,
		stopChan: make(chan struct{}),
	}

	// Initialiser l'historique des secrets
	if config.SecretHistorySize > 0 {
		service.config.SecretHistory = make([]string, config.SecretHistorySize)
		for i := 0; i < config.SecretHistorySize; i++ {
			secret, err := service.generateSecureSecret()
			if err != nil {
				logger.Error("Failed to generate initial secret", zap.Error(err))
				continue
			}
			service.config.SecretHistory[i] = secret
		}
	}

	return service
}

// Start démarre la rotation automatique des secrets
func (s *SecretRotationService) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("secret rotation service already running")
	}

	// Réinitialiser le canal stopChan si nécessaire
	select {
	case <-s.stopChan:
		// Canal fermé, le recréer
		s.stopChan = make(chan struct{})
	default:
		// Canal ouvert, le recréer pour être sûr
		s.stopChan = make(chan struct{})
	}

	s.isRunning = true
	s.logger.Info("Secret rotation service started",
		zap.Duration("rotation_interval", s.config.RotationInterval),
		zap.Int("history_size", s.config.SecretHistorySize))

	go s.rotationLoop()

	return nil
}

// Stop arrête la rotation automatique des secrets
func (s *SecretRotationService) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return
	}

	// Éviter de fermer un canal déjà fermé
	select {
	case <-s.stopChan:
		// Canal déjà fermé
	default:
		close(s.stopChan)
	}

	s.isRunning = false
	s.logger.Info("Secret rotation service stopped")
}

// GetCurrentSecret retourne le secret actuel
func (s *SecretRotationService) GetCurrentSecret() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.isRunning || len(s.config.SecretHistory) == 0 {
		return ""
	}

	return s.config.SecretHistory[s.config.CurrentSecretIndex]
}

// GetSecretForValidation retourne tous les secrets valides pour la validation
func (s *SecretRotationService) GetSecretForValidation() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	secrets := make([]string, 0, len(s.config.SecretHistory))
	for _, secret := range s.config.SecretHistory {
		if secret != "" {
			secrets = append(secrets, secret)
		}
	}

	return secrets
}

// rotationLoop boucle principale de rotation des secrets
func (s *SecretRotationService) rotationLoop() {
	ticker := time.NewTicker(s.config.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.rotateSecret()
		case <-s.stopChan:
			return
		}
	}
}

// rotateSecret effectue la rotation du secret
func (s *SecretRotationService) rotateSecret() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Générer un nouveau secret
	newSecret, err := s.generateSecureSecret()
	if err != nil {
		s.logger.Error("Failed to generate new secret", zap.Error(err))
		return
	}

	// Mettre à jour l'index
	s.config.CurrentSecretIndex = (s.config.CurrentSecretIndex + 1) % s.config.SecretHistorySize

	// Remplacer le secret le plus ancien
	s.config.SecretHistory[s.config.CurrentSecretIndex] = newSecret

	s.logger.Info("JWT secret rotated",
		zap.Int("current_index", s.config.CurrentSecretIndex),
		zap.String("new_secret_preview", newSecret[:8]+"..."))

	// Émettre des métriques
	s.emitRotationMetrics()
}

// generateSecureSecret génère un secret sécurisé
func (s *SecretRotationService) generateSecureSecret() (string, error) {
	// Générer un nouveau secret sécurisé
	bytes := make([]byte, s.config.SecretLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure secret: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// emitRotationMetrics émet des métriques de rotation
func (s *SecretRotationService) emitRotationMetrics() {
	// Ici on pourrait émettre des métriques Prometheus
	// Pour l'instant, on se contente de logs
	s.logger.Debug("Secret rotation metrics",
		zap.Int("current_index", s.config.CurrentSecretIndex),
		zap.Int("history_size", len(s.config.SecretHistory)),
		zap.Duration("rotation_interval", s.config.RotationInterval))
}

// GetRotationStatus retourne le statut de la rotation
func (s *SecretRotationService) GetRotationStatus() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]interface{}{
		"enabled":           s.config.RotationEnabled,
		"is_running":        s.isRunning,
		"current_index":     s.config.CurrentSecretIndex,
		"history_size":      len(s.config.SecretHistory),
		"rotation_interval": s.config.RotationInterval.String(),
		"next_rotation":     time.Now().Add(s.config.RotationInterval).Format(time.RFC3339),
	}
}

// ForceRotation force une rotation immédiate du secret
func (s *SecretRotationService) ForceRotation() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.isRunning {
		return fmt.Errorf("secret rotation service not running")
	}

	// Générer un nouveau secret
	newSecret, err := s.generateSecureSecret()
	if err != nil {
		return fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Mettre à jour l'index
	s.config.CurrentSecretIndex = (s.config.CurrentSecretIndex + 1) % s.config.SecretHistorySize

	// Remplacer le secret le plus ancien
	s.config.SecretHistory[s.config.CurrentSecretIndex] = newSecret

	s.logger.Info("JWT secret rotation forced",
		zap.Int("current_index", s.config.CurrentSecretIndex),
		zap.String("new_secret_preview", newSecret[:8]+"..."))

	return nil
}

// ValidateTokenWithHistory valide un token avec l'historique des secrets
func (s *SecretRotationService) ValidateTokenWithHistory(tokenString string) (*ServiceClaims, error) {
	secrets := s.GetSecretForValidation()

	for _, secret := range secrets {
		if claims, err := ValidateJWTWithSecret(tokenString, secret); err == nil {
			return claims, nil
		}
	}

	return nil, fmt.Errorf("token validation failed with all secrets")
}

// ValidateJWTWithSecret valide un JWT avec un secret spécifique
func ValidateJWTWithSecret(tokenString, secret string) (*ServiceClaims, error) {
	// Parse du token avec le secret spécifique
	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*ServiceClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
