package plugin

import (
	"context"
	"time"
)

// Plugin interface principale pour tous les plugins
type Plugin interface {
	// Lifecycle
	Initialize(config PluginConfig) error
	Start() error
	Stop() error
	HealthCheck() HealthStatus

	// Communication
	HandleRequest(req *PluginRequest) (*PluginResponse, error)
	HandleEvent(event *DomainEvent) error

	// Metadata
	GetInfo() *PluginInfo
	GetMetrics() *PluginMetrics
}

// PluginConfig configuration d'un plugin
type PluginConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Path         string                 `json:"path"`
	Runtime      string                 `json:"runtime"` // "wasm", "docker", "native"
	Resources    ResourceLimits         `json:"resources"`
	Endpoints    []PluginEndpoint       `json:"endpoints"`
	Dependencies []PluginDependency     `json:"dependencies"`
	FeatureFlags []FeatureFlag          `json:"feature_flags"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ResourceLimits limites de ressources pour un plugin
type ResourceLimits struct {
	CPU     float64 `json:"cpu_cores"`
	Memory  int64   `json:"memory_mb"`
	Disk    int64   `json:"disk_mb"`
	Network int64   `json:"network_mbps"`
}

// PluginEndpoint endpoint exposé par un plugin
type PluginEndpoint struct {
	Path      string            `json:"path"`
	Method    string            `json:"method"`
	Auth      bool              `json:"auth"`
	RateLimit int               `json:"rate_limit"`
	Headers   map[string]string `json:"headers"`
}

// PluginDependency dépendance d'un plugin
type PluginDependency struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

// FeatureFlag feature flag pour un plugin
type FeatureFlag struct {
	Name        string `json:"name"`
	Default     bool   `json:"default"`
	Description string `json:"description"`
}

// PluginInfo informations sur un plugin
type PluginInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Status      PluginStatus      `json:"status"`
	Endpoints   []PluginEndpoint  `json:"endpoints"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PluginStatus statut d'un plugin
type PluginStatus string

const (
	PluginStatusActive   PluginStatus = "active"
	PluginStatusInactive PluginStatus = "inactive"
	PluginStatusFailed   PluginStatus = "failed"
	PluginStatusLoading  PluginStatus = "loading"
	PluginStatusStopped  PluginStatus = "stopped"
)

// PluginRequest requête vers un plugin
type PluginRequest struct {
	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Headers   map[string]string      `json:"headers"`
	Body      interface{}            `json:"body"`
	UserID    int64                  `json:"user_id"`
	Context   map[string]interface{} `json:"context"`
	Timestamp *time.Time             `json:"timestamp,omitempty"`
}

// PluginResponse réponse d'un plugin
type PluginResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
}

// DomainEvent événement de domaine
type DomainEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
}

// HealthStatus statut de santé d'un plugin
type HealthStatus struct {
	Status    string            `json:"status"` // "healthy", "unhealthy", "degraded"
	Message   string            `json:"message"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details"`
}

// HealthCheck configuration du health check
type HealthCheck struct {
	Endpoint  string        `json:"endpoint"`
	Interval  time.Duration `json:"interval"`
	Timeout   time.Duration `json:"timeout"`
	Status    HealthStatus  `json:"status"`
	LastCheck time.Time     `json:"last_check"`
}

// PluginMetrics métriques d'un plugin
type PluginMetrics struct {
	RequestCount      int64         `json:"request_count"`
	ErrorCount        int64         `json:"error_count"`
	ResponseTime      float64       `json:"response_time_ms"`
	CPUUsage          float64       `json:"cpu_usage_percent"`
	MemoryUsage       int64         `json:"memory_usage_mb"`
	ActiveConnections int64         `json:"active_connections"`
	LastRequestTime   time.Time     `json:"last_request_time"`
	Uptime            time.Duration `json:"uptime"`
}

