---
id: testing-guide
title: Guide de tests
sidebar_label: Guide de tests
description: Guide complet des tests pour le projet Veza
---

# Guide de tests

Ce guide détaille les stratégies, outils et bonnes pratiques de tests pour le projet Veza, couvrant tous les niveaux de tests et les différents composants.

## Stratégie de tests

### Pyramide de tests
```
    /\
   /  \     Tests E2E (peu nombreux, lents)
  /____\    
 /      \   Tests d'intégration (quelques-uns, moyens)
/________\  Tests unitaires (nombreux, rapides)
```

### Types de tests

#### 1. Tests unitaires
- **Objectif** : Tester une fonction/méthode isolée
- **Vitesse** : Très rapides (< 1ms)
- **Fiabilité** : Très fiables
- **Couvrance** : 80%+ recommandée

#### 2. Tests d'intégration
- **Objectif** : Tester l'interaction entre composants
- **Vitesse** : Rapides (1-100ms)
- **Fiabilité** : Fiables
- **Couvrance** : 60%+ recommandée

#### 3. Tests de bout en bout (E2E)
- **Objectif** : Tester le système complet
- **Vitesse** : Lents (secondes-minutes)
- **Fiabilité** : Moins fiables (dépendances externes)
- **Couvrance** : 20%+ recommandée

## Tests Go (Backend API)

### Structure des tests
```go
// veza-backend-api/internal/services/user_service_test.go
package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

// Test suite pour UserService
type UserServiceTestSuite struct {
    suite.Suite
    service    *UserService
    mockRepo   *MockUserRepository
    mockAuth   *MockAuthService
    ctx        context.Context
}

func (suite *UserServiceTestSuite) SetupTest() {
    suite.mockRepo = &MockUserRepository{}
    suite.mockAuth = &MockAuthService{}
    suite.service = NewUserService(suite.mockRepo, suite.mockAuth)
    suite.ctx = context.Background()
}

func (suite *UserServiceTestSuite) TearDownTest() {
    suite.mockRepo.AssertExpectations(suite.T())
    suite.mockAuth.AssertExpectations(suite.T())
}

// Tests individuels
func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
    // Arrange
    userID := int64(123)
    expectedUser := &User{
        ID:       userID,
        Email:    "test@example.com",
        Username: "testuser",
    }
    
    suite.mockRepo.On("GetByID", suite.ctx, userID).Return(expectedUser, nil)
    
    // Act
    user, err := suite.service.GetUserByID(suite.ctx, userID)
    
    // Assert
    suite.NoError(err)
    suite.Equal(expectedUser, user)
}

func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
    // Arrange
    userID := int64(999)
    suite.mockRepo.On("GetByID", suite.ctx, userID).Return(nil, ErrUserNotFound)
    
    // Act
    user, err := suite.service.GetUserByID(suite.ctx, userID)
    
    // Assert
    suite.Error(err)
    suite.Nil(user)
    suite.Equal(ErrUserNotFound, err)
}

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
    // Arrange
    input := CreateUserInput{
        Email:    "new@example.com",
        Username: "newuser",
        Password: "password123",
    }
    
    expectedUser := &User{
        ID:       456,
        Email:    input.Email,
        Username: input.Username,
    }
    
    suite.mockRepo.On("GetByEmail", suite.ctx, input.Email).Return(nil, ErrUserNotFound)
    suite.mockRepo.On("Create", suite.ctx, mock.AnythingOfType("*User")).Return(nil)
    suite.mockAuth.On("SendWelcomeEmail", input.Email).Return(nil)
    
    // Act
    user, err := suite.service.CreateUser(suite.ctx, input)
    
    // Assert
    suite.NoError(err)
    suite.NotNil(user)
    suite.Equal(input.Email, user.Email)
    suite.Equal(input.Username, user.Username)
}

// Exécution de la suite
func TestUserServiceTestSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}
```

