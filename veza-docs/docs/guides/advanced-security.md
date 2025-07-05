---
id: advanced-security
sidebar_label: Sécurité Avancée
---

# 🔒 Guide de Sécurité Avancée - Veza Platform

> **Guide complet pour sécuriser la plateforme Veza en production**

## 📋 Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Authentification et Autorisation](#authentification-et-autorisation)
- [Sécurisation des API](#scurisation-des-api)
- [Sécurité des Données](#scure-des-donnes)
- [Sécurité Réseau](#scure-rseau)
- [Monitoring et Alerting](#monitoring-et-alerting)
- [Checklist Sécurité](#checklist-scure)

## 🎯 Vue d'ensemble

Ce guide détaille les mesures de sécurité avancées pour protéger la plateforme Veza contre les menaces modernes, incluant l'authentification, l'autorisation, la sécurisation des API, la protection des données et le monitoring.

### 🛡️ Menaces Identifiées

- **Injection SQL** : Attaques contre la base de données
- **XSS (Cross-Site Scripting)** : Exécution de code malveillant
- **CSRF (Cross-Site Request Forgery)** : Requêtes non autorisées
- **DDoS (Distributed Denial of Service)** : Surcharge des services
- **Brute Force** : Tentatives de connexion répétées
- **Man-in-the-Middle** : Interception des communications
- **Privilege Escalation** : Élévation des droits d'accès

## 🔐 Authentification et Autorisation

### 1. 🔑 JWT (JSON Web Tokens)

```go
// internal/security/jwt_manager.go
package security

import (
    "crypto/rsa"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
    issuer     string
    audience   string
}

type Claims struct {
    UserID   int64    `json:"user_id"`
    Email    string   `json:"email"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims
}

func NewJWTManager(privateKeyPath, publicKeyPath string) (*JWTManager, error) {
    privateKey, err := loadPrivateKey(privateKeyPath)
    if err != nil {
        return nil, err
    }
    
    publicKey, err := loadPublicKey(publicKeyPath)
    if err != nil {
        return nil, err
    }
    
    return &JWTManager{
        privateKey: privateKey,
        publicKey:  publicKey,
        issuer:     "veza-platform",
        audience:   "veza-users",
    }, nil
}

func (jm *JWTManager) GenerateToken(userID int64, email, username string, roles []string) (string, error) {
    now := time.Now()
    claims := &Claims{
        UserID:   userID,
        Email:    email,
        Username: username,
        Roles:    roles,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    jm.issuer,
            Audience:  []string{jm.audience},
            IssuedAt:  jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
            NotBefore: jwt.NewNumericDate(now),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(jm.privateKey)
}

func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, jwt.ErrSignatureInvalid
        }
        return jm.publicKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, jwt.ErrSignatureInvalid
}
```

### 2. 🔐 OAuth 2.0 / OpenID Connect

```go
// internal/security/oauth_manager.go
package security

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type OAuthManager struct {
    config *oauth2.Config
    client *http.Client
}

func NewOAuthManager(clientID, clientSecret, redirectURL string) *OAuthManager {
    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURL:  redirectURL,
        Scopes: []string{
            "openid",
            "profile",
            "email",
        },
        Endpoint: google.Endpoint,
    }
    
    return &OAuthManager{
        config: config,
        client: &http.Client{Timeout: 10 * time.Second},
    }
}

func (om *OAuthManager) GetAuthURL(state string) string {
    return om.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (om *OAuthManager) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return om.config.Exchange(ctx, code)
}

func (om *OAuthManager) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    client := om.config.Client(ctx, token)
    
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var userInfo UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        return nil, err
    }
    
    return &userInfo, nil
}
```

### 3. 👥 RBAC (Role-Based Access Control)

```go
// internal/security/rbac.go
package security

import (
    "context"
    "fmt"
)

