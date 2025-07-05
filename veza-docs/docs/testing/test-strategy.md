---
id: test-strategy
title: Stratégie de Tests
sidebar_label: Stratégie de Tests
description: Stratégie complète de tests pour la plateforme Veza
---

# Stratégie de Tests - Veza Platform

Ce document définit la stratégie de tests complète pour assurer la qualité et la fiabilité de la plateforme Veza.

## Vue d'ensemble

### Pyramide de tests
```mermaid
graph TB
    subgraph "Tests E2E"
        E2E[🧪 Tests End-to-End<br/>10% du temps]
    end
    
    subgraph "Tests d'Intégration"
        INT[🔗 Tests d'Intégration<br/>20% du temps]
    end
    
    subgraph "Tests Unitaires"
        UNIT[⚡ Tests Unitaires<br/>70% du temps]
    end
    
    E2E --> INT
    INT --> UNIT
```

### Types de tests
```mermaid
graph LR
    subgraph "Tests Fonctionnels"
        UT[Tests Unitaires]
        IT[Tests d'Intégration]
        E2E[Tests E2E]
    end
    
    subgraph "Tests Non-Fonctionnels"
        PERF[Tests de Performance]
        SEC[Tests de Sécurité]
        LOAD[Tests de Charge]
        STRESS[Tests de Stress]
    end
    
    subgraph "Tests Automatisés"
        CI[Tests CI/CD]
        REG[Tests de Régression]
        SMOKE[Tests Smoke]
    end
```

## Tests Unitaires

### Principes
- **Isolation** : Chaque test est indépendant
- **Rapidité** : Exécution en < 1 seconde
- **Déterminisme** : Résultats prévisibles
- **Cohérence** : Même résultat à chaque exécution

### Exemples Go
```go
// internal/services/user_service_test.go
package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/google/uuid"
    
    "veza/internal/domain/entities"
    "veza/internal/domain/repositories"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*entities.User), args.Error(1)
}

func TestUserService_GetUserByID(t *testing.T) {
    tests := []struct {
        name    string
        id      uuid.UUID
        setup   func(*MockUserRepository)
        want    *entities.User
        wantErr bool
    }{
        {
            name: "successful user retrieval",
            id:   uuid.New(),
            setup: func(repo *MockUserRepository) {
                expectedUser := &entities.User{
                    ID:        uuid.New(),
                    Email:     "test@example.com",
                    Username:  "testuser",
                    CreatedAt: time.Now(),
                    UpdatedAt: time.Now(),
                }
                repo.On("GetByID", mock.Anything, uuid.New()).Return(expectedUser, nil)
            },
            want: &entities.User{
                Email:    "test@example.com",
                Username: "testuser",
            },
            wantErr: false,
        },
        {
            name: "user not found",
            id:   uuid.New(),
            setup: func(repo *MockUserRepository) {
                repo.On("GetByID", mock.Anything, uuid.New()).Return(nil, repositories.ErrUserNotFound)
            },
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockUserRepository{}
            service := NewUserService(mockRepo)
            
            if tt.setup != nil {
                tt.setup(mockRepo)
            }
            
            got, err := service.GetUserByID(context.Background(), tt.id)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, got)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, got)
                assert.Equal(t, tt.want.Email, got.Email)
                assert.Equal(t, tt.want.Username, got.Username)
            }
            
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Exemples Rust
```rust
// src/services/user_service.rs
#[cfg(test)]
mod tests {
    use super::*;
    use tokio::test;
    use uuid::Uuid;
    
    #[test]
    async fn test_get_user_by_id_success() {
        let mock_repo = MockUserRepository::new();
        let service = UserService::new(Box::new(mock_repo));
        
        let user_id = Uuid::new_v4();
        let expected_user = User {
            id: user_id,
            email: "test@example.com".to_string(),
            username: "testuser".to_string(),
            created_at: Utc::now(),
            updated_at: Utc::now(),
        };
        
        mock_repo.expect_get_by_id()
            .with(mock::eq(user_id))
            .times(1)
            .returning(move |_| Ok(expected_user.clone()));
        
        let result = service.get_user_by_id(user_id).await;
        
        assert!(result.is_ok());
        let user = result.unwrap();
        assert_eq!(user.email, "test@example.com");
        assert_eq!(user.username, "testuser");
    }
    