### Tests de handlers HTTP
```go
// veza-backend-api/internal/api/user/handler_test.go
package user

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserHandler_GetUser(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := &MockUserService{}
    handler := NewUserHandler(mockService)
    
    router := gin.New()
    router.GET("/users/:id", handler.GetUser)
    
    // Test cases
    tests := []struct {
        name           string
        userID         string
        setupMock      func()
        expectedStatus int
        expectedBody   string
    }{
        {
            name:   "success",
            userID: "123",
            setupMock: func() {
                user := &User{ID: 123, Email: "test@example.com"}
                mockService.On("GetUserByID", mock.Anything, int64(123)).Return(user, nil)
            },
            expectedStatus: http.StatusOK,
            expectedBody:   `{"id":123,"email":"test@example.com"}`,
        },
        {
            name:   "not found",
            userID: "999",
            setupMock: func() {
                mockService.On("GetUserByID", mock.Anything, int64(999)).Return(nil, ErrUserNotFound)
            },
            expectedStatus: http.StatusNotFound,
            expectedBody:   `{"error":"user not found"}`,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            tt.setupMock()
            
            // Act
            req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)
            assert.JSONEq(t, tt.expectedBody, w.Body.String())
        })
    }
}
```

### Tests de base de données
```go
// veza-backend-api/internal/adapters/postgres/user_repository_test.go
package postgres

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func TestUserRepository_Integration(t *testing.T) {
    // Setup test database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Migrate schema
    err = db.AutoMigrate(&User{})
    require.NoError(t, err)
    
    repo := NewUserRepository(db)
    ctx := context.Background()
    
    t.Run("Create and Get User", func(t *testing.T) {
        // Arrange
        user := &User{
            Email:     "test@example.com",
            Username:  "testuser",
            Password:  "hashedpassword",
        }
        
        // Act
        err := repo.Create(ctx, user)
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)
        
        // Assert
        retrieved, err := repo.GetByID(ctx, user.ID)
        assert.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
        assert.Equal(t, user.Username, retrieved.Username)
    })
    
    t.Run("Get User by Email", func(t *testing.T) {
        // Act
        user, err := repo.GetByEmail(ctx, "test@example.com")
        
        // Assert
        assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, "test@example.com", user.Email)
    })
}
```

## Tests Rust (Chat & Stream Servers)

### Tests unitaires
```rust
// veza-chat-server/src/core/user.rs
#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Utc;

    #[test]
    fn test_user_creation() {
        let user = User::new(
            "test@example.com".to_string(),
            "testuser".to_string(),
            "password123".to_string(),
        );

        assert_eq!(user.email, "test@example.com");
        assert_eq!(user.username, "testuser");
        assert!(user.created_at <= Utc::now());
        assert!(user.is_active);
    }

    #[test]
    fn test_user_validation() {
        let valid_user = User::new(
            "valid@example.com".to_string(),
            "validuser".to_string(),
            "password123".to_string(),
        );
        assert!(valid_user.validate().is_ok());

        let invalid_user = User::new(
            "invalid-email".to_string(),
            "".to_string(),
            "".to_string(),
        );
        assert!(invalid_user.validate().is_err());
    }

    #[test]
    fn test_user_password_hashing() {
        let mut user = User::new(
            "test@example.com".to_string(),
            "testuser".to_string(),
            "password123".to_string(),
        );

        let original_password = user.password.clone();
        user.hash_password();
        
        assert_ne!(user.password, original_password);
        assert!(user.verify_password("password123"));
        assert!(!user.verify_password("wrongpassword"));
    }
}
```