type Role struct {
    ID          int64    `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Permissions []string `json:"permissions"`
}

type Permission struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    Resource    string `json:"resource"`
    Action      string `json:"action"`
    Description string `json:"description"`
}

type RBACManager struct {
    roles       map[string]*Role
    permissions map[string]*Permission
}

func NewRBACManager() *RBACManager {
    return &RBACManager{
        roles:       make(map[string]*Role),
        permissions: make(map[string]*Permission),
    }
}

func (rm *RBACManager) AddRole(role *Role) {
    rm.roles[role.Name] = role
}

func (rm *RBACManager) AddPermission(permission *Permission) {
    rm.permissions[permission.Name] = permission
}

func (rm *RBACManager) HasPermission(userRoles []string, resource, action string) bool {
    for _, roleName := range userRoles {
        if role, exists := rm.roles[roleName]; exists {
            for _, permissionName := range role.Permissions {
                if permission, exists := rm.permissions[permissionName]; exists {
                    if permission.Resource == resource && permission.Action == action {
                        return true
                    }
                }
            }
        }
    }
    return false
}

// Middleware RBAC
func (rm *RBACManager) RBACMiddleware(resource, action string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extraire les rôles de l'utilisateur depuis le contexte
            userRoles := r.Context().Value("user_roles").([]string)
            
            if !rm.HasPermission(userRoles, resource, action) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## 🛡️ Sécurisation des API

### 1. 🚦 Rate Limiting

```go
// internal/security/rate_limiter.go
package security

import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    redis  *redis.Client
    limiter *rate.Limiter
}

type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize         int
    WindowSize        time.Duration
}

func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig) *RateLimiter {
    return &RateLimiter{
        redis:   redisClient,
        limiter: rate.NewLimiter(rate.Limit(config.RequestsPerMinute/60), config.BurstSize),
    }
}

func (rl *RateLimiter) IsAllowed(ctx context.Context, key string) (bool, error) {
    // Vérifier le rate limit Redis
    current, err := rl.redis.Get(ctx, key).Int()
    if err == redis.Nil {
        // Première requête
        err = rl.redis.SetEX(ctx, key, 1, time.Minute).Err()
        return err == nil, err
    } else if err != nil {
        return false, err
    }
    
    if current >= 100 { // Limite par minute
        return false, nil
    }
    
    // Incrémenter le compteur
    err = rl.redis.Incr(ctx, key).Err()
    return err == nil, err
}

