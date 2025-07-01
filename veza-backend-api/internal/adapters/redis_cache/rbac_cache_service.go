package redis_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/constants"
)

// RBACCacheService cache spécialisé pour les permissions RBAC
type RBACCacheService struct {
	redis      *redis.Client
	localCache sync.Map // Cache local pour permissions fréquentes
	logger     *zap.Logger
	metrics    *RBACCacheMetrics

	// Configuration
	permissionTTL time.Duration
	roleTTL       time.Duration
	userRoleTTL   time.Duration
}

// RBACCacheMetrics métriques spécifiques aux permissions
type RBACCacheMetrics struct {
	PermissionChecks  int64 `json:"permission_checks"`
	PermissionHits    int64 `json:"permission_hits"`
	RoleChecks        int64 `json:"role_checks"`
	RoleHits          int64 `json:"role_hits"`
	UserRoleChecks    int64 `json:"user_role_checks"`
	UserRoleHits      int64 `json:"user_role_hits"`
	AvgCheckLatencyMs int64 `json:"avg_check_latency_ms"`

	mutex sync.RWMutex
}

// PermissionResult résultat d'une vérification de permission
type PermissionResult struct {
	Allowed    bool      `json:"allowed"`
	UserID     int64     `json:"user_id"`
	Resource   string    `json:"resource"`
	Action     string    `json:"action"`
	Role       string    `json:"role"`
	CheckedAt  time.Time `json:"checked_at"`
	FromCache  bool      `json:"from_cache"`
	CacheLevel int       `json:"cache_level"`
}

// RolePermissions structure pour cache des permissions de rôle
type RolePermissions struct {
	Role        string              `json:"role"`
	Permissions map[string][]string `json:"permissions"` // resource -> actions[]
	CachedAt    time.Time           `json:"cached_at"`
	ExpiresAt   time.Time           `json:"expires_at"`
}

