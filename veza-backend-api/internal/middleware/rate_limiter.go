package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RateLimitConfig configuration pour le rate limiting
type RateLimitConfig struct {
	// Limites par endpoint
	EndpointLimits map[string]EndpointLimit

	// Limites globales par IP
	GlobalIPLimit  int           // Requêtes par minute
	GlobalIPWindow time.Duration // Fenêtre de temps

	// Limites par utilisateur authentifié
	UserLimit  int           // Requêtes par minute
	UserWindow time.Duration // Fenêtre de temps

	// Protection DDoS
	DDoSThreshold   int           // Seuil pour déclencher la protection
	DDoSBanDuration time.Duration // Durée du ban

	// Whitelist/Blacklist
	WhitelistIPs []string // IPs exemptées
	BlacklistIPs []string // IPs bannies

	// Redis
	RedisClient *redis.Client
	KeyPrefix   string

	// Logging
	Logger *zap.Logger
}

// EndpointLimit limite spécifique pour un endpoint
type EndpointLimit struct {
	Path           string        // Pattern de l'endpoint
	Method         string        // Méthode HTTP
	Limit          int           // Nombre de requêtes
	Window         time.Duration // Fenêtre de temps
	AuthRequired   bool          // Si l'endpoint nécessite l'auth
	BypassForAdmin bool          // Bypass pour les admins
}

// RateLimitResult résultat de la vérification
type RateLimitResult struct {
	Allowed    bool
	Remaining  int
	ResetTime  time.Time
	RetryAfter time.Duration
	Reason     string
}

// DistributedRateLimiter rate limiter distribué avec Redis
type DistributedRateLimiter struct {
	config *RateLimitConfig
	redis  *redis.Client
	logger *zap.Logger
}

// NewDistributedRateLimiter crée un nouveau rate limiter distribué
func NewDistributedRateLimiter(config *RateLimitConfig) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		config: config,
		redis:  config.RedisClient,
		logger: config.Logger,
	}
}

// Middleware retourne le middleware Gin pour le rate limiting
func (rl *DistributedRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraire les informations de la requête
		clientIP := rl.getClientIP(c)
		userID := rl.getUserID(c)
		endpoint := c.FullPath()
		method := c.Request.Method

		// Vérifier la blacklist
		if rl.isBlacklisted(clientIP) {
			rl.rejectRequest(c, "IP blacklisted", http.StatusForbidden, 0)
			return
		}

		// Vérifier la whitelist
		if rl.isWhitelisted(clientIP) {
			c.Next()
			return
		}

		// Vérifier la protection DDoS
		if banned, banDuration := rl.checkDDoSProtection(clientIP); banned {
			rl.rejectRequest(c, "DDoS protection triggered", http.StatusTooManyRequests, banDuration)
			return
		}

		// Vérifier les limites spécifiques à l'endpoint
		if limit, exists := rl.getEndpointLimit(endpoint, method); exists {
			if result := rl.checkEndpointLimit(clientIP, userID, endpoint, method, limit); !result.Allowed {
				rl.rejectRequestWithResult(c, result)
				return
			}
		}

		// Vérifier les limites globales
		if result := rl.checkGlobalLimits(clientIP, userID); !result.Allowed {
			rl.rejectRequestWithResult(c, result)
			return
		}

		// Enregistrer la requête pour les statistiques
		rl.recordRequest(clientIP, userID, endpoint, method)

		c.Next()
	}
}

// getClientIP extrait l'IP du client
func (rl *DistributedRateLimiter) getClientIP(c *gin.Context) string {
	// Vérifier les headers pour les proxies
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	return c.ClientIP()
}

// getUserID extrait l'ID utilisateur si authentifié
func (rl *DistributedRateLimiter) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			return strconv.Itoa(id)
		}
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// isWhitelisted vérifie si l'IP est dans la whitelist
func (rl *DistributedRateLimiter) isWhitelisted(ip string) bool {
	for _, whiteIP := range rl.config.WhitelistIPs {
		if ip == whiteIP {
			return true
		}
	}
	return false
}