// Middleware Rate Limiting
func (rl *RateLimiter) RateLimitMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Identifier le client (IP, User ID, etc.)
            clientKey := r.Header.Get("X-Forwarded-For")
            if clientKey == "" {
                clientKey = r.RemoteAddr
            }
            
            allowed, err := rl.IsAllowed(r.Context(), "rate_limit:"+clientKey)
            if err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }
            
            if !allowed {
                w.Header().Set("Retry-After", "60")
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2. 🛡️ CORS (Cross-Origin Resource Sharing)

```go
// internal/security/cors.go
package security

import (
    "net/http"
    "strings"
)

type CORSConfig struct {
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    ExposedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}

func NewCORSConfig() *CORSConfig {
    return &CORSConfig{
        AllowedOrigins: []string{
            "https://veza.app",
            "https://admin.veza.app",
            "https://api.veza.app",
        },
        AllowedMethods: []string{
            "GET", "POST", "PUT", "DELETE", "OPTIONS",
        },
        AllowedHeaders: []string{
            "Content-Type", "Authorization", "X-Requested-With",
        },
        ExposedHeaders: []string{
            "X-Total-Count", "X-Page-Count",
        },
        AllowCredentials: true,
        MaxAge:           86400, // 24 heures
    }
}

func (c *CORSConfig) CORSMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // Vérifier l'origine
            if c.isOriginAllowed(origin) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            }
            
            // Headers autorisés
            w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowedMethods, ", "))
            w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowedHeaders, ", "))
            w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.ExposedHeaders, ", "))
            w.Header().Set("Access-Control-Max-Age", string(c.MaxAge))
            
            if c.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }
            
            // Répondre immédiatement aux requêtes OPTIONS
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

func (c *CORSConfig) isOriginAllowed(origin string) bool {
    for _, allowed := range c.AllowedOrigins {
        if allowed == origin {
            return true
        }
    }
    return false
}
```

### 3. 🔍 Validation des Entrées

```go
// internal/security/validator.go
package security

import (
    "regexp"
    "strings"
    
    "github.com/go-playground/validator/v10"
)

type InputValidator struct {
    validate *validator.Validate
}

func NewInputValidator() *InputValidator {
    v := validator.New()
    
    // Validateurs personnalisés
    v.RegisterValidation("safe_string", validateSafeString)
    v.RegisterValidation("email_format", validateEmailFormat)
    v.RegisterValidation("strong_password", validateStrongPassword)
    
    return &InputValidator{validate: v}
}

func (iv *InputValidator) ValidateStruct(s interface{}) error {
    return iv.validate.Struct(s)
}

// Validateur pour les chaînes sûres
func validateSafeString(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    
    // Vérifier les caractères dangereux
    dangerousPatterns := []string{
        `<script`, `javascript:`, `onload=`, `onerror=`,
		`<iframe`, `<object`, `<embed`, `vbscript:`,
    }
    
    for _, pattern := range dangerousPatterns {
        if strings.Contains(strings.ToLower(value), pattern) {
            return false
        }
    }
    
    return true
}

// Validateur pour le format email
func validateEmailFormat(fl validator.FieldLevel) bool {
    email := fl.Field().String()
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// Validateur pour les mots de passe forts
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    if len(password) < 8 {
        return false
    }
    
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
    
    return hasUpper && hasLower && hasNumber && hasSpecial
}
```

## 🔒 Sécurité des Données

### 1. 🔐 Chiffrement des Données

```go
// internal/security/encryption.go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
)

type EncryptionManager struct {
    key []byte
}

func NewEncryptionManager(key []byte) *EncryptionManager {
    return &EncryptionManager{key: key}
}

func (em *EncryptionManager) Encrypt(plaintext []byte) (string, error) {
    block, err := aes.NewCipher(em.key)
    if err != nil {
        return "", err
    }
    
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }
    
    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
    
    return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (em *EncryptionManager) Decrypt(encryptedText string) ([]byte, error) {
    ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
    if err != nil {
        return nil, err
    }
    
    block, err := aes.NewCipher(em.key)
    if err != nil {
        return nil, err
    }
    
    if len(ciphertext) < aes.BlockSize {
        return nil, err
    }
    
    iv := ciphertext[:aes.BlockSize]
    ciphertext = ciphertext[aes.BlockSize:]
    
    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(ciphertext, ciphertext)
    
    return ciphertext, nil
}
```

### 2. 📊 Conformité RGPD

```go
// internal/security/gdpr.go
package security

import (
    "context"
    "time"
)

type GDPRManager struct {
    db Database
}

type DataRetentionPolicy struct {
    DataType     string        `json:"data_type"`
    RetentionPeriod time.Duration `json:"retention_period"`
    AutoDelete   bool          `json:"auto_delete"`
}

func NewGDPRManager(db Database) *GDPRManager {
    return &GDPRManager{db: db}
}

// Droit à l'effacement (Right to be forgotten)
func (gm *GDPRManager) DeleteUserData(ctx context.Context, userID int64) error {
    // Supprimer les données personnelles
    queries := []string{
        "DELETE FROM user_profiles WHERE user_id = $1",
        "DELETE FROM user_preferences WHERE user_id = $1",
        "DELETE FROM user_sessions WHERE user_id = $1",
        "UPDATE users SET email = NULL, phone = NULL WHERE id = $1",
    }
    
    for _, query := range queries {
        if err := gm.db.ExecContext(ctx, query, userID); err != nil {
            return err
        }
    }
    
    return nil
}

// Droit à la portabilité des données
func (gm *GDPRManager) ExportUserData(ctx context.Context, userID int64) (*UserDataExport, error) {
    var export UserDataExport
    
    // Récupérer toutes les données de l'utilisateur
    if err := gm.db.QueryRowContext(ctx, 
        "SELECT id, email, username, created_at FROM users WHERE id = $1", 
        userID).Scan(&export.User); err != nil {
        return nil, err
    }
    
    // Récupérer les messages
    rows, err := gm.db.QueryContext(ctx,
        "SELECT id, content, created_at FROM messages WHERE user_id = $1",
        userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var msg Message
        if err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt); err != nil {
            return nil, err
        }
        export.Messages = append(export.Messages, msg)
    }
    
    return &export, nil
}

// Politique de rétention des données
func (gm *GDPRManager) ApplyRetentionPolicy(ctx context.Context) error {
    policies := []DataRetentionPolicy{
        {DataType: "chat_messages", RetentionPeriod: 365 * 24 * time.Hour, AutoDelete: true},
        {DataType: "user_sessions", RetentionPeriod: 30 * 24 * time.Hour, AutoDelete: true},
        {DataType: "analytics_events", RetentionPeriod: 730 * 24 * time.Hour, AutoDelete: true},
    }
    
    for _, policy := range policies {
        cutoffDate := time.Now().Add(-policy.RetentionPeriod)
        
        query := "DELETE FROM " + policy.DataType + " WHERE created_at < $1"
        if err := gm.db.ExecContext(ctx, query, cutoffDate); err != nil {
            return err
        }
    }
    
    return nil
}
```

## 🌐 Sécurité Réseau

### 1. 🔒 TLS/SSL Configuration

```nginx
# nginx.conf - Configuration TLS sécurisée
server {
    listen 443 ssl http2;
    server_name veza.app;
    
    # Certificats SSL
    ssl_certificate /etc/ssl/certs/veza.crt;
    ssl_certificate_key /etc/ssl/private/veza.key;
    
    # Configuration SSL sécurisée
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # HSTS (HTTP Strict Transport Security)
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Headers de sécurité
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';" always;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 🛡️ Firewall Configuration

```bash
#!/bin/bash
# firewall-setup.sh

# Configuration iptables pour Veza Platform

# Règles de base
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# Politique par défaut
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Connexions établies
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Interface loopback
iptables -A INPUT -i lo -j ACCEPT

# SSH (port 22)
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -j ACCEPT

# HTTP/HTTPS
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# API Backend
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT

# Chat Server
iptables -A INPUT -p tcp --dport 8081 -j ACCEPT

# Stream Server
iptables -A INPUT -p tcp --dport 8082 -j ACCEPT

# WebSocket
iptables -A INPUT -p tcp --dport 8083 -j ACCEPT

# Protection contre DDoS
iptables -A INPUT -p tcp --dport 80 -m limit --limit 25/minute --limit-burst 100 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -m limit --limit 25/minute --limit-burst 100 -j ACCEPT

# Protection contre les scans de ports
iptables -A INPUT -p tcp --tcp-flags ALL NONE -j DROP
iptables -A INPUT -p tcp --tcp-flags ALL ALL -j DROP

# Sauvegarder les règles
iptables-save > /etc/iptables/rules.v4
```

### 3. 🔍 Audit et Logging

```go
// internal/security/audit.go
package security

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type AuditEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    UserID      int64                  `json:"user_id"`
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
    Details     map[string]interface{} `json:"details"`
    Success     bool                   `json:"success"`
    Error       string                 `json:"error,omitempty"`
}