// DiscoveryRegistry interface pour le registry de découverte
type DiscoveryRegistry interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RegisterPlugin(pluginID string, info *PluginInfo) error
	UnregisterPlugin(pluginID string) error
	ListPlugins() ([]*PluginInfo, error)
	GetPlugin(pluginID string) (*PluginInfo, error)
	WatchPlugins() (<-chan PluginEvent, error)
}

// PluginEvent événement de plugin
type PluginEvent struct {
	Type      string      `json:"type"` // "registered", "unregistered", "updated"
	PluginID  string      `json:"plugin_id"`
	Info      *PluginInfo `json:"info"`
	Timestamp time.Time   `json:"timestamp"`
}

// PluginLoader interface pour le chargeur de plugins
type PluginLoader interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	LoadPlugin(ctx context.Context, pluginID string, config PluginConfig) (Plugin, error)
	UnloadPlugin(pluginID string) error
	ListLoadedPlugins() []string
}

// ServiceMesh interface pour le service mesh
type ServiceMesh interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetCircuitBreaker() CircuitBreaker
	GetRateLimiter() RateLimiter
	GetLoadBalancer() LoadBalancer
	GetObservability() Observability
}

// CircuitBreaker interface pour le circuit breaker
type CircuitBreaker interface {
	Allow(pluginID string) error
	RecordSuccess(pluginID string)
	RecordError(pluginID string)
	GetStatus(pluginID string) CircuitBreakerStatus
}

// CircuitBreakerStatus statut du circuit breaker
type CircuitBreakerStatus struct {
	State           string     `json:"state"` // "closed", "open", "half-open"
	FailureCount    int        `json:"failure_count"`
	SuccessCount    int        `json:"success_count"`
	LastFailureTime *time.Time `json:"last_failure_time,omitempty"`
}

// RateLimiter interface pour le rate limiter
type RateLimiter interface {
	Allow(pluginID string) error
	GetLimit(pluginID string) RateLimit
	SetLimit(pluginID string, limit RateLimit) error
}

// RateLimit configuration du rate limit
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// LoadBalancer interface pour le load balancer
type LoadBalancer interface {
	GetInstance(pluginID string) (string, error)
	RegisterInstance(pluginID string, instance string) error
	UnregisterInstance(pluginID string, instance string) error
	GetInstances(pluginID string) ([]string, error)
}

// Observability interface pour l'observabilité
type Observability interface {
	RecordMetric(pluginID string, metric string, value float64, labels map[string]string)
	RecordLog(pluginID string, level string, message string, fields map[string]interface{})
	RecordTrace(pluginID string, traceID string, spanID string, operation string, duration time.Duration)
}

// SandboxManager interface pour la gestion du sandbox
type SandboxManager interface {
	CreateSandbox(pluginID string, config SandboxConfig) (Sandbox, error)
	DestroySandbox(pluginID string) error
	ListSandboxes() []string
}

// SandboxConfig configuration du sandbox
type SandboxConfig struct {
	Resources ResourceLimits `json:"resources"`
	Network   NetworkConfig  `json:"network"`
	Security  SecurityConfig `json:"security"`
}

// NetworkConfig configuration réseau
type NetworkConfig struct {
	Isolation  bool     `json:"isolation"`
	AllowedIPs []string `json:"allowed_ips"`
	Ports      []int    `json:"ports"`
}

// SecurityConfig configuration de sécurité
type SecurityConfig struct {
	ReadOnlyFS   bool     `json:"read_only_fs"`
	NoNetwork    bool     `json:"no_network"`
	Capabilities []string `json:"capabilities"`
}

// Sandbox interface pour le sandbox
type Sandbox interface {
	ID() string
	Status() SandboxStatus
	Execute(command string, args []string) ([]byte, error)
	Kill() error
}

// SandboxStatus statut du sandbox
type SandboxStatus struct {
	State       string    `json:"state"` // "running", "stopped", "error"
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage int64     `json:"memory_usage"`
	StartTime   time.Time `json:"start_time"`
}
