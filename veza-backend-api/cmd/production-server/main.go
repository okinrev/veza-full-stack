package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/api"
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"
	"github.com/okinrev/veza-web-app/internal/websocket"
)

var startTime = time.Now()

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func main() {
	log.Println("🚀 VEZA BACKEND - PRODUCTION SERVER v1.0")
	log.Println("==========================================")
	log.Println("🏗️  Architecture: Hexagonal + Clean Architecture")
	log.Println("🛡️  Security: JWT + Rate Limiting + CORS")
	log.Println("📊 Performance: Optimized for High Load")
	log.Println("⚡ Features: Full API + WebSocket Support")
	log.Println("")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	// Configuration
	cfg := config.New()

	// Initialize structured logger
	var logger *zap.Logger
	var err error
	if cfg.Server.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// Set production mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize Redis for rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getIntEnv("REDIS_DB", 0),
	})

	// Test Redis connection
	if err := redisClient.Ping(redisClient.Context()).Err(); err != nil {
		logger.Warn("Redis connection failed, rate limiting will use fallback mode", zap.Error(err))
		redisClient = nil
	} else {
		logger.Info("✅ Redis connected successfully")
	}

	// Database connection with retry logic
	var db *database.DB

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		db, err = database.NewConnection(cfg.Database.URL)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		log.Fatal("❌ Database connection failed after retries:", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	} else {
		log.Println("✅ Database migrations completed")
	}

	// Initialize router
	router := gin.New()

	// Production middlewares
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.Use(gin.Recovery())

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	})

	// Advanced Rate Limiting with Redis
	if redisClient != nil {
		rateLimitConfig := middleware.GetDefaultRateLimitConfig(redisClient, logger)
		rateLimiter := middleware.NewDistributedRateLimiter(rateLimitConfig)
		router.Use(rateLimiter.Middleware())
		logger.Info("✅ Advanced Rate Limiting enabled with Redis")
	} else {
		// Fallback to simple rate limiting
		router.Use(middleware.RateLimiterAdvanced())
		logger.Info("⚠️  Using fallback rate limiting (no Redis)")
	}

	// CORS middleware for production
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins in production
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://veza.app",
			"https://app.veza.com",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Api-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint for load balancer
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"service":    "veza-backend-production",
			"version":    "1.0.0",
			"timestamp":  time.Now().Unix(),
			"uptime":     time.Since(startTime).Seconds(),
			"database":   "connected",
			"websocket":  "enabled",
			"rate_limit": redisClient != nil,
			"redis":      redisClient != nil,
		})
	})

	// Readiness check for Kubernetes
	router.GET("/ready", func(c *gin.Context) {
		// Check if database is accessible
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"reason": "database_unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().Unix(),
		})
	})

	// Metrics endpoint for monitoring
	router.GET("/metrics", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.JSON(http.StatusOK, gin.H{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc,
			"memory_sys":   m.Sys,
			"timestamp":    time.Now().Unix(),
		})
	})

	// Initialize WebSocket chat manager
	chatManager := websocket.NewChatManager(cfg.JWT.Secret)
	go chatManager.Run()

	// WebSocket endpoint
	router.GET("/ws/chat", func(c *gin.Context) {
		chatManager.HandleWebSocket(c)
	})

	// API routes
	api.SetupRoutes(router, db, cfg)

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("🎯 Production server starting",
		zap.String("port", port),
		zap.Bool("redis_enabled", redisClient != nil),
		zap.String("environment", cfg.Server.Environment),
	)

	log.Printf("🎯 Production server starting on port %s", port)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("🔗 API docs: http://localhost:%s/api/docs", port)
	log.Printf("🔗 WebSocket: ws://localhost:%s/ws/chat", port)

	if err := router.Run(":" + port); err != nil {
		logger.Fatal("❌ Server startup failed", zap.Error(err))
	}
}
