---
title: Bonnes Pratiques Sécurité
sidebar_label: Sécurité
---

# 🔒 Bonnes Pratiques Sécurité

Ce guide présente les recommandations de sécurité pour Veza.

# Guidelines de Sécurité - Veza Platform

## Vue d'ensemble

Ce guide détaille les guidelines de sécurité pour la plateforme Veza, couvrant l'authentification, l'autorisation, la protection des données et les bonnes pratiques.

## Table des matières

- [Authentification](#authentification)
- [Autorisation](#autorisation)
- [Protection des Données](#protection-des-données)
- [Bonnes Pratiques](#bonnes-pratiques)
- [Pièges à Éviter](#pièges-à-éviter)
- [Ressources](#ressources)

## Authentification

### 1. JWT Implementation

```go
// security-guidelines/auth/jwt.go
package auth

import (
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v4"
)

type JWTService struct {
    privateKey *rsa.PrivateKey
    publicKey  *rsa.PublicKey
}

func NewJWTService(privateKeyPath, publicKeyPath string) (*JWTService, error) {
    privateKey, err := loadPrivateKey(privateKeyPath)
    if err != nil {
        return nil, err
    }
    
    publicKey, err := loadPublicKey(publicKeyPath)
    if err != nil {
        return nil, err
    }
    
    return &JWTService{
        privateKey: privateKey,
        publicKey:  publicKey,
    }, nil
}

func (j *JWTService) GenerateToken(userID string, roles []string) (string, error) {
    claims := jwt.MapClaims{
        "user_id": userID,
        "roles":   roles,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
        "iat":     time.Now().Unix(),
        "iss":     "veza-platform",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    return token.SignedString(j.privateKey)
}

func (j *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return j.publicKey, nil
    })
}
```

### 2. Password Hashing

```go
// security-guidelines/auth/password.go
package auth

import (
    "crypto/rand"
    "encoding/base64"
    "golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
    cost int
}

func NewPasswordService() *PasswordService {
    return &PasswordService{
        cost: bcrypt.DefaultCost,
    }
}

func (p *PasswordService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
    return string(bytes), err
}

func (p *PasswordService) CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func (p *PasswordService) GenerateSecureToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}
```

## Autorisation

### 1. RBAC Implementation

```go
// security-guidelines/authorization/rbac.go
package authorization

import (
    "context"
    "errors"
)

type Role string

const (
    RoleAdmin    Role = "admin"
    RoleUser     Role = "user"
    RoleModerator Role = "moderator"
)

type Permission string

const (
    PermissionRead   Permission = "read"
    PermissionWrite  Permission = "write"
    PermissionDelete Permission = "delete"
    PermissionAdmin  Permission = "admin"
)

type RBACService struct {
    rolePermissions map[Role][]Permission
}

func NewRBACService() *RBACService {
    return &RBACService{
        rolePermissions: map[Role][]Permission{
            RoleAdmin: {
                PermissionRead, PermissionWrite, PermissionDelete, PermissionAdmin,
            },
            RoleUser: {
                PermissionRead, PermissionWrite,
            },
            RoleModerator: {
                PermissionRead, PermissionWrite, PermissionDelete,
            },
        },
    }
}

func (r *RBACService) HasPermission(userRoles []Role, requiredPermission Permission) bool {
    for _, role := range userRoles {
        permissions, exists := r.rolePermissions[role]
        if !exists {
            continue
        }
        
        for _, permission := range permissions {
            if permission == requiredPermission {
                return true
            }
        }
    }
    return false
}

func (r *RBACService) RequirePermission(permission Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRoles := getUserRolesFromContext(r.Context())
            
            if !r.HasPermission(userRoles, permission) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Protection des Données

### 1. Encryption

```go
// security-guidelines/encryption/encryption.go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
)

type EncryptionService struct {
    key []byte
}

func NewEncryptionService(key []byte) (*EncryptionService, error) {
    if len(key) != 32 {
        return nil, errors.New("key must be 32 bytes")
    }
    
    return &EncryptionService{key: key}, nil
}

func (e *EncryptionService) Encrypt(data []byte) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *EncryptionService) Decrypt(encryptedData string) ([]byte, error) {
    data, err := base64.StdEncoding.DecodeString(encryptedData)
    if err != nil {
        return nil, err
    }
    
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, errors.New("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

## Bonnes Pratiques

### 1. Security Headers

```go
// security-guidelines/middleware/security-headers.go
package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Headers de sécurité
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        next.ServeHTTP(w, r)
    })
}
```

### 2. Input Validation

```go
// security-guidelines/validation/input-validation.go
package validation

import (
    "regexp"
    "strings"
)

type Validator struct{}

func NewValidator() *Validator {
    return &Validator{}
}

func (v *Validator) ValidateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

func (v *Validator) ValidatePassword(password string) (bool, string) {
    if len(password) < 8 {
        return false, "Password must be at least 8 characters long"
    }
    
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        return false, "Password must contain at least one uppercase letter"
    }
    
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        return false, "Password must contain at least one lowercase letter"
    }
    
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        return false, "Password must contain at least one number"
    }
    
    return true, ""
}

func (v *Validator) SanitizeInput(input string) string {
    // Suppression des caractères dangereux
    input = strings.ReplaceAll(input, "<script>", "")
    input = strings.ReplaceAll(input, "</script>", "")
    input = strings.ReplaceAll(input, "javascript:", "")
    
    return strings.TrimSpace(input)
}
```

## Pièges à Éviter

### 1. SQL Injection

❌ **Mauvais** :
```go
// Vulnérable aux injections SQL
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
```

✅ **Bon** :
```go
// Utilisation de requêtes préparées
query := "SELECT * FROM users WHERE id = $1"
rows, err := db.Query(query, userID)
```

### 2. XSS

❌ **Mauvais** :
```go
// Vulnérable aux XSS
w.Write([]byte(userInput))
```

✅ **Bon** :
```go
// Échappement des données
import "html"
safeOutput := html.EscapeString(userInput)
w.Write([]byte(safeOutput))
```

### 3. Weak Passwords

❌ **Mauvais** :
```go
// Pas de validation de mot de passe
func createUser(password string) {
    // Création sans validation
}
```

✅ **Bon** :
```go
// Validation du mot de passe
func createUser(password string) error {
    validator := NewValidator()
    if valid, msg := validator.ValidatePassword(password); !valid {
        return errors.New(msg)
    }
    // Création avec validation
}
```

## Ressources

### Documentation Interne

- [Guide de Sécurité](../security/README.md)
- [Guide d'Authentification](./authentication-setup.md)
- [Guide d'Audit](./audit-logging.md)

### Outils Recommandés

- **OWASP ZAP** : Tests de sécurité
- **SonarQube** : Analyse de code
- **Vault** : Gestion des secrets
- **Let's Encrypt** : Certificats SSL

### Commandes Utiles

```bash
# Scan de sécurité
zap-baseline.py -t https://veza.com

# Analyse de code
sonar-scanner

# Test de vulnérabilités
npm audit
go list -json -deps ./... | nancy sleuth
```

---

**Dernière mise à jour** : $(date)
**Version du guide** : 1.0.0
**Mainteneur** : Équipe Sécurité Veza 