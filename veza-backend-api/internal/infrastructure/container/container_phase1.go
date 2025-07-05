package container

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/services"
	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
	"github.com/okinrev/veza-web-app/internal/ports/http"
)

// ContainerPhase1 container simplifié pour la Phase 1
type ContainerPhase1 struct {
	// Configuration
	Config *config.Config

	// Infrastructure
	Logger *zap.Logger

	// Services (domain) - mocks pour Phase 1
	AuthService services.AuthService

	// Handlers (ports)
	AuthHandler http.AuthHandler
}

// NewPhase1 crée un container simplifié pour la Phase 1
func NewPhase1(cfg *config.Config) (*ContainerPhase1, error) {
	container := &ContainerPhase1{
		Config: cfg,
	}

	// Logger simple
	var err error
	if cfg.IsDevelopment() {
		container.Logger, err = zap.NewDevelopment()
	} else {
		container.Logger, err = zap.NewProduction()
	}
	if err != nil {
		return nil, fmt.Errorf("erreur initialisation logger: %w", err)
	}

	// Service Auth simple pour Phase 1 - TODO: implémenter NewAuthServiceMock
	// container.AuthService = services.NewAuthServiceMock()
	container.AuthService = nil // Temporairement nil

	// Handler Auth - TODO: adapter la signature
	// container.AuthHandler, err = http.NewAuthHandler(container.AuthService, container.Logger)
	// container.AuthHandler = &http.AuthHandler{} // Stub temporaire
	// TODO: Implémenter AuthHandler

	log.Printf("✅ Container Phase 1 initialisé avec succès")
	return container, nil
}

// Close nettoie les ressources
func (c *ContainerPhase1) Close() {
	if c.Logger != nil {
		c.Logger.Sync()
	}
}