type AuditLogger struct {
    redis *redis.Client
}

func NewAuditLogger(redis *redis.Client) *AuditLogger {
    return &AuditLogger{redis: redis}
}

func (al *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
    event.ID = generateUUID()
    event.Timestamp = time.Now()
    
    // Sérialiser l'événement
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    // Stocker dans Redis avec TTL
    key := "audit:" + event.ID
    return al.redis.SetEX(ctx, key, data, 24*time.Hour).Err()
}

func (al *AuditLogger) GetUserEvents(ctx context.Context, userID int64, limit int) ([]*AuditEvent, error) {
    // Rechercher les événements d'un utilisateur
    pattern := "audit:*"
    keys, err := al.redis.Keys(ctx, pattern).Result()
    if err != nil {
        return nil, err
    }
    
    var events []*AuditEvent
    for _, key := range keys {
        data, err := al.redis.Get(ctx, key).Result()
        if err != nil {
            continue
        }
        
        var event AuditEvent
        if err := json.Unmarshal([]byte(data), &event); err != nil {
            continue
        }
        
        if event.UserID == userID {
            events = append(events, &event)
        }
        
        if len(events) >= limit {
            break
        }
    }
    
    return events, nil
}

// Middleware d'audit
func (al *AuditLogger) AuditMiddleware(action, resource string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Capturer la réponse
            wrapped := &responseWriter{ResponseWriter: w}
            
            next.ServeHTTP(wrapped, r)
            
            // Créer l'événement d'audit
            userID := r.Context().Value("user_id").(int64)
            event := &AuditEvent{
                UserID:    userID,
                Action:    action,
                Resource:  resource,
                IPAddress: r.RemoteAddr,
                UserAgent: r.UserAgent(),
                Success:   wrapped.statusCode < 400,
                Details: map[string]interface{}{
                    "method":     r.Method,
                    "path":       r.URL.Path,
                    "duration":   time.Since(start).Milliseconds(),
                    "status":     wrapped.statusCode,
                },
            }
            
            if wrapped.statusCode >= 400 {
                event.Error = "HTTP Error"
            }
            
            // Logger l'événement
            al.LogEvent(r.Context(), event)
        })
    }
}
```

## 📊 Monitoring et Alerting

### 1. 🔍 Détection d'Anomalies

```go
// internal/security/anomaly_detector.go
package security