### Tests d'intégration
```rust
// veza-chat-server/tests/user_integration_test.rs
use veza_chat_server::core::user::{User, UserRepository};
use veza_chat_server::adapters::postgres::PostgresUserRepository;
use sqlx::PgPool;

#[tokio::test]
async fn test_user_repository_integration() {
    // Setup test database
    let pool = PgPool::connect("postgresql://test:test@localhost/test_db")
        .await
        .expect("Failed to connect to test database");

    let repo = PostgresUserRepository::new(pool);

    // Test user creation
    let user = User::new(
        "integration@example.com".to_string(),
        "integration_user".to_string(),
        "password123".to_string(),
    );

    let created_user = repo.create(&user).await.expect("Failed to create user");
    assert_eq!(created_user.email, user.email);
    assert_eq!(created_user.username, user.username);

    // Test user retrieval
    let retrieved_user = repo.get_by_id(created_user.id)
        .await
        .expect("Failed to get user")
        .expect("User not found");

    assert_eq!(retrieved_user.email, user.email);
    assert_eq!(retrieved_user.username, user.username);
}
```

### Tests de performance
```rust
// veza-chat-server/benches/user_benchmarks.rs
use criterion::{criterion_group, criterion_main, Criterion};
use veza_chat_server::core::user::User;

fn user_creation_benchmark(c: &mut Criterion) {
    c.bench_function("user_creation", |b| {
        b.iter(|| {
            User::new(
                "benchmark@example.com".to_string(),
                "benchmark_user".to_string(),
                "password123".to_string(),
            )
        })
    });
}

fn password_hashing_benchmark(c: &mut Criterion) {
    let mut user = User::new(
        "benchmark@example.com".to_string(),
        "benchmark_user".to_string(),
        "password123".to_string(),
    );

    c.bench_function("password_hashing", |b| {
        b.iter(|| {
            user.hash_password();
        })
    });
}

criterion_group!(benches, user_creation_benchmark, password_hashing_benchmark);
criterion_main!(benches);
```

## Tests JavaScript/TypeScript (Documentation)

### Tests de composants React
```typescript
// veza-docs/src/components/__tests__/HomepageFeatures.test.tsx
import React from 'react';
import { render, screen } from '@testing-library/react';
import HomepageFeatures from '../HomepageFeatures';

describe('HomepageFeatures', () => {
    test('renders all feature sections', () => {
        render(<HomepageFeatures />);
        
        expect(screen.getByText('Real-time Audio Streaming')).toBeInTheDocument();
        expect(screen.getByText('Advanced Chat System')).toBeInTheDocument();
        expect(screen.getByText('Scalable Architecture')).toBeInTheDocument();
    });

    test('displays feature descriptions', () => {
        render(<HomepageFeatures />);
        
        expect(screen.getByText(/High-quality audio streaming/)).toBeInTheDocument();
        expect(screen.getByText(/Real-time messaging/)).toBeInTheDocument();
        expect(screen.getByText(/Microservices architecture/)).toBeInTheDocument();
    });

    test('has proper accessibility attributes', () => {
        render(<HomepageFeatures />);
        
        const features = screen.getAllByRole('article');
        expect(features).toHaveLength(3);
        
        features.forEach(feature => {
            expect(feature).toHaveAttribute('aria-label');
        });
    });
});
```

### Tests d'utilitaires
```typescript
// veza-docs/src/utils/__tests__/formatting.test.ts
import { formatDuration, formatFileSize, formatDate } from '../formatting';

describe('formatting utilities', () => {
    describe('formatDuration', () => {
        test('formats seconds correctly', () => {
            expect(formatDuration(65)).toBe('1:05');
            expect(formatDuration(3661)).toBe('1:01:01');
            expect(formatDuration(0)).toBe('0:00');
        });

        test('handles negative values', () => {
            expect(formatDuration(-30)).toBe('0:00');
        });
    });

    describe('formatFileSize', () => {
        test('formats bytes correctly', () => {
            expect(formatFileSize(1024)).toBe('1 KB');
            expect(formatFileSize(1048576)).toBe('1 MB');
            expect(formatFileSize(1073741824)).toBe('1 GB');
        });

        test('handles zero and negative values', () => {
            expect(formatFileSize(0)).toBe('0 B');
            expect(formatFileSize(-1024)).toBe('0 B');
        });
    });

    describe('formatDate', () => {
        test('formats dates correctly', () => {
            const date = new Date('2024-01-15T10:30:00Z');
            expect(formatDate(date)).toBe('15 Jan 2024');
        });

        test('handles relative dates', () => {
            const now = new Date();
            const oneHourAgo = new Date(now.getTime() - 3600000);
            expect(formatDate(oneHourAgo, true)).toBe('1 hour ago');
        });
    });
});
```