// UserRoleCache cache des rôles utilisateur
type UserRoleCache struct {
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewRBACCacheService crée un nouveau service de cache RBAC
func NewRBACCacheService(
	redisClient *redis.Client,
	logger *zap.Logger,
) *RBACCacheService {
	service := &RBACCacheService{
		redis:         redisClient,
		logger:        logger,
		metrics:       &RBACCacheMetrics{},
		permissionTTL: 15 * time.Minute, // Permissions changent rarement
		roleTTL:       1 * time.Hour,    // Rôles très stables
		userRoleTTL:   30 * time.Minute, // Rôles utilisateur moyennement stables
	}

	// Pré-charger les permissions de base
	go service.preloadBasePermissions()

	// Démarrer le monitoring
	go service.startMetricsReporter()

	return service
}

// ============================================================================
// VÉRIFICATION DE PERMISSIONS ULTRA-OPTIMISÉE
// ============================================================================

// CheckPermissionFast vérification ultra-rapide des permissions
func (r *RBACCacheService) CheckPermissionFast(
	ctx context.Context,
	userID int64,
	resource, action string,
) (*PermissionResult, error) {
	start := time.Now()
	defer func() {
		r.recordCheckLatency(time.Since(start))
	}()

	r.metrics.mutex.Lock()
	r.metrics.PermissionChecks++
	r.metrics.mutex.Unlock()

	// Clé de cache composite pour performance maximale
	permKey := fmt.Sprintf("perm:%d:%s:%s", userID, resource, action)

	// Niveau 1 : Cache local (sub-millisecond)
	if cached, found := r.localCache.Load(permKey); found {
		if result, ok := cached.(*PermissionResult); ok {
			if time.Now().Before(result.CheckedAt.Add(5 * time.Minute)) {
				r.recordPermissionHit()
				result.FromCache = true
				result.CacheLevel = 1

				r.logger.Debug("Permission trouvée en cache L1",
					zap.Int64("user_id", userID),
					zap.String("resource", resource),
					zap.String("action", action),
					zap.Bool("allowed", result.Allowed))

				return result, nil
			}
		}
	}

	// Niveau 2 : Redis distribué (1-2ms)
	if resultData, err := r.redis.Get(ctx, permKey).Result(); err == nil {
		var result PermissionResult
		if err := json.Unmarshal([]byte(resultData), &result); err == nil {
			r.recordPermissionHit()
			result.FromCache = true
			result.CacheLevel = 2

			// Remonter en cache local
			r.localCache.Store(permKey, &result)

			r.logger.Debug("Permission trouvée en cache L2",
				zap.Int64("user_id", userID),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.Bool("allowed", result.Allowed))

			return &result, nil
		}
	}

	// Niveau 3 : Calcul et mise en cache
	result, err := r.computeAndCachePermission(ctx, userID, resource, action)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// computeAndCachePermission calcule et met en cache une permission
func (r *RBACCacheService) computeAndCachePermission(
	ctx context.Context,
	userID int64,
	resource, action string,
) (*PermissionResult, error) {
	// Récupérer le rôle utilisateur (avec cache)
	userRole, err := r.getUserRoleCached(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération rôle: %w", err)
	}

	// Vérifier les permissions du rôle (avec cache)
	allowed, err := r.checkRolePermissionCached(ctx, userRole, resource, action)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification permission rôle: %w", err)
	}

	result := &PermissionResult{
		Allowed:    allowed,
		UserID:     userID,
		Resource:   resource,
		Action:     action,
		Role:       userRole,
		CheckedAt:  time.Now(),
		FromCache:  false,
		CacheLevel: 0,
	}

	// Mise en cache parallèle multi-niveaux
	go r.cachePermissionResult(ctx, result)

	r.logger.Debug("Permission calculée et mise en cache",
		zap.Int64("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.String("role", userRole),
		zap.Bool("allowed", allowed))

	return result, nil
}

// cachePermissionResult met en cache le résultat sur tous les niveaux
func (r *RBACCacheService) cachePermissionResult(ctx context.Context, result *PermissionResult) {
	permKey := fmt.Sprintf("perm:%d:%s:%s", result.UserID, result.Resource, result.Action)

	var wg sync.WaitGroup

	// Cache local
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.localCache.Store(permKey, result)
	}()

	// Cache Redis
	wg.Add(1)
	go func() {
		defer wg.Done()
		if data, err := json.Marshal(result); err == nil {
			r.redis.Set(ctx, permKey, data, r.permissionTTL)
		}
	}()

	wg.Wait()
}

// ============================================================================
// CACHE DES RÔLES UTILISATEUR
// ============================================================================

// getUserRoleCached récupère le rôle d'un utilisateur avec cache
func (r *RBACCacheService) getUserRoleCached(ctx context.Context, userID int64) (string, error) {
	r.metrics.mutex.Lock()
	r.metrics.UserRoleChecks++
	r.metrics.mutex.Unlock()

	userRoleKey := fmt.Sprintf("user_role:%d", userID)

	// Essayer Redis en premier pour les rôles utilisateur
	if roleData, err := r.redis.Get(ctx, userRoleKey).Result(); err == nil {
		var userRole UserRoleCache
		if err := json.Unmarshal([]byte(roleData), &userRole); err == nil {
			if time.Now().Before(userRole.ExpiresAt) {
				r.recordUserRoleHit()
				return userRole.Role, nil
			}
		}
	}

	// TODO: Ici, vous devriez récupérer depuis la base de données
	// Pour l'instant, retourner un rôle par défaut
	defaultRole := string(constants.RoleUser)

	// Mettre en cache
	userRole := UserRoleCache{
		UserID:    userID,
		Role:      defaultRole,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(r.userRoleTTL),
	}

	if data, err := json.Marshal(userRole); err == nil {
		r.redis.Set(ctx, userRoleKey, data, r.userRoleTTL)
	}

	return defaultRole, nil
}

// ============================================================================
// CACHE DES PERMISSIONS DE RÔLE
// ============================================================================

// checkRolePermissionCached vérifie les permissions d'un rôle avec cache
func (r *RBACCacheService) checkRolePermissionCached(
	ctx context.Context,
	role, resource, action string,
) (bool, error) {
	r.metrics.mutex.Lock()
	r.metrics.RoleChecks++
	r.metrics.mutex.Unlock()

	rolePermKey := fmt.Sprintf("role_perm:%s", role)

	// Vérifier le cache Redis
	if permData, err := r.redis.Get(ctx, rolePermKey).Result(); err == nil {
		var rolePerms RolePermissions
		if err := json.Unmarshal([]byte(permData), &rolePerms); err == nil {
			if time.Now().Before(rolePerms.ExpiresAt) {
				r.recordRoleHit()
				return r.hasPermission(&rolePerms, resource, action), nil
			}
		}
	}

	// Charger et mettre en cache les permissions du rôle
	rolePerms, err := r.loadRolePermissions(ctx, role)
	if err != nil {
		return false, err
	}

	return r.hasPermission(rolePerms, resource, action), nil
}

// loadRolePermissions charge les permissions d'un rôle
func (r *RBACCacheService) loadRolePermissions(ctx context.Context, role string) (*RolePermissions, error) {
	// Définir les permissions par défaut selon les rôles
	permissions := make(map[string][]string)

	switch role {
	case string(constants.RoleAdmin):
		permissions = map[string][]string{
			"*": {"*"}, // Admin a tous les droits
		}
	case string(constants.RoleUser):
		permissions = map[string][]string{
			"profile": {"read", "update"},
			"chat":    {"read", "write"},
			"stream":  {"read"},
			"upload":  {"create"},
		}
	case string(constants.RoleModerator):
		permissions = map[string][]string{
			"profile": {"read", "update"},
			"chat":    {"read", "write", "moderate"},
			"stream":  {"read", "moderate"},
			"upload":  {"create", "moderate"},
			"users":   {"read"},
		}
	default:
		// Rôle inconnu, permissions minimales
		permissions = map[string][]string{
			"profile": {"read"},
		}
	}

	rolePerms := &RolePermissions{
		Role:        role,
		Permissions: permissions,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(r.roleTTL),
	}

	// Mettre en cache
	rolePermKey := fmt.Sprintf("role_perm:%s", role)
	if data, err := json.Marshal(rolePerms); err == nil {
		r.redis.Set(ctx, rolePermKey, data, r.roleTTL)
	}

	return rolePerms, nil
}

// hasPermission vérifie si un rôle a une permission spécifique
func (r *RBACCacheService) hasPermission(rolePerms *RolePermissions, resource, action string) bool {
	// Vérifier les permissions globales (admin)
	if actions, exists := rolePerms.Permissions["*"]; exists {
		for _, a := range actions {
			if a == "*" || a == action {
				return true
			}
		}
	}

	// Vérifier les permissions spécifiques à la resource
	if actions, exists := rolePerms.Permissions[resource]; exists {
		for _, a := range actions {
			if a == "*" || a == action {
				return true
			}
		}
	}

	return false
}

// ============================================================================
// PRÉ-CHARGEMENT DES PERMISSIONS DE BASE
// ============================================================================

// preloadBasePermissions pré-charge les permissions les plus fréquentes
func (r *RBACCacheService) preloadBasePermissions() {
	ctx := context.Background()

	// Permissions de base à pré-charger
	basePermissions := []struct {
		role     string
		resource string
		action   string
	}{
		{string(constants.RoleUser), "profile", "read"},
		{string(constants.RoleUser), "profile", "update"},
		{string(constants.RoleUser), "chat", "read"},
		{string(constants.RoleUser), "chat", "write"},
		{string(constants.RoleModerator), "chat", "moderate"},
		{string(constants.RoleAdmin), "users", "read"},
		{string(constants.RoleAdmin), "users", "write"},
	}

	for _, perm := range basePermissions {
		_, err := r.loadRolePermissions(ctx, perm.role)
		if err != nil {
			r.logger.Warn("Erreur pré-chargement permission",
				zap.String("role", perm.role),
				zap.Error(err))
		}
	}

	r.logger.Info("Permissions de base pré-chargées",
		zap.Int("count", len(basePermissions)))
}

// ============================================================================
// INVALIDATION INTELLIGENTE
// ============================================================================

// InvalidateUserPermissions invalide toutes les permissions d'un utilisateur
func (r *RBACCacheService) InvalidateUserPermissions(ctx context.Context, userID int64) error {
	// Pattern pour toutes les permissions de l'utilisateur
	pattern := fmt.Sprintf("perm:%d:*", userID)

	// Supprimer du cache local
	r.localCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.HasPrefix(keyStr, fmt.Sprintf("perm:%d:", userID)) {
				r.localCache.Delete(key)
			}
		}
		return true
	})

	// Supprimer de Redis
	keys, err := r.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		r.redis.Del(ctx, keys...)
	}

	// Supprimer le cache du rôle utilisateur
	userRoleKey := fmt.Sprintf("user_role:%d", userID)
	r.redis.Del(ctx, userRoleKey)

	r.logger.Debug("Permissions utilisateur invalidées",
		zap.Int64("user_id", userID),
		zap.Int("keys_deleted", len(keys)))

	return nil
}