import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type AnomalyDetector struct {
    redis *redis.Client
}

type AnomalyRule struct {
    ID          string  `json:"id"`
    Name        string  `json:"name"`
    Threshold   int     `json:"threshold"`
    Window      int     `json:"window"` // en secondes
    Action      string  `json:"action"`
    Enabled     bool    `json:"enabled"`
}

func NewAnomalyDetector(redis *redis.Client) *AnomalyDetector {
    return &AnomalyDetector{redis: redis}
}

func (ad *AnomalyDetector) CheckAnomaly(ctx context.Context, rule *AnomalyRule, key string) (bool, error) {
    // Compter les événements dans la fenêtre
    count, err := ad.redis.Get(ctx, key).Int()
    if err == redis.Nil {
        return false, nil
    } else if err != nil {
        return false, err
    }
    
    return count > rule.Threshold, nil
}

func (ad *AnomalyDetector) RecordEvent(ctx context.Context, key string, window time.Duration) error {
    pipe := ad.redis.Pipeline()
    
    // Incrémenter le compteur
    pipe.Incr(ctx, key)
    
    // Définir l'expiration
    pipe.Expire(ctx, key, window)
    
    _, err := pipe.Exec(ctx)
    return err
}

// Règles d'anomalie prédéfinies
var DefaultAnomalyRules = []AnomalyRule{
    {
        ID:        "login_attempts",
        Name:      "Failed Login Attempts",
        Threshold: 5,
        Window:    300, // 5 minutes
        Action:    "block_ip",
        Enabled:   true,
    },
    {
        ID:        "api_requests",
        Name:      "API Request Rate",
        Threshold: 1000,
        Window:    60, // 1 minute
        Action:    "rate_limit",
        Enabled:   true,
    },
    {
        ID:        "suspicious_activity",
        Name:      "Suspicious Activity",
        Threshold: 10,
        Window:    3600, // 1 heure
        Action:    "alert",
        Enabled:   true,
    },
}
```

### 2. 🚨 Système d'Alerting

```go
// internal/security/alerting.go
package security

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type SecurityAlert struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    Level       string                 `json:"level"` // low, medium, high, critical
    Type        string                 `json:"type"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details"`
    Acknowledged bool                  `json:"acknowledged"`
    Resolved    bool                   `json:"resolved"`
}

type AlertingSystem struct {
    redis *redis.Client
}

func NewAlertingSystem(redis *redis.Client) *AlertingSystem {
    return &AlertingSystem{redis: redis}
}

func (as *AlertingSystem) CreateAlert(ctx context.Context, alert *SecurityAlert) error {
    alert.ID = generateUUID()
    alert.Timestamp = time.Now()
    
    data, err := json.Marshal(alert)
    if err != nil {
        return err
    }
    
    // Stocker l'alerte
    key := "alert:" + alert.ID
    if err := as.redis.Set(ctx, key, data, 0).Err(); err != nil {
        return err
    }
    
    // Ajouter à la liste des alertes actives
    if err := as.redis.ZAdd(ctx, "active_alerts", &redis.Z{
        Score:  float64(alert.Timestamp.Unix()),
        Member: alert.ID,
    }).Err(); err != nil {
        return err
    }
    
    // Envoyer notification si critique
    if alert.Level == "critical" {
        return as.sendCriticalNotification(ctx, alert)
    }
    
    return nil
}

