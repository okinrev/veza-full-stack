package plugin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PluginGateway orchestrateur principal des plugins
type PluginGateway struct {
	registry    *DiscoveryRegistry
	loader      *PluginLoader
	serviceMesh *ServiceMesh
	config      *GatewayConfig
	logger      *zap.Logger
	plugins     map[string]*PluginInstance
	mutex       sync.RWMutex
}

// GatewayConfig configuration du gateway
type GatewayConfig struct {
	MaxPlugins     int           `json:"max_plugins"`
	PluginTimeout  time.Duration `json:"plugin_timeout"`
	CircuitBreaker CircuitBreakerConfig
	RateLimiting   RateLimitConfig
	FeatureFlags   FeatureFlagConfig
	Discovery      DiscoveryConfig
}

// CircuitBreakerConfig configuration du circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	Timeout          time.Duration `json:"timeout"`
	HalfOpenLimit    int           `json:"half_open_limit"`
}

// RateLimitConfig configuration du rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// FeatureFlagConfig configuration des feature flags
type FeatureFlagConfig struct {
	Provider     string          `json:"provider"` // "growthbook", "launchdarkly"
	Environment  string          `json:"environment"`
	DefaultFlags map[string]bool `json:"default_flags"`
}

// DiscoveryConfig configuration de la découverte
type DiscoveryConfig struct {
	Provider    string `json:"provider"` // "consul", "etcd"
	ServiceName string `json:"service_name"`
	ServicePort int    `json:"service_port"`
	HealthCheck bool   `json:"health_check"`
}

// PluginInstance instance d'un plugin chargé
type PluginInstance struct {
	Info        *PluginInfo
	Plugin      Plugin
	Status      PluginStatus
	Metrics     *PluginMetrics
	LastSeen    time.Time
	HealthCheck *HealthCheck
}

// NewPluginGateway crée un nouveau Plugin Gateway
func NewPluginGateway(config *GatewayConfig, logger *zap.Logger) (*PluginGateway, error) {
	// Initialiser la configuration par défaut
	if config == nil {
		config = &GatewayConfig{
			MaxPlugins:    100,
			PluginTimeout: 30 * time.Second,
			CircuitBreaker: CircuitBreakerConfig{
				FailureThreshold: 5,
				Timeout:          30 * time.Second,
				HalfOpenLimit:    3,
			},
			RateLimiting: RateLimitConfig{
				RequestsPerSecond: 1000,
				BurstSize:         100,
				WindowSize:        1 * time.Second,
			},
			FeatureFlags: FeatureFlagConfig{
				Provider:    "growthbook",
				Environment: "production",
				DefaultFlags: map[string]bool{
					"enable_plugins": true,
				},
			},
			Discovery: DiscoveryConfig{
				Provider:    "consul",
				ServiceName: "veza-plugin-gateway",
				ServicePort: 8080,
				HealthCheck: true,
			},
		}
	}

	// Créer les composants
	registry, err := NewDiscoveryRegistry(config.Discovery, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery registry: %w", err)
	}

	loader, err := NewPluginLoader(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin loader: %w", err)
	}

	serviceMesh, err := NewServiceMesh(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create service mesh: %w", err)
	}

	gateway := &PluginGateway{
		registry:    registry,
		loader:      loader,
		serviceMesh: serviceMesh,
		config:      config,
		logger:      logger,
		plugins:     make(map[string]*PluginInstance),
	}

	logger.Info("🚀 Plugin Gateway initialized",
		zap.Int("max_plugins", config.MaxPlugins),
		zap.Duration("plugin_timeout", config.PluginTimeout),
		zap.String("discovery_provider", config.Discovery.Provider),
	)

	return gateway, nil
}

// Start démarre le Plugin Gateway
func (g *PluginGateway) Start(ctx context.Context) error {
	g.logger.Info("🔄 Starting Plugin Gateway")

	// Démarrer le registry
	if err := g.registry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Démarrer le loader
	if err := g.loader.Start(ctx); err != nil {
		return fmt.Errorf("failed to start loader: %w", err)
	}

	// Démarrer le service mesh
	if err := g.serviceMesh.Start(ctx); err != nil {
		return fmt.Errorf("failed to start service mesh: %w", err)
	}

	// Charger les plugins existants
	if err := g.loadExistingPlugins(ctx); err != nil {
		g.logger.Warn("Failed to load existing plugins", zap.Error(err))
	}

	// Démarrer le monitoring
	go g.monitoringWorker(ctx)

	g.logger.Info("✅ Plugin Gateway started successfully")
	return nil
}