## Tests de bout en bout (E2E)

### Tests API avec Postman/Newman
```json
// veza-backend-api/tests/postman/veza-api.postman_collection.json
{
  "info": {
    "name": "Veza API Tests",
    "description": "Tests E2E pour l'API Veza"
  },
  "item": [
    {
      "name": "Authentication",
      "item": [
        {
          "name": "Register User",
          "request": {
            "method": "POST",
            "url": "{{baseUrl}}/api/v1/auth/register",
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"test@example.com\",\n  \"username\": \"testuser\",\n  \"password\": \"password123\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            }
          },
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.test(\"Status code is 201\", function () {",
                  "    pm.response.to.have.status(201);",
                  "});",
                  "",
                  "pm.test(\"Response has user data\", function () {",
                  "    const response = pm.response.json();",
                  "    pm.expect(response.user).to.have.property('id');",
                  "    pm.expect(response.user.email).to.eql('test@example.com');",
                  "});",
                  "",
                  "pm.test(\"Response has tokens\", function () {",
                  "    const response = pm.response.json();",
                  "    pm.expect(response.tokens).to.have.property('access_token');",
                  "    pm.expect(response.tokens).to.have.property('refresh_token');",
                  "});"
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Tests WebSocket avec Playwright
```typescript
// veza-chat-server/tests/e2e/websocket.test.ts
import { test, expect } from '@playwright/test';