// isBlacklisted vérifie si l'IP est dans la blacklist
func (rl *DistributedRateLimiter) isBlacklisted(ip string) bool {
	ctx := context.Background()

	// Vérifier la blacklist statique
	for _, blackIP := range rl.config.BlacklistIPs {
		if ip == blackIP {
			return true
		}
	}

	// Vérifier la blacklist dynamique Redis
	blacklistKey := fmt.Sprintf("%sblacklist:%s", rl.config.KeyPrefix, ip)
	exists, err := rl.redis.Exists(ctx, blacklistKey).Result()
	if err != nil {
		rl.logger.Warn("Erreur vérification blacklist", zap.Error(err))
		return false
	}

	return exists > 0
}

// checkDDoSProtection vérifie la protection DDoS
func (rl *DistributedRateLimiter) checkDDoSProtection(ip string) (banned bool, duration time.Duration) {
	ctx := context.Background()

	// Vérifier si l'IP est déjà bannie
	banKey := fmt.Sprintf("%sddos_ban:%s", rl.config.KeyPrefix, ip)
	ttl, err := rl.redis.TTL(ctx, banKey).Result()
	if err == nil && ttl > 0 {
		return true, ttl
	}

	// Vérifier le nombre de requêtes dans la dernière minute
	counterKey := fmt.Sprintf("%sddos_counter:%s", rl.config.KeyPrefix, ip)
	count, err := rl.redis.Incr(ctx, counterKey).Result()
	if err != nil {
		rl.logger.Warn("Erreur compteur DDoS", zap.Error(err))
		return false, 0
	}

	// Expirer le compteur après 1 minute
	if count == 1 {
		rl.redis.Expire(ctx, counterKey, time.Minute)
	}

	// Si le seuil est dépassé, bannir l'IP
	if int(count) > rl.config.DDoSThreshold {
		rl.redis.Set(ctx, banKey, "banned", rl.config.DDoSBanDuration)

		rl.logger.Warn("DDoS protection triggered",
			zap.String("ip", ip),
			zap.Int64("requests", count),
			zap.Duration("ban_duration", rl.config.DDoSBanDuration))

		return true, rl.config.DDoSBanDuration
	}

	return false, 0
}

// getEndpointLimit récupère la limite pour un endpoint spécifique
func (rl *DistributedRateLimiter) getEndpointLimit(path, method string) (EndpointLimit, bool) {
	key := fmt.Sprintf("%s:%s", method, path)
	if limit, exists := rl.config.EndpointLimits[key]; exists {
		return limit, true
	}

	// Vérifier les patterns génériques
	for pattern, limit := range rl.config.EndpointLimits {
		if strings.Contains(pattern, "*") {
			// Logique de matching basique pour les wildcards
			if rl.matchPattern(pattern, key) {
				return limit, true
			}
		}
	}

	return EndpointLimit{}, false
}

// checkEndpointLimit vérifie la limite pour un endpoint spécifique
func (rl *DistributedRateLimiter) checkEndpointLimit(ip, userID, endpoint, method string, limit EndpointLimit) RateLimitResult {
	ctx := context.Background()

	// Utiliser l'ID utilisateur si disponible et si l'endpoint le permet
	identifier := ip
	if userID != "" && limit.AuthRequired {
		identifier = fmt.Sprintf("user:%s", userID)
	}

	key := fmt.Sprintf("%sendpoint:%s:%s:%s", rl.config.KeyPrefix, method, endpoint, identifier)

	return rl.checkLimit(ctx, key, limit.Limit, limit.Window)
}

// checkGlobalLimits vérifie les limites globales
func (rl *DistributedRateLimiter) checkGlobalLimits(ip, userID string) RateLimitResult {
	ctx := context.Background()

	// Vérifier la limite par IP
	ipKey := fmt.Sprintf("%sglobal_ip:%s", rl.config.KeyPrefix, ip)
	if result := rl.checkLimit(ctx, ipKey, rl.config.GlobalIPLimit, rl.config.GlobalIPWindow); !result.Allowed {
		result.Reason = "IP rate limit exceeded"
		return result
	}

	// Vérifier la limite par utilisateur si authentifié
	if userID != "" {
		userKey := fmt.Sprintf("%sglobal_user:%s", rl.config.KeyPrefix, userID)
		if result := rl.checkLimit(ctx, userKey, rl.config.UserLimit, rl.config.UserWindow); !result.Allowed {
			result.Reason = "User rate limit exceeded"
			return result
		}
	}

	return RateLimitResult{Allowed: true}
}