// Stop arrête le Plugin Gateway
func (g *PluginGateway) Stop(ctx context.Context) error {
	g.logger.Info("🛑 Stopping Plugin Gateway")

	// Arrêter tous les plugins
	g.mutex.Lock()
	for pluginID, instance := range g.plugins {
		if err := instance.Plugin.Stop(); err != nil {
			g.logger.Error("Failed to stop plugin", zap.String("plugin_id", pluginID), zap.Error(err))
		}
	}
	g.mutex.Unlock()

	// Arrêter les composants
	if err := g.serviceMesh.Stop(ctx); err != nil {
		g.logger.Error("Failed to stop service mesh", zap.Error(err))
	}

	if err := g.loader.Stop(ctx); err != nil {
		g.logger.Error("Failed to stop loader", zap.Error(err))
	}

	if err := g.registry.Stop(ctx); err != nil {
		g.logger.Error("Failed to stop registry", zap.Error(err))
	}

	g.logger.Info("✅ Plugin Gateway stopped")
	return nil
}

// LoadPlugin charge un nouveau plugin
func (g *PluginGateway) LoadPlugin(ctx context.Context, pluginID string, config PluginConfig) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Vérifier la limite de plugins
	if len(g.plugins) >= g.config.MaxPlugins {
		return fmt.Errorf("plugin limit reached: %d", g.config.MaxPlugins)
	}

	// Vérifier si le plugin existe déjà
	if _, exists := g.plugins[pluginID]; exists {
		return fmt.Errorf("plugin already loaded: %s", pluginID)
	}

	g.logger.Info("📦 Loading plugin",
		zap.String("plugin_id", pluginID),
		zap.String("config_path", config.Path),
	)

	// Charger le plugin via le loader
	plugin, err := g.loader.LoadPlugin(ctx, pluginID, config)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginID, err)
	}

	// Initialiser le plugin
	if err := plugin.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", pluginID, err)
	}

	// Démarrer le plugin
	if err := plugin.Start(); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", pluginID, err)
	}

	// Créer l'instance
	instance := &PluginInstance{
		Info:     plugin.GetInfo(),
		Plugin:   plugin,
		Status:   PluginStatusActive,
		Metrics:  &PluginMetrics{},
		LastSeen: time.Now(),
		HealthCheck: &HealthCheck{
			Endpoint:  "/health",
			Interval:  30 * time.Second,
			Timeout:   5 * time.Second,
			LastCheck: time.Now(),
		},
	}

	// Enregistrer le plugin
	g.plugins[pluginID] = instance

	// Enregistrer dans le registry
	if err := g.registry.RegisterPlugin(pluginID, instance.Info); err != nil {
		g.logger.Error("Failed to register plugin", zap.String("plugin_id", pluginID), zap.Error(err))
	}

	g.logger.Info("✅ Plugin loaded successfully",
		zap.String("plugin_id", pluginID),
		zap.String("version", instance.Info.Version),
		zap.Int("endpoints_count", len(instance.Info.Endpoints)),
	)

	return nil
}

// UnloadPlugin décharge un plugin
func (g *PluginGateway) UnloadPlugin(ctx context.Context, pluginID string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	instance, exists := g.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	g.logger.Info("🗑️ Unloading plugin", zap.String("plugin_id", pluginID))

	// Arrêter le plugin
	if err := instance.Plugin.Stop(); err != nil {
		g.logger.Error("Failed to stop plugin", zap.String("plugin_id", pluginID), zap.Error(err))
	}

	// Supprimer du registry
	if err := g.registry.UnregisterPlugin(pluginID); err != nil {
		g.logger.Error("Failed to unregister plugin", zap.String("plugin_id", pluginID), zap.Error(err))
	}

	// Supprimer de la liste
	delete(g.plugins, pluginID)

	g.logger.Info("✅ Plugin unloaded", zap.String("plugin_id", pluginID))
	return nil
}

// HandleRequest traite une requête via le plugin approprié
func (g *PluginGateway) HandleRequest(ctx context.Context, req *PluginRequest) (*PluginResponse, error) {
	// Identifier le plugin basé sur le path
	pluginID, err := g.identifyPlugin(req.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to identify plugin: %w", err)
	}

	g.mutex.RLock()
	instance, exists := g.plugins[pluginID]
	g.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Vérifier le health check
	if !g.isPluginHealthy(instance) {
		return nil, fmt.Errorf("plugin unhealthy: %s", pluginID)
	}

	// Appliquer le rate limiting
	if err := g.serviceMesh.RateLimiter.Allow(pluginID); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Appliquer le circuit breaker
	if err := g.serviceMesh.CircuitBreaker.Allow(pluginID); err != nil {
		return nil, fmt.Errorf("circuit breaker open: %w", err)
	}

	// Traiter la requête avec timeout
	ctx, cancel := context.WithTimeout(ctx, g.config.PluginTimeout)
	defer cancel()

	response, err := instance.Plugin.HandleRequest(req)
	if err != nil {
		// Marquer l'erreur dans le circuit breaker
		g.serviceMesh.CircuitBreaker.RecordError(pluginID)
		return nil, fmt.Errorf("plugin request failed: %w", err)
	}

	// Mettre à jour les métriques
	g.updatePluginMetrics(pluginID, req, response)

	return response, nil
}