    #[test]
    async fn test_get_user_by_id_not_found() {
        let mock_repo = MockUserRepository::new();
        let service = UserService::new(Box::new(mock_repo));
        
        let user_id = Uuid::new_v4();
        
        mock_repo.expect_get_by_id()
            .with(mock::eq(user_id))
            .times(1)
            .returning(|_| Err(Error::UserNotFound));
        
        let result = service.get_user_by_id(user_id).await;
        
        assert!(result.is_err());
        match result.unwrap_err() {
            Error::UserNotFound => {},
            _ => panic!("Expected UserNotFound error"),
        }
    }
}
```

## Tests d'Intégration

### Principes
- **Base de données réelle** : Tests avec DB de test
- **Services externes** : Mocks pour les APIs externes
- **Configuration** : Environnement de test dédié
- **Nettoyage** : Reset après chaque test

### Exemples Go
```go
// internal/api/handlers/user_handler_integration_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "veza/internal/domain/entities"
)

func TestUserHandler_CreateUser_Integration(t *testing.T) {
    // Setup
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    router := gin.New()
    handler := NewUserHandler(db)
    router.POST("/users", handler.CreateUser)
    
    // Test data
    userData := map[string]interface{}{
        "email":    "test@example.com",
        "username": "testuser",
        "password": "password123",
    }
    
    jsonData, _ := json.Marshal(userData)
    
    // Execute
    req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    
    assert.NotEmpty(t, response["id"])
    assert.Equal(t, userData["email"], response["email"])
    assert.Equal(t, userData["username"], response["username"])
}

func setupTestDatabase(t *testing.T) *sql.DB {
    // Configuration de la base de données de test
    db, err := sql.Open("postgres", "postgres://test:test@localhost:5432/veza_test?sslmode=disable")
    require.NoError(t, err)
    
    // Migrations
    err = runMigrations(db)
    require.NoError(t, err)
    
    return db
}

func cleanupTestDatabase(t *testing.T, db *sql.DB) {
    // Nettoyage des tables
    tables := []string{"users", "chat_rooms", "messages"}
    for _, table := range tables {
        _, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
        require.NoError(t, err)
    }
    
    db.Close()
}
```

### Exemples Rust
```rust
// tests/integration/user_service_test.rs
use veza_chat_server::services::UserService;
use veza_chat_server::repositories::UserRepository;
use veza_chat_server::database::setup_test_db;

#[tokio::test]
async fn test_create_user_integration() {
    // Setup
    let db = setup_test_db().await;
    let repo = UserRepository::new(db.clone());
    let service = UserService::new(Box::new(repo));
    
    // Test data
    let user_data = CreateUserRequest {
        email: "test@example.com".to_string(),
        username: "testuser".to_string(),
        password: "password123".to_string(),
    };
    
    // Execute
    let result = service.create_user(user_data).await;
    
    // Assert
    assert!(result.is_ok());
    let user = result.unwrap();
    assert_eq!(user.email, "test@example.com");
    assert_eq!(user.username, "testuser");
    
    // Verify in database
    let saved_user = service.get_user_by_id(user.id).await.unwrap();
    assert_eq!(saved_user.email, "test@example.com");
}