// checkLimit vérifie une limite avec l'algorithme sliding window
func (rl *DistributedRateLimiter) checkLimit(ctx context.Context, key string, limit int, window time.Duration) RateLimitResult {
	now := time.Now()
	windowStart := now.Add(-window)

	// Script Lua pour l'opération atomique
	luaScript := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_seconds = tonumber(ARGV[4])
		
		-- Nettoyer les entrées expirées
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Compter les requêtes actuelles
		local current = redis.call('ZCARD', key)
		
		if current >= limit then
			-- Récupérer le timestamp de la plus ancienne requête
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset_time = now + window_seconds
			if #oldest > 0 then
				reset_time = oldest[2] + window_seconds
			end
			
			return {0, current, reset_time}
		else
			-- Ajouter la requête actuelle
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, window_seconds)
			
			local remaining = limit - current - 1
			local reset_time = now + window_seconds
			
			return {1, remaining, reset_time}
		end
	`

	result, err := rl.redis.Eval(ctx, luaScript, []string{key},
		now.Unix(),
		windowStart.Unix(),
		limit,
		int(window.Seconds())).Result()

	if err != nil {
		rl.logger.Error("Erreur script Lua rate limit", zap.Error(err))
		return RateLimitResult{Allowed: true} // Fail open
	}

	// Parser le résultat
	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := int(resultSlice[1].(int64))
	resetTime := time.Unix(resultSlice[2].(int64), 0)

	retryAfter := time.Duration(0)
	if !allowed {
		retryAfter = time.Until(resetTime)
	}

	return RateLimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}

// recordRequest enregistre une requête pour les statistiques
func (rl *DistributedRateLimiter) recordRequest(ip, userID, endpoint, method string) {
	ctx := context.Background()
	timestamp := time.Now().Unix()

	// Enregistrer pour les statistiques
	statsKey := fmt.Sprintf("%sstats:%s", rl.config.KeyPrefix, time.Now().Format("2006-01-02-15"))
	rl.redis.ZAdd(ctx, statsKey, &redis.Z{
		Score:  float64(timestamp),
		Member: fmt.Sprintf("%s|%s|%s|%s", ip, userID, method, endpoint),
	})

	// Expirer après 24h
	rl.redis.Expire(ctx, statsKey, 24*time.Hour)
}

// matchPattern fonction simple de matching pour les wildcards
func (rl *DistributedRateLimiter) matchPattern(pattern, path string) bool {
	// Implémentation basique - en production, utiliser une lib plus robuste
	if strings.Contains(pattern, "*") {
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

// rejectRequest rejette une requête avec un message d'erreur
func (rl *DistributedRateLimiter) rejectRequest(c *gin.Context, reason string, statusCode int, retryAfter time.Duration) {
	result := RateLimitResult{
		Allowed:    false,
		Reason:     reason,
		RetryAfter: retryAfter,
	}
	rl.rejectRequestWithResult(c, result)
}

// rejectRequestWithResult rejette une requête avec un résultat détaillé
func (rl *DistributedRateLimiter) rejectRequestWithResult(c *gin.Context, result RateLimitResult) {
	// Headers de rate limiting
	c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.GlobalIPLimit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

	if result.RetryAfter > 0 {
		c.Header("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
	}

	// Log de l'événement
	rl.logger.Warn("Rate limit exceeded",
		zap.String("ip", rl.getClientIP(c)),
		zap.String("user_id", rl.getUserID(c)),
		zap.String("endpoint", c.FullPath()),
		zap.String("method", c.Request.Method),
		zap.String("reason", result.Reason),
		zap.Duration("retry_after", result.RetryAfter))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":               "Rate limit exceeded",
		"message":             result.Reason,
		"retry_after_seconds": int(result.RetryAfter.Seconds()),
		"reset_time":          result.ResetTime.Unix(),
	})
	c.Abort()
}

// GetDefaultConfig retourne une configuration par défaut SÉCURISÉE
func GetDefaultRateLimitConfig(redisClient *redis.Client, logger *zap.Logger) *RateLimitConfig {
	return &RateLimitConfig{
		EndpointLimits: map[string]EndpointLimit{
			// Authentification - PROTECTION RENFORCÉE
			"POST:/api/v1/auth/login": {
				Path:           "/api/v1/auth/login",
				Method:         "POST",
				Limit:          3, // Réduit de 5 à 3 tentatives
				Window:         15 * time.Minute,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			"POST:/api/auth/login": {
				Path:           "/api/auth/login",
				Method:         "POST",
				Limit:          3, // Réduit de 5 à 3 tentatives
				Window:         15 * time.Minute,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			"POST:/api/v1/auth/register": {
				Path:           "/api/v1/auth/register",
				Method:         "POST",
				Limit:          2, // Réduit de 3 à 2 tentatives
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			"POST:/api/auth/register": {
				Path:           "/api/auth/register",
				Method:         "POST",
				Limit:          2, // Réduit de 3 à 2 tentatives
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			"POST:/api/v1/auth/refresh": {
				Path:           "/api/v1/auth/refresh",
				Method:         "POST",
				Limit:          8, // Réduit de 10 à 8 tentatives
				Window:         time.Hour,
				AuthRequired:   true,
				BypassForAdmin: true,
			},
			"POST:/api/auth/refresh": {
				Path:           "/api/auth/refresh",
				Method:         "POST",
				Limit:          8, // Réduit de 10 à 8 tentatives
				Window:         time.Hour,
				AuthRequired:   true,
				BypassForAdmin: true,
			},
			// Endpoints sensibles ajoutés
			"POST:/api/v1/auth/forgot-password": {
				Path:           "/api/v1/auth/forgot-password",
				Method:         "POST",
				Limit:          3,
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			"POST:/api/v1/auth/reset-password": {
				Path:           "/api/v1/auth/reset-password",
				Method:         "POST",
				Limit:          5,
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: false,
			},
			// Endpoints d'écriture - limites strictes
			"POST:/api/v1/*": {
				Path:           "/api/v1/*",
				Method:         "POST",
				Limit:          30, // 30 POST par heure
				Window:         time.Hour,
				AuthRequired:   true,
				BypassForAdmin: true,
			},
			"PUT:/api/v1/*": {
				Path:           "/api/v1/*",
				Method:         "PUT",
				Limit:          20, // 20 PUT par heure
				Window:         time.Hour,
				AuthRequired:   true,
				BypassForAdmin: true,
			},
			"DELETE:/api/v1/*": {
				Path:           "/api/v1/*",
				Method:         "DELETE",
				Limit:          10, // 10 DELETE par heure
				Window:         time.Hour,
				AuthRequired:   true,
				BypassForAdmin: true,
			},
			// Endpoints de lecture - limites généreuses mais contrôlées
			"GET:/api/v1/*": {
				Path:           "/api/v1/*",
				Method:         "GET",
				Limit:          500, // Réduit de 1000 à 500
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: true,
			},
			"GET:/api/*": {
				Path:           "/api/*",
				Method:         "GET",
				Limit:          500, // Réduit de 1000 à 500
				Window:         time.Hour,
				AuthRequired:   false,
				BypassForAdmin: true,
			},
		},

		// Limites globales RENFORCÉES
		GlobalIPLimit:  30, // Réduit de 60 à 30 req/min par IP
		GlobalIPWindow: time.Minute,

		UserLimit:  80, // Réduit de 100 à 80 req/min par utilisateur
		UserWindow: time.Minute,

		// Protection DDoS RENFORCÉE
		DDoSThreshold:   100,              // Réduit de 200 à 100 req/min déclenche DDoS
		DDoSBanDuration: 30 * time.Minute, // Augmenté de 10 à 30 minutes

		WhitelistIPs: []string{"127.0.0.1", "::1", "localhost"},
		BlacklistIPs: []string{},

		RedisClient: redisClient,
		KeyPrefix:   "veza:ratelimit:",
		Logger:      logger,
	}
}
