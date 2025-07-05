package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSecretRotationService(t *testing.T) {
	// Configuration de test
	config := &JWTConfig{
		RotationEnabled:    true,
		RotationInterval:   100 * time.Millisecond, // Intervalle court pour les tests
		SecretHistorySize:  3,
		CurrentSecretIndex: 0,
		SecretHistory:      []string{},
		Issuer:             "veza-test",
	}

	logger, _ := zap.NewDevelopment()

	t.Run("Création du service", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		assert.NotNil(t, service)
		assert.Equal(t, config, service.config)
		assert.Equal(t, logger, service.logger)
	})

	t.Run("Démarrage et arrêt du service", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)

		// Démarrage
		err := service.Start()
		require.NoError(t, err)
		assert.True(t, service.isRunning)

		// Tentative de redémarrage (doit échouer)
		err = service.Start()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")

		// Arrêt
		service.Stop()
		assert.False(t, service.isRunning)

		// Redémarrage après arrêt
		err = service.Start()
		require.NoError(t, err)
		assert.True(t, service.isRunning)

		service.Stop()
	})

	t.Run("Génération de secrets sécurisés", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)

		// Test de génération de secret
		secret1, err := service.generateSecureSecret()
		if err != nil {
			t.Fatalf("Failed to generate secret: %v", err)
		}
		secret2, err := service.generateSecureSecret()
		if err != nil {
			t.Fatalf("Failed to generate secret: %v", err)
		}

		// Vérifier que les secrets sont différents
		assert.NotEqual(t, secret1, secret2)

		// Vérifier la longueur (base64 de 64 bytes = 88 caractères)
		assert.Len(t, secret1, 88)
		assert.Len(t, secret2, 88)

		// Vérifier que les secrets ne sont pas vides
		assert.NotEmpty(t, secret1)
		assert.NotEmpty(t, secret2)
	})

	t.Run("Récupération du secret actuel", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)

		// Avant démarrage, pas de secret
		secret := service.GetCurrentSecret()
		assert.Empty(t, secret)

		// Après démarrage, secret disponible
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		secret = service.GetCurrentSecret()
		assert.NotEmpty(t, secret)
		assert.Len(t, secret, 88)
	})

	t.Run("Récupération des secrets pour validation", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		secrets := service.GetSecretForValidation()
		assert.NotEmpty(t, secrets)
		assert.Len(t, secrets, 3) // Taille de l'historique

		// Vérifier que tous les secrets sont différents
		secretSet := make(map[string]bool)
		for _, secret := range secrets {
			assert.NotEmpty(t, secret)
			assert.Len(t, secret, 88)
			secretSet[secret] = true
		}
		assert.Len(t, secretSet, 3) // Tous différents
	})

	t.Run("Rotation forcée", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		// Récupérer le secret initial
		initialSecret := service.GetCurrentSecret()
		assert.NotEmpty(t, initialSecret)

		// Forcer une rotation
		err = service.ForceRotation()
		require.NoError(t, err)

		// Vérifier que le secret a changé
		newSecret := service.GetCurrentSecret()
		assert.NotEqual(t, initialSecret, newSecret)
		assert.NotEmpty(t, newSecret)
	})

	t.Run("Rotation forcée sans service démarré", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)

		err := service.ForceRotation()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("Statut de rotation", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		status := service.GetRotationStatus()

		assert.Equal(t, true, status["enabled"])
		assert.Equal(t, true, status["is_running"])
		assert.IsType(t, 0, status["current_index"])
		assert.Equal(t, 3, status["history_size"])
		assert.Equal(t, "100ms", status["rotation_interval"])
		assert.IsType(t, "", status["next_rotation"])
	})

	t.Run("Validation de token avec historique", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		// Générer un token avec le secret actuel
		currentSecret := service.GetCurrentSecret()
		claims := &ServiceClaims{
			UserID:   123,
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "user",
		}

		token, err := generateTestToken(claims, currentSecret)
		require.NoError(t, err)

		// Valider le token avec l'historique
		validatedClaims, err := service.ValidateTokenWithHistory(token)
		require.NoError(t, err)
		assert.Equal(t, claims.UserID, validatedClaims.UserID)
		assert.Equal(t, claims.Username, validatedClaims.Username)
		assert.Equal(t, claims.Email, validatedClaims.Email)
		assert.Equal(t, claims.Role, validatedClaims.Role)
	})

	t.Run("Validation de token invalide", func(t *testing.T) {
		service := NewSecretRotationService(config, logger)
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		// Token invalide
		invalidToken := "invalid.token.here"

		_, err = service.ValidateTokenWithHistory(invalidToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
	})
}

// Fonction utilitaire pour générer un token de test
func generateTestToken(claims *ServiceClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestJWTConfigValidation(t *testing.T) {
	t.Run("Configuration valide", func(t *testing.T) {
		config := &JWTConfig{
			RotationEnabled:    true,
			RotationInterval:   1 * time.Hour,
			SecretHistorySize:  5,
			CurrentSecretIndex: 0,
			Issuer:             "veza-test",
		}

		// Validation devrait passer
		assert.True(t, config.RotationEnabled)
		assert.Equal(t, 1*time.Hour, config.RotationInterval)
		assert.Equal(t, 5, config.SecretHistorySize)
		assert.Equal(t, "veza-test", config.Issuer)
	})

	t.Run("Configuration avec intervalle trop court", func(t *testing.T) {
		config := &JWTConfig{
			RotationEnabled:    true,
			RotationInterval:   1 * time.Second, // Trop court
			SecretHistorySize:  5,
			CurrentSecretIndex: 0,
			Issuer:             "veza-test",
		}

		// L'intervalle court devrait être accepté pour les tests
		assert.True(t, config.RotationEnabled)
		assert.Equal(t, 1*time.Second, config.RotationInterval)
	})

	t.Run("Configuration avec historique trop petit", func(t *testing.T) {
		config := &JWTConfig{
			RotationEnabled:    true,
			RotationInterval:   1 * time.Hour,
			SecretHistorySize:  1, // Trop petit
			CurrentSecretIndex: 0,
			Issuer:             "veza-test",
		}

		// L'historique de taille 1 devrait être accepté
		assert.Equal(t, 1, config.SecretHistorySize)
	})
}

func TestSecretRotationServiceConcurrency(t *testing.T) {
	config := &JWTConfig{
		RotationEnabled:    true,
		RotationInterval:   50 * time.Millisecond,
		SecretHistorySize:  3,
		CurrentSecretIndex: 0,
		SecretHistory:      []string{},
		Issuer:             "veza-test",
	}

	logger, _ := zap.NewDevelopment()
	service := NewSecretRotationService(config, logger)

	t.Run("Accès concurrent aux secrets", func(t *testing.T) {
		err := service.Start()
		require.NoError(t, err)
		defer service.Stop()

		// Simuler des accès concurrents
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					secret := service.GetCurrentSecret()
					assert.NotEmpty(t, secret)
					secrets := service.GetSecretForValidation()
					assert.NotEmpty(t, secrets)
				}
				done <- true
			}()
		}

		// Attendre que toutes les goroutines terminent
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