async fn setup_test_db() -> Pool<Postgres> {
    let database_url = std::env::var("TEST_DATABASE_URL")
        .unwrap_or_else(|_| "postgres://test:test@localhost:5432/veza_test".to_string());
    
    let pool = PgPoolOptions::new()
        .max_connections(5)
        .connect(&database_url)
        .await
        .expect("Failed to connect to test database");
    
    // Run migrations
    sqlx::migrate!("./migrations")
        .run(&pool)
        .await
        .expect("Failed to run migrations");
    
    pool
}
```

## Tests End-to-End

### Principes
- **Scénarios complets** : Parcours utilisateur complet
- **Environnement de test** : Environnement dédié
- **Données de test** : Données contrôlées
- **Assertions** : Vérifications multi-couches

### Exemples avec Playwright
```typescript
// tests/e2e/user-flow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('User Flow', () => {
    test('complete user registration and login', async ({ page }) => {
        // Navigate to registration page
        await page.goto('/register');
        
        // Fill registration form
        await page.fill('[data-testid="email-input"]', 'test@example.com');
        await page.fill('[data-testid="username-input"]', 'testuser');
        await page.fill('[data-testid="password-input"]', 'password123');
        await page.click('[data-testid="register-button"]');
        
        // Verify successful registration
        await expect(page).toHaveURL('/dashboard');
        await expect(page.locator('[data-testid="welcome-message"]'))
            .toContainText('Welcome, testuser');
        
        // Logout
        await page.click('[data-testid="logout-button"]');
        await expect(page).toHaveURL('/login');
        
        // Login with created account
        await page.fill('[data-testid="email-input"]', 'test@example.com');
        await page.fill('[data-testid="password-input"]', 'password123');
        await page.click('[data-testid="login-button"]');
        
        // Verify successful login
        await expect(page).toHaveURL('/dashboard');
        await expect(page.locator('[data-testid="user-menu"]'))
            .toContainText('testuser');
    });
    
    test('chat functionality', async ({ page, context }) => {
        // Login as first user
        await page.goto('/login');
        await page.fill('[data-testid="email-input"]', 'user1@example.com');
        await page.fill('[data-testid="password-input"]', 'password123');
        await page.click('[data-testid="login-button"]');
        
        // Create a new chat room
        await page.click('[data-testid="create-room-button"]');
        await page.fill('[data-testid="room-name-input"]', 'Test Room');
        await page.click('[data-testid="create-room-submit"]');
        
        // Verify room creation
        await expect(page.locator('[data-testid="room-title"]'))
            .toContainText('Test Room');
        
        // Send a message
        await page.fill('[data-testid="message-input"]', 'Hello, world!');
        await page.click('[data-testid="send-message-button"]');
        
        // Verify message appears
        await expect(page.locator('[data-testid="message-list"]'))
            .toContainText('Hello, world!');
        
        // Open second browser for second user
        const page2 = await context.newPage();
        await page2.goto('/login');
        await page2.fill('[data-testid="email-input"]', 'user2@example.com');
        await page2.fill('[data-testid="password-input"]', 'password123');
        await page2.click('[data-testid="login-button"]');
        
        // Join the same room
        await page2.click('[data-testid="join-room-button"]');
        await page2.click('[data-testid="room-item-Test Room"]');
        
        // Verify both users see the message
        await expect(page2.locator('[data-testid="message-list"]'))
            .toContainText('Hello, world!');
    });
});
```

## Tests de Performance

### Principes
- **Métriques claires** : Temps de réponse, throughput
- **Scénarios réalistes** : Charge représentative
- **Monitoring** : Métriques système et application
- **Seuils** : Objectifs de performance définis

### Exemples avec k6
```javascript
// tests/performance/api-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
    stages: [
        { duration: '2m', target: 100 }, // Ramp up
        { duration: '5m', target: 100 }, // Stay at 100 users
        { duration: '2m', target: 0 },   // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
        http_req_failed: ['rate<0.1'],    // Error rate must be below 10%
        errors: ['rate<0.1'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function() {
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${__ENV.API_TOKEN}`,
        },
    };
    
    // Test user creation
    const userData = {
        email: `user${Date.now()}@example.com`,
        username: `user${Date.now()}`,
        password: 'password123',
    };
    
    const createResponse = http.post(`${BASE_URL}/api/users`, JSON.stringify(userData), params);
    
    check(createResponse, {
        'user creation status is 201': (r) => r.status === 201,
        'user creation time < 500ms': (r) => r.timings.duration < 500,
    });
    
    if (createResponse.status !== 201) {
        errorRate.add(1);
    }
    
    // Test user retrieval
    const userId = createResponse.json('id');
    const getResponse = http.get(`${BASE_URL}/api/users/${userId}`, params);
    
    check(getResponse, {
        'user retrieval status is 200': (r) => r.status === 200,
        'user retrieval time < 200ms': (r) => r.timings.duration < 200,
    });
    
    if (getResponse.status !== 200) {
        errorRate.add(1);
    }
    
    sleep(1);
}
```

### Exemples avec Artillery
```yaml
# tests/performance/chat-load-test.yml
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 300
      arrivalRate: 50
      name: "Sustained load"
    - duration: 60
      arrivalRate: 100
      name: "Peak load"
  defaults:
    headers:
      Content-Type: 'application/json'
      Authorization: 'Bearer {{ $randomString() }}'

scenarios:
  - name: "Chat room operations"
    weight: 70
    flow:
      - post:
          url: "/api/chat/rooms"
          json:
            name: "Test Room {{ $randomString() }}"
            description: "Performance test room"
      - get:
          url: "/api/chat/rooms"
      - post:
          url: "/api/chat/rooms/{{ roomId }}/messages"
          json:
            content: "Test message {{ $randomString() }}"
      - get:
          url: "/api/chat/rooms/{{ roomId }}/messages"

  - name: "User operations"
    weight: 30
    flow:
      - post:
          url: "/api/users"
          json:
            email: "user{{ $randomString() }}@example.com"
            username: "user{{ $randomString() }}"
            password: "password123"
      - get:
          url: "/api/users/{{ userId }}"
```

## Tests de Sécurité

### Principes
- **Vulnérabilités communes** : OWASP Top 10
- **Authentification** : Tests de bypass
- **Autorisation** : Tests d'accès non autorisé
- **Validation** : Tests d'injection

### Exemples avec OWASP ZAP
```python
# tests/security/security_scan.py
import subprocess
import json
import sys

def run_zap_scan(target_url):
    """Run OWASP ZAP security scan"""
    
    # Start ZAP daemon
    subprocess.run([
        'zap.sh', '-daemon',
        '-port', '8090',
        '-config', 'api.addrs.addr.name=.*',
        '-config', 'api.addrs.addr.regex=true',
        '-config', 'api.key=your-api-key'
    ])
    
    # Run spider scan
    spider_result = subprocess.run([
        'zap-cli', '--port', '8090',
        'spider', target_url
    ], capture_output=True, text=True)
    
    # Run active scan
    active_result = subprocess.run([
        'zap-cli', '--port', '8090',
        'active-scan', target_url
    ], capture_output=True, text=True)
    
    # Generate report
    report_result = subprocess.run([
        'zap-cli', '--port', '8090',
        'report', '--output-format', 'json',
        '--output', 'security-report.json'
    ], capture_output=True, text=True)
    
    return 'security-report.json'

def analyze_security_report(report_file):
    """Analyze security scan results"""
    with open(report_file, 'r') as f:
        report = json.load(f)
    
    high_issues = []
    medium_issues = []
    
    for site in report.get('site', []):
        for alert in site.get('alerts', []):
            if alert.get('risk') == 'High':
                high_issues.append({
                    'name': alert.get('name'),
                    'description': alert.get('description'),
                    'url': alert.get('url')
                })
            elif alert.get('risk') == 'Medium':
                medium_issues.append({
                    'name': alert.get('name'),
                    'description': alert.get('description'),
                    'url': alert.get('url')
                })
    
    return high_issues, medium_issues

if __name__ == '__main__':
    target_url = sys.argv[1] if len(sys.argv) > 1 else 'http://localhost:8080'
    
    print(f"Running security scan on {target_url}")
    report_file = run_zap_scan(target_url)
    
    high_issues, medium_issues = analyze_security_report(report_file)
    
    print(f"\nSecurity Scan Results:")
    print(f"High risk issues: {len(high_issues)}")
    print(f"Medium risk issues: {len(medium_issues)}")
    
    if high_issues:
        print("\nHigh risk issues found:")
        for issue in high_issues:
            print(f"- {issue['name']}: {issue['description']}")
    
    if medium_issues:
        print("\nMedium risk issues found:")
        for issue in medium_issues:
            print(f"- {issue['name']}: {issue['description']}")
```

## Automatisation CI/CD

### GitHub Actions
```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run unit tests
      run: |
        cd veza-backend-api
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./veza-backend-api/coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: veza_test
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
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run integration tests
      run: |
        cd veza-backend-api
        DATABASE_URL="postgres://postgres:postgres@localhost:5432/veza_test?sslmode=disable" \
        REDIS_URL="redis://localhost:6379" \
        go test -v -tags=integration ./...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
    
    - name: Install dependencies
      run: npm ci
    
    - name: Start application
      run: |
        docker-compose up -d
        sleep 30
    
    - name: Run E2E tests
      run: npm run test:e2e
    
    - name: Upload screenshots
      uses: actions/upload-artifact@v3
      if: failure()
      with:
        name: e2e-screenshots
        path: tests/e2e/screenshots/

  performance-tests:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '18'
    
    - name: Install k6
      run: |
        sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
        echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
        sudo apt-get update
        sudo apt-get install k6
    
    - name: Start application
      run: |
        docker-compose up -d
        sleep 30
    
    - name: Run performance tests
      run: k6 run tests/performance/api-load-test.js
```

## Métriques et Reporting

### Couverture de code
```bash
# Go
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Rust
cargo tarpaulin --out Html

# TypeScript
npm run test:coverage
```

### Métriques de qualité
```yaml
# .github/workflows/quality.yml
name: Code Quality

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  quality-gates:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Run SonarQube analysis
      uses: sonarqube-quality-gate-action@master
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
      with:
        scannerHome: ${{ github.workspace }}/sonar-scanner
        args: >
          -Dsonar.projectKey=veza-platform
          -Dsonar.sources=.
          -Dsonar.host.url=${{ secrets.SONAR_HOST_URL }}
          -Dsonar.login=${{ secrets.SONAR_TOKEN }}
    
    - name: Quality Gate
      uses: sonarqube-quality-gate-action@master
      env:
        SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
      with:
        scannerHome: ${{ github.workspace }}/sonar-scanner
```

## Conclusion

Cette stratégie de tests assure la qualité et la fiabilité de la plateforme Veza à travers différents niveaux de tests, de l'unité aux tests E2E, en passant par les tests de performance et de sécurité.

### Ressources supplémentaires
- [Guide de développement](./development-environment.md)
- [Guide de monitoring](../monitoring/README.md)
- [Guide de sécurité](../security/README.md) 