// InvalidateRolePermissions invalide les permissions d'un rôle
func (r *RBACCacheService) InvalidateRolePermissions(ctx context.Context, role string) error {
	rolePermKey := fmt.Sprintf("role_perm:%s", role)

	// Supprimer de Redis
	r.redis.Del(ctx, rolePermKey)

	// Recharger immédiatement pour les prochaines requêtes
	go r.loadRolePermissions(ctx, role)

	r.logger.Debug("Permissions de rôle invalidées",
		zap.String("role", role))

	return nil
}

// ============================================================================
// MÉTRIQUES ET MONITORING
// ============================================================================

func (r *RBACCacheService) recordPermissionHit() {
	r.metrics.mutex.Lock()
	r.metrics.PermissionHits++
	r.metrics.mutex.Unlock()
}

func (r *RBACCacheService) recordRoleHit() {
	r.metrics.mutex.Lock()
	r.metrics.RoleHits++
	r.metrics.mutex.Unlock()
}

func (r *RBACCacheService) recordUserRoleHit() {
	r.metrics.mutex.Lock()
	r.metrics.UserRoleHits++
	r.metrics.mutex.Unlock()
}

func (r *RBACCacheService) recordCheckLatency(latency time.Duration) {
	r.metrics.mutex.Lock()
	r.metrics.AvgCheckLatencyMs = (r.metrics.AvgCheckLatencyMs + latency.Milliseconds()) / 2
	r.metrics.mutex.Unlock()
}