func (as *AlertingSystem) GetActiveAlerts(ctx context.Context) ([]*SecurityAlert, error) {
    // Récupérer les IDs des alertes actives
    ids, err := as.redis.ZRange(ctx, "active_alerts", 0, -1).Result()
    if err != nil {
        return nil, err
    }
    
    var alerts []*SecurityAlert
    for _, id := range ids {
        data, err := as.redis.Get(ctx, "alert:"+id).Result()
        if err != nil {
            continue
        }
        
        var alert SecurityAlert
        if err := json.Unmarshal([]byte(data), &alert); err != nil {
            continue
        }
        
        if !alert.Resolved {
            alerts = append(alerts, &alert)
        }
    }
    
    return alerts, nil
}

func (as *AlertingSystem) sendCriticalNotification(ctx context.Context, alert *SecurityAlert) error {
    // Envoyer notification Slack
    notification := map[string]interface{}{
        "text": "🚨 CRITICAL SECURITY ALERT",
        "attachments": []map[string]interface{}{
            {
                "color": "danger",
                "fields": []map[string]interface{}{
                    {"title": "Type", "value": alert.Type, "short": true},
                    {"title": "Level", "value": alert.Level, "short": true},
                    {"title": "Message", "value": alert.Message, "short": false},
                    {"title": "Timestamp", "value": alert.Timestamp.Format(time.RFC3339), "short": true},
                },
            },
        },
    }
    
    // Envoyer via webhook Slack
    return sendSlackNotification(notification)
}
```

## ✅ Checklist Sécurité

### 🔐 Authentification
- [ ] JWT avec clés RSA asymétriques
- [ ] Refresh tokens avec rotation
- [ ] OAuth 2.0 / OpenID Connect
- [ ] MFA (Multi-Factor Authentication)
- [ ] Politique de mots de passe forts
- [ ] Gestion des sessions sécurisée

### 🛡️ Autorisation
- [ ] RBAC (Role-Based Access Control)
- [ ] Permissions granulaires
- [ ] Vérification des permissions à chaque requête
- [ ] Audit des accès
- [ ] Principe du moindre privilège

### 🌐 Sécurité des API
- [ ] Rate limiting par IP/utilisateur
- [ ] Validation stricte des entrées
- [ ] Protection CSRF
- [ ] Headers de sécurité (CORS, CSP, etc.)
- [ ] Sanitisation des données

### 🔒 Protection des Données
- [ ] Chiffrement en transit (TLS 1.3)
- [ ] Chiffrement au repos
- [ ] Hachage des mots de passe (bcrypt/argon2)
- [ ] Conformité RGPD
- [ ] Politique de rétention des données

### 🛡️ Sécurité Réseau
- [ ] Firewall configuré
- [ ] TLS/SSL obligatoire
- [ ] Headers de sécurité
- [ ] Protection DDoS
- [ ] Monitoring réseau

### 📊 Monitoring
- [ ] Logs d'audit complets
- [ ] Détection d'anomalies
- [ ] Alerting en temps réel
- [ ] Métriques de sécurité
- [ ] Tableaux de bord

### 🔄 Maintenance
- [ ] Mises à jour de sécurité automatiques
- [ ] Tests de pénétration réguliers
- [ ] Audit de sécurité trimestriel
- [ ] Formation sécurité équipe
- [ ] Plan de réponse aux incidents

---

## 🔗 Liens croisés

- [Architecture Globale](../architecture/global-architecture.md)
- [Monitoring](../monitoring/metrics/metrics-overview.md)
- [Déploiement](../deployment/README.md)
- [Troubleshooting](../troubleshooting/README.md)

---

## Pour aller plus loin

- [Guide de Performance](../guides/performance-optimization.md)
- [Configuration Avancée](../guides/advanced-configuration.md)
- [Tests](../testing/README.md)
- [API Reference](../api/README.md) 