test.describe('WebSocket Chat Tests', () => {
    test('user can connect to chat room', async ({ page }) => {
        // Navigate to chat page
        await page.goto('/chat');
        
        // Wait for WebSocket connection
        await page.waitForSelector('[data-testid="connection-status"]');
        
        const status = await page.locator('[data-testid="connection-status"]').textContent();
        expect(status).toBe('Connected');
    });

    test('user can send and receive messages', async ({ page }) => {
        await page.goto('/chat');
        
        // Send a message
        await page.fill('[data-testid="message-input"]', 'Hello, world!');
        await page.click('[data-testid="send-button"]');
        
        // Wait for message to appear
        await page.waitForSelector('[data-testid="message"]');
        
        const messages = await page.locator('[data-testid="message"]').allTextContents();
        expect(messages).toContain('Hello, world!');
    });

    test('handles connection errors gracefully', async ({ page }) => {
        // Mock WebSocket to fail
        await page.route('ws://localhost:8081', route => route.abort());
        
        await page.goto('/chat');
        
        // Should show error message
        await page.waitForSelector('[data-testid="connection-error"]');
        const error = await page.locator('[data-testid="connection-error"]').textContent();
        expect(error).toContain('Connection failed');
    });
});
```

## Tests de performance

### Tests de charge avec k6
```javascript
// veza-backend-api/tests/load/api-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '2m', target: 100 }, // Ramp up
        { duration: '5m', target: 100 }, // Stay at 100 users
        { duration: '2m', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
        http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    // Test user registration
    const registerPayload = JSON.stringify({
        email: `user${Date.now()}@example.com`,
        username: `user${Date.now()}`,
        password: 'password123',
    });

    const registerRes = http.post(`${BASE_URL}/api/v1/auth/register`, registerPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(registerRes, {
        'register status is 201': (r) => r.status === 201,
        'register response time < 500ms': (r) => r.timings.duration < 500,
    });

    if (registerRes.status === 201) {
        const tokens = registerRes.json('tokens');
        
        // Test authenticated endpoints
        const headers = {
            'Authorization': `Bearer ${tokens.access_token}`,
            'Content-Type': 'application/json',
        };

        // Get user profile
        const profileRes = http.get(`${BASE_URL}/api/v1/users/me`, { headers });
        
        check(profileRes, {
            'profile status is 200': (r) => r.status === 200,
            'profile response time < 200ms': (r) => r.timings.duration < 200,
        });

        // Get tracks
        const tracksRes = http.get(`${BASE_URL}/api/v1/tracks`, { headers });
        
        check(tracksRes, {
            'tracks status is 200': (r) => r.status === 200,
            'tracks response time < 300ms': (r) => r.timings.duration < 300,
        });
    }

    sleep(1);
}
```

### Tests de stress avec Artillery
```yaml
# veza-backend-api/tests/stress/stress-test.yml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 300
      arrivalRate: 50
      name: "Sustained load"
    - duration: 120
      arrivalRate: 100
      name: "Peak load"
  defaults:
    headers:
      Content-Type: 'application/json'

scenarios:
  - name: "API Stress Test"
    weight: 100
    flow:
      - post:
          url: "/api/v1/auth/login"
          json:
            email: "{{ $randomString() }}@example.com"
            password: "password123"
          capture:
            - json: "$.tokens.access_token"
              as: "token"
      
      - get:
          url: "/api/v1/users/me"
          headers:
            Authorization: "Bearer {{ token }}"
      
      - get:
          url: "/api/v1/tracks"
          headers:
            Authorization: "Bearer {{ token }}"
      
      - think: 1
```

## Tests de sécurité

### Tests d'injection SQL
```go
// veza-backend-api/internal/services/user_service_test.go
func TestUserService_SQLInjectionProtection(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  bool
    }{
        {
            name:  "normal email",
            input: "user@example.com",
            want:  true,
        },
        {
            name:  "SQL injection attempt",
            input: "'; DROP TABLE users; --",
            want:  false,
        },
        {
            name:  "another injection attempt",
            input: "' OR '1'='1",
            want:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.input)
            if tt.want {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Tests XSS
```typescript
// veza-docs/src/utils/__tests__/sanitization.test.ts
import { sanitizeHTML, sanitizeInput } from '../sanitization';

describe('XSS Protection', () => {
    test('sanitizes HTML input', () => {
        const maliciousInputs = [
            '<script>alert("xss")</script>',
            'javascript:alert("xss")',
            '<img src=x onerror=alert("xss")>',
            '<iframe src="javascript:alert(\'xss\')"></iframe>',
        ];

        maliciousInputs.forEach(input => {
            const sanitized = sanitizeHTML(input);
            expect(sanitized).not.toContain('<script>');
            expect(sanitized).not.toContain('javascript:');
            expect(sanitized).not.toContain('onerror=');
        });
    });

    test('sanitizes user input', () => {
        const input = '<script>alert("xss")</script>Hello World';
        const sanitized = sanitizeInput(input);
        expect(sanitized).toBe('Hello World');
    });
});
```

## Configuration des tests

### Configuration Go
```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - golint
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - typecheck
    - unused
    - misspell
    - unparam
    - goconst
    - gocyclo
    - dupl
    - gosec

run:
  tests: true
  timeout: 5m

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
```

### Configuration Rust
```toml
# .clippy.toml
# Configuration pour les tests
[profile.test]
opt-level = 0
debug = true

# Configuration pour les benchmarks
[profile.bench]
opt-level = 3
debug = false

# Règles Clippy spécifiques
# none = false
# all = true
# pedantic = true
# nursery = true
# cargo = true
```

### Configuration Jest
```javascript
// jest.config.js
module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'jsdom',
    setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
    moduleNameMapping: {
        '^@/(.*)$': '<rootDir>/src/$1',
    },
    collectCoverageFrom: [
        'src/**/*.{ts,tsx}',
        '!src/**/*.d.ts',
        '!src/index.tsx',
    ],
    coverageThreshold: {
        global: {
            branches: 80,
            functions: 80,
            lines: 80,
            statements: 80,
        },
    },
};
```

## Exécution des tests

### Scripts de test
```bash
#!/bin/bash
# scripts/run-tests.sh

echo "🧪 Exécution des tests..."

# Tests Go
echo "🔍 Tests Backend API..."
cd veza-backend-api
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Tests Rust
echo "🔍 Tests Chat Server..."
cd ../veza-chat-server
cargo test --verbose
cargo tarpaulin --out Html

echo "🔍 Tests Stream Server..."
cd ../veza-stream-server
cargo test --verbose
cargo tarpaulin --out Html

# Tests JavaScript
echo "🔍 Tests Documentation..."
cd ../veza-docs
npm test -- --coverage

# Tests E2E
echo "🔍 Tests E2E..."
cd ../veza-backend-api
newman run tests/postman/veza-api.postman_collection.json

echo "✅ Tous les tests sont terminés"
```

### Intégration continue
```yaml
# .github/workflows/tests.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Set up Rust
      uses: actions-rs/toolchain@v1
      with:
        toolchain: stable
    
    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        cache: 'npm'
    
    - name: Run Go tests
      run: |
        cd veza-backend-api
        go test ./... -v -race -coverprofile=coverage.out
        go tool cover -func=coverage.out
    
    - name: Run Rust tests
      run: |
        cd veza-chat-server
        cargo test --verbose
        cargo clippy -- -D warnings
    
    - name: Run JavaScript tests
      run: |
        cd veza-docs
        npm ci
        npm test -- --coverage --watchAll=false
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        files: ./veza-backend-api/coverage.out
```

## Bonnes pratiques

### 1. Nommage des tests
```go
// ✅ Bon - Nom descriptif avec contexte
func TestUserService_GetUserByID_ReturnsUser_WhenUserExists(t *testing.T) { }
func TestUserService_GetUserByID_ReturnsError_WhenUserNotFound(t *testing.T) { }
func TestUserService_CreateUser_ValidatesEmail_AndReturnsError_WhenInvalid(t *testing.T) { }

// ❌ Mauvais - Nom générique
func TestGetUser(t *testing.T) { }
func TestCreateUser(t *testing.T) { }
```

### 2. Structure AAA (Arrange-Act-Assert)
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    input := CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    }
    
    // Act
    user, err := service.CreateUser(context.Background(), input)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, input.Email, user.Email)
}
```

### 3. Tests isolés
```go
// ✅ Bon - Test isolé avec mocks
func TestUserService_GetUserByID(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    // Test indépendant
    mockRepo.On("GetByID", mock.Anything, int64(123)).Return(&User{}, nil)
    
    user, err := service.GetUserByID(context.Background(), 123)
    assert.NoError(t, err)
    assert.NotNil(t, user)
}

// ❌ Mauvais - Test dépendant de la base de données
func TestUserService_GetUserByID(t *testing.T) {
    db := setupTestDB() // Dépendance externe
    service := NewUserService(db)
    
    // Test fragile
    user, err := service.GetUserByID(context.Background(), 123)
    // ...
}
```

### 4. Couverture de tests
```bash
# Vérifier la couverture
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Objectif : 80%+ de couverture
# Branches critiques : 100% de couverture
```

## Conclusion

Ce guide de tests couvre tous les aspects des tests pour le projet Veza. Une bonne stratégie de tests est essentielle pour maintenir la qualité et la fiabilité du code.

### Ressources supplémentaires
- [Guide de développement](./development-environment.md)
- [Standards de code](./code-review.md)
- [Architecture du projet](../architecture/backend-architecture.md) 