// GetRBACMetrics retourne les métriques RBAC
func (r *RBACCacheService) GetRBACMetrics() *RBACCacheMetrics {
	r.metrics.mutex.RLock()
	defer r.metrics.mutex.RUnlock()

	return &RBACCacheMetrics{
		PermissionChecks:  r.metrics.PermissionChecks,
		PermissionHits:    r.metrics.PermissionHits,
		RoleChecks:        r.metrics.RoleChecks,
		RoleHits:          r.metrics.RoleHits,
		UserRoleChecks:    r.metrics.UserRoleChecks,
		UserRoleHits:      r.metrics.UserRoleHits,
		AvgCheckLatencyMs: r.metrics.AvgCheckLatencyMs,
	}
}

func (r *RBACCacheService) startMetricsReporter() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		metrics := r.GetRBACMetrics()

		permissionHitRatio := float64(0)
		if metrics.PermissionChecks > 0 {
			permissionHitRatio = float64(metrics.PermissionHits) / float64(metrics.PermissionChecks)
		}

		roleHitRatio := float64(0)
		if metrics.RoleChecks > 0 {
			roleHitRatio = float64(metrics.RoleHits) / float64(metrics.RoleChecks)
		}

		r.logger.Info("Métriques cache RBAC",
			zap.Int64("permission_checks", metrics.PermissionChecks),
			zap.Float64("permission_hit_ratio", permissionHitRatio),
			zap.Int64("role_checks", metrics.RoleChecks),
			zap.Float64("role_hit_ratio", roleHitRatio),
			zap.Int64("avg_latency_ms", metrics.AvgCheckLatencyMs))
	}
}

// HealthCheck vérifie la santé du service RBAC
func (r *RBACCacheService) HealthCheck(ctx context.Context) error {
	// Test de vérification de permission
	testResult, err := r.CheckPermissionFast(ctx, 1, "test", "read")
	if err != nil {
		return fmt.Errorf("échec test permission: %w", err)
	}

	if testResult == nil {
		return fmt.Errorf("résultat permission nul")
	}

	return nil
}