// GetPluginInfo récupère les informations d'un plugin
func (g *PluginGateway) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	instance, exists := g.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return instance.Info, nil
}

// ListPlugins liste tous les plugins chargés
func (g *PluginGateway) ListPlugins() []*PluginInfo {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	plugins := make([]*PluginInfo, 0, len(g.plugins))
	for _, instance := range g.plugins {
		plugins = append(plugins, instance.Info)
	}

	return plugins
}

// GetPluginMetrics récupère les métriques d'un plugin
func (g *PluginGateway) GetPluginMetrics(pluginID string) (*PluginMetrics, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	instance, exists := g.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return instance.Metrics, nil
}

// identifyPlugin identifie le plugin basé sur le path
func (g *PluginGateway) identifyPlugin(path string) (string, error) {
	// Logique de routing basée sur le path
	// Exemple: /api/v1/recommendations -> ai-recommendations
	// Exemple: /api/v1/analytics -> advanced-analytics

	switch {
	case len(path) > 0 && path[:4] == "/api":
		// Extraire le service du path
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			service := parts[3]
			return service, nil
		}
	}

	return "", fmt.Errorf("unable to identify plugin for path: %s", path)
}

// isPluginHealthy vérifie si un plugin est en bonne santé
func (g *PluginGateway) isPluginHealthy(instance *PluginInstance) bool {
	// Vérifier le dernier health check
	if time.Since(instance.HealthCheck.LastCheck) > instance.HealthCheck.Interval {
		// Effectuer un nouveau health check
		status := instance.Plugin.HealthCheck()
		instance.HealthCheck.LastCheck = time.Now()
		instance.HealthCheck.Status = status

		if status.Status != "healthy" {
			instance.Status = PluginStatusFailed
			return false
		}
	}

	return instance.Status == PluginStatusActive
}

// updatePluginMetrics met à jour les métriques d'un plugin
func (g *PluginGateway) updatePluginMetrics(pluginID string, req *PluginRequest, resp *PluginResponse) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	instance, exists := g.plugins[pluginID]
	if !exists {
		return
	}

	// Mettre à jour les métriques
	instance.Metrics.RequestCount++
	instance.Metrics.LastRequestTime = time.Now()

	if resp.StatusCode >= 400 {
		instance.Metrics.ErrorCount++
	}

	// Calculer le temps de réponse (si disponible)
	if req.Timestamp != nil {
		responseTime := time.Since(*req.Timestamp).Milliseconds()
		instance.Metrics.ResponseTime = float64(responseTime)
	}
}

// loadExistingPlugins charge les plugins existants depuis le registry
func (g *PluginGateway) loadExistingPlugins(ctx context.Context) error {
	plugins, err := g.registry.ListPlugins()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	for _, pluginInfo := range plugins {
		if pluginInfo.Status == PluginStatusActive {
			// Charger le plugin
			config := PluginConfig{
				ID:   pluginInfo.ID,
				Path: pluginInfo.Metadata["path"],
			}

			if err := g.LoadPlugin(ctx, pluginInfo.ID, config); err != nil {
				g.logger.Error("Failed to load existing plugin",
					zap.String("plugin_id", pluginInfo.ID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// monitoringWorker worker pour le monitoring des plugins
func (g *PluginGateway) monitoringWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.performHealthChecks()
			g.cleanupInactivePlugins()
		}
	}
}

// performHealthChecks effectue les health checks de tous les plugins
func (g *PluginGateway) performHealthChecks() {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	for pluginID, instance := range g.plugins {
		if time.Since(instance.HealthCheck.LastCheck) >= instance.HealthCheck.Interval {
			status := instance.Plugin.HealthCheck()
			instance.HealthCheck.LastCheck = time.Now()
			instance.HealthCheck.Status = status

			if status.Status != "healthy" {
				g.logger.Warn("Plugin health check failed",
					zap.String("plugin_id", pluginID),
					zap.String("status", status.Status),
					zap.String("message", status.Message),
				)
			}
		}
	}
}

// cleanupInactivePlugins nettoie les plugins inactifs
func (g *PluginGateway) cleanupInactivePlugins() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now()
	for pluginID, instance := range g.plugins {
		// Supprimer les plugins inactifs depuis plus de 24h
		if instance.Status == PluginStatusFailed && now.Sub(instance.LastSeen) > 24*time.Hour {
			g.logger.Info("Cleaning up inactive plugin", zap.String("plugin_id", pluginID))
			delete(g.plugins, pluginID)
		}
	}
}
