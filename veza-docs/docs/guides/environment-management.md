---
title: Gestion des Environnements
sidebar_label: Environnements
---

# 🌎 Gestion des Environnements

Ce guide explique la gestion des environnements sur Veza.

## Vue d'ensemble

Ce guide détaille la gestion des environnements pour la plateforme Veza, couvrant la configuration, le déploiement et la maintenance des différents environnements.

## Table des matières

- [Types d'Environnements](#types-denvironnements)
- [Configuration](#configuration)
- [Déploiement](#déploiement)
- [Bonnes Pratiques](#bonnes-pratiques)
- [Pièges à Éviter](#pièges-à-éviter)
- [Ressources](#ressources)

## Types d'Environnements

### 1. Environnements de Développement

```yaml
# environment-management/environments/development.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-development-env
  namespace: veza
data:
  # Configuration développement
  environment:
    name: "development"
    type: "dev"
    domain: "dev.veza.com"
  
  # Services
  services:
    backend_api:
      port: 8080
      replicas: 1
      resources:
        cpu: "100m"
        memory: "128Mi"
    
    chat_server:
      port: 8081
      replicas: 1
      resources:
        cpu: "100m"
        memory: "128Mi"
    
    stream_server:
      port: 8082
      replicas: 1
      resources:
        cpu: "100m"
        memory: "128Mi"
  
  # Base de données
  database:
    host: "localhost"
    port: 5432
    name: "veza_dev"
    user: "veza_dev"
    password: "dev_password"
  
  # Cache
  cache:
    host: "localhost"
    port: 6379
    password: ""
  
  # Logging
  logging:
    level: "debug"
    format: "json"
    output: "stdout"
  
  # Monitoring
  monitoring:
    enabled: true
    prometheus_port: 9090
    grafana_port: 3000
```

### 2. Environnements de Staging

```yaml
# environment-management/environments/staging.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-staging-env
  namespace: veza
data:
  # Configuration staging
  environment:
    name: "staging"
    type: "staging"
    domain: "staging.veza.com"
  
  # Services
  services:
    backend_api:
      port: 8080
      replicas: 2
      resources:
        cpu: "500m"
        memory: "512Mi"
    
    chat_server:
      port: 8081
      replicas: 2
      resources:
        cpu: "500m"
        memory: "512Mi"
    
    stream_server:
      port: 8082
      replicas: 2
      resources:
        cpu: "500m"
        memory: "512Mi"
  
  # Base de données
  database:
    host: "staging-db.veza.com"
    port: 5432
    name: "veza_staging"
    user: "veza_staging"
    password: "${DB_PASSWORD}"
  
  # Cache
  cache:
    host: "staging-redis.veza.com"
    port: 6379
    password: "${REDIS_PASSWORD}"
  
  # Logging
  logging:
    level: "info"
    format: "json"
    output: "file"
    file_path: "/var/log/veza"
  
  # Monitoring
  monitoring:
    enabled: true
    prometheus_port: 9090
    grafana_port: 3000
    alerting: true
```

### 3. Environnements de Production

```yaml
# environment-management/environments/production.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-production-env
  namespace: veza
data:
  # Configuration production
  environment:
    name: "production"
    type: "prod"
    domain: "veza.com"
  
  # Services
  services:
    backend_api:
      port: 8080
      replicas: 5
      resources:
        cpu: "1000m"
        memory: "1Gi"
      autoscaling:
        min_replicas: 3
        max_replicas: 10
        target_cpu_utilization: 70
    
    chat_server:
      port: 8081
      replicas: 3
      resources:
        cpu: "1000m"
        memory: "1Gi"
      autoscaling:
        min_replicas: 2
        max_replicas: 8
        target_cpu_utilization: 70
    
    stream_server:
      port: 8082
      replicas: 3
      resources:
        cpu: "1000m"
        memory: "1Gi"
      autoscaling:
        min_replicas: 2
        max_replicas: 8
        target_cpu_utilization: 70
  
  # Base de données
  database:
    host: "prod-db.veza.com"
    port: 5432
    name: "veza_production"
    user: "veza_production"
    password: "${DB_PASSWORD}"
    ssl_mode: "require"
    connection_pool:
      max_connections: 100
      min_connections: 10
  
  # Cache
  cache:
    host: "prod-redis.veza.com"
    port: 6379
    password: "${REDIS_PASSWORD}"
    cluster_mode: true
    replicas: 3
  
  # Logging
  logging:
    level: "warn"
    format: "json"
    output: "file"
    file_path: "/var/log/veza"
    rotation:
      max_size: "100MB"
      max_age: "7d"
      max_backups: 10
  
  # Monitoring
  monitoring:
    enabled: true
    prometheus_port: 9090
    grafana_port: 3000
    alerting: true
    tracing: true
    metrics_retention: "30d"
```

## Configuration

### 1. Gestion des Variables d'Environnement

```bash
#!/bin/bash
# environment-management/scripts/env-manager.sh

# Configuration
ENV_DIR="environments"
SECRETS_DIR="secrets"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Fonctions de logging
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validation d'environnement
validate_environment() {
    local env_name="$1"
    
    if [ ! -f "$ENV_DIR/$env_name.yaml" ]; then
        log_error "Environnement $env_name non trouvé"
        return 1
    fi
    
    log_info "Environnement $env_name validé"
    return 0
}

# Génération des variables d'environnement
generate_env_vars() {
    local env_name="$1"
    
    log_info "Génération des variables pour $env_name"
    
    # Lecture de la configuration
    local config_file="$ENV_DIR/$env_name.yaml"
    
    # Extraction des variables
    yq eval '.data.environment | to_entries | .[] | .key + "=" + .value' "$config_file" > ".env.$env_name"
    
    # Ajout des secrets
    if [ -f "$SECRETS_DIR/$env_name.env" ]; then
        cat "$SECRETS_DIR/$env_name.env" >> ".env.$env_name"
    fi
    
    log_info "Variables générées dans .env.$env_name"
}

# Déploiement d'environnement
deploy_environment() {
    local env_name="$1"
    
    log_info "Déploiement de l'environnement $env_name"
    
    # Validation
    if ! validate_environment "$env_name"; then
        return 1
    fi
    
    # Génération des variables
    generate_env_vars "$env_name"
    
    # Application de la configuration
    kubectl apply -f "k8s/$env_name/"
    
    # Vérification du déploiement
    kubectl rollout status deployment/veza-backend-api -n veza
    kubectl rollout status deployment/veza-chat-server -n veza
    kubectl rollout status deployment/veza-stream-server -n veza
    
    log_info "Environnement $env_name déployé avec succès"
}

# Sauvegarde d'environnement
backup_environment() {
    local env_name="$1"
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    
    log_info "Sauvegarde de l'environnement $env_name"
    
    mkdir -p "$backup_dir"
    
    # Sauvegarde de la configuration
    cp "$ENV_DIR/$env_name.yaml" "$backup_dir/"
    
    # Sauvegarde des secrets
    if [ -f "$SECRETS_DIR/$env_name.env" ]; then
        cp "$SECRETS_DIR/$env_name.env" "$backup_dir/"
    fi
    
    # Sauvegarde de la base de données
    if [ "$env_name" != "development" ]; then
        pg_dump -h "$(yq eval '.data.database.host' "$ENV_DIR/$env_name.yaml")" \
                -U "$(yq eval '.data.database.user' "$ENV_DIR/$env_name.yaml")" \
                -d "$(yq eval '.data.database.name' "$ENV_DIR/$env_name.yaml")" \
                > "$backup_dir/database.sql"
    fi
    
    log_info "Sauvegarde terminée: $backup_dir"
}

# Restauration d'environnement
restore_environment() {
    local env_name="$1"
    local backup_dir="$2"
    
    log_info "Restauration de l'environnement $env_name depuis $backup_dir"
    
    if [ ! -d "$backup_dir" ]; then
        log_error "Répertoire de sauvegarde non trouvé: $backup_dir"
        return 1
    fi
    
    # Restauration de la configuration
    if [ -f "$backup_dir/$env_name.yaml" ]; then
        cp "$backup_dir/$env_name.yaml" "$ENV_DIR/"
    fi
    
    # Restauration des secrets
    if [ -f "$backup_dir/$env_name.env" ]; then
        cp "$backup_dir/$env_name.env" "$SECRETS_DIR/"
    fi
    
    # Restauration de la base de données
    if [ -f "$backup_dir/database.sql" ]; then
        psql -h "$(yq eval '.data.database.host' "$ENV_DIR/$env_name.yaml")" \
             -U "$(yq eval '.data.database.user' "$ENV_DIR/$env_name.yaml")" \
             -d "$(yq eval '.data.database.name' "$ENV_DIR/$env_name.yaml")" \
             < "$backup_dir/database.sql"
    fi
    
    log_info "Restauration terminée"
}

# Menu principal
case "${1:-}" in
    deploy)
        deploy_environment "${2:-}"
        ;;
    backup)
        backup_environment "${2:-}"
        ;;
    restore)
        restore_environment "${2:-}" "${3:-}"
        ;;
    validate)
        validate_environment "${2:-}"
        ;;
    *)
        echo "Usage: $0 {deploy|backup|restore|validate} [environment] [backup_dir]"
        exit 1
        ;;
esac
```

## Déploiement

### 1. Scripts de Déploiement

```yaml
# environment-management/deployment/deployment-pipeline.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-deployment-pipeline
  namespace: veza
data:
  # Pipeline de déploiement
  pipeline:
    stages:
      - name: "validation"
        steps:
          - "validate_configuration"
          - "run_tests"
          - "security_scan"
      
      - name: "build"
        steps:
          - "build_images"
          - "push_images"
          - "update_manifests"
      
      - name: "deploy"
        steps:
          - "deploy_to_staging"
          - "run_integration_tests"
          - "deploy_to_production"
      
      - name: "verification"
        steps:
          - "health_checks"
          - "performance_tests"
          - "monitoring_verification"
  
  # Configuration des environnements
  environments:
    development:
      auto_deploy: true
      manual_approval: false
      rollback_enabled: true
    
    staging:
      auto_deploy: true
      manual_approval: false
      rollback_enabled: true
    
    production:
      auto_deploy: false
      manual_approval: true
      rollback_enabled: true
      blue_green: true
```

### 2. Configuration Kubernetes

```yaml
# environment-management/k8s/production/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: veza-production
  labels:
    name: veza-production
    environment: production
---
# environment-management/k8s/production/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-production-config
  namespace: veza-production
data:
  # Configuration de l'application
  app_config:
    environment: "production"
    log_level: "warn"
    metrics_enabled: "true"
    tracing_enabled: "true"
  
  # Configuration des services
  services:
    backend_api_url: "https://api.veza.com"
    chat_server_url: "wss://chat.veza.com"
    stream_server_url: "https://stream.veza.com"
  
  # Configuration de la base de données
  database:
    host: "prod-db.veza.com"
    port: "5432"
    name: "veza_production"
    ssl_mode: "require"
    max_connections: "100"
    connection_timeout: "30s"
  
  # Configuration du cache
  cache:
    host: "prod-redis.veza.com"
    port: "6379"
    cluster_mode: "true"
    max_connections: "50"
    connection_timeout: "5s"
---
# environment-management/k8s/production/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: veza-production-secrets
  namespace: veza-production
type: Opaque
data:
  # Secrets encodés en base64
  db_password: <base64-encoded-password>
  redis_password: <base64-encoded-password>
  jwt_secret: <base64-encoded-secret>
  api_key: <base64-encoded-api-key>
```

## Bonnes Pratiques

### 1. Règles de Gestion d'Environnement

```yaml
# environment-management/best-practices/environment-rules.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-environment-rules
  namespace: veza
data:
  # Règles générales
  general_rules:
    - "Séparer les environnements par domaine"
    - "Utiliser des configurations spécifiques"
    - "Gérer les secrets de manière sécurisée"
    - "Documenter les changements"
    - "Tester avant déploiement"
  
  # Règles de sécurité
  security_rules:
    - "Chiffrer les données sensibles"
    - "Utiliser des secrets Kubernetes"
    - "Limiter l'accès aux environnements"
    - "Auditer les accès"
    - "Rotater les clés régulièrement"
  
  # Règles de performance
  performance_rules:
    - "Optimiser les ressources par environnement"
    - "Configurer l'auto-scaling"
    - "Monitorer les performances"
    - "Optimiser les requêtes base de données"
    - "Utiliser le caching approprié"
  
  # Règles de déploiement
  deployment_rules:
    - "Utiliser des pipelines CI/CD"
    - "Tester en staging avant production"
    - "Configurer le rollback automatique"
    - "Monitorer les déploiements"
    - "Documenter les procédures"
```

### 2. Monitoring d'Environnement

```go
// environment-management/monitoring/environment-monitor.go
package monitoring

import (
    "context"
    "time"
)

// Moniteur d'environnement
type EnvironmentMonitor struct {
    environment string
    metrics     map[string]float64
}

func NewEnvironmentMonitor(environment string) *EnvironmentMonitor {
    return &EnvironmentMonitor{
        environment: environment,
        metrics:     make(map[string]float64),
    }
}

func (em *EnvironmentMonitor) MonitorHealth(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            em.checkHealth()
        }
    }
}

func (em *EnvironmentMonitor) checkHealth() {
    // Vérification des services
    services := []string{"backend-api", "chat-server", "stream-server"}
    
    for _, service := range services {
        if em.isServiceHealthy(service) {
            em.metrics[service+"_health"] = 1
        } else {
            em.metrics[service+"_health"] = 0
        }
    }
    
    // Vérification de la base de données
    if em.isDatabaseHealthy() {
        em.metrics["database_health"] = 1
    } else {
        em.metrics["database_health"] = 0
    }
    
    // Vérification du cache
    if em.isCacheHealthy() {
        em.metrics["cache_health"] = 1
    } else {
        em.metrics["cache_health"] = 0
    }
}

func (em *EnvironmentMonitor) isServiceHealthy(service string) bool {
    // Logique de vérification de santé
    return true
}

func (em *EnvironmentMonitor) isDatabaseHealthy() bool {
    // Logique de vérification de la base de données
    return true
}

func (em *EnvironmentMonitor) isCacheHealthy() bool {
    // Logique de vérification du cache
    return true
}
```

## Pièges à Éviter

### 1. Configuration en Dur

❌ **Mauvais** :
```go
// Configuration en dur
const (
    DBHost = "localhost"
    DBPort = 5432
    DBName = "veza"
)
```

✅ **Bon** :
```go
// Configuration par environnement
type Config struct {
    Database DatabaseConfig `yaml:"database"`
    Cache    CacheConfig    `yaml:"cache"`
    Logging  LoggingConfig  `yaml:"logging"`
}

func LoadConfig(environment string) (*Config, error) {
    // Chargement depuis un fichier de configuration
    configFile := fmt.Sprintf("config/%s.yaml", environment)
    // ...
}
```

### 2. Secrets en Clair

❌ **Mauvais** :
```yaml
# Secrets en clair
database:
  password: "my_password"
```

✅ **Bon** :
```yaml
# Secrets gérés par Kubernetes
database:
  password: "${DB_PASSWORD}"
```

### 3. Pas de Monitoring

❌ **Mauvais** :
```yaml
# Pas de monitoring configuré
monitoring:
  enabled: false
```

✅ **Bon** :
```yaml
# Monitoring complet
monitoring:
  enabled: true
  prometheus: true
  grafana: true
  alerting: true
  tracing: true
```

## Ressources

### Documentation Interne

- [Guide de Déploiement](../deployment/README.md)
- [Guide de Monitoring](../monitoring/README.md)
- [Guide de Sécurité](../security/README.md)

### Outils Recommandés

- **Kubernetes** : Orchestration
- **Helm** : Gestion des packages
- **ArgoCD** : GitOps
- **Prometheus** : Monitoring

### Commandes Utiles

```bash
# Gestion d'environnement
./env-manager.sh deploy production
./env-manager.sh backup staging
./env-manager.sh restore production backup_dir

# Kubernetes
kubectl get pods -n veza-production
kubectl logs deployment/backend-api -n veza-production
kubectl exec -it pod/backend-api -n veza-production -- /bin/bash

# Monitoring
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
kubectl port-forward svc/grafana 3000:3000 -n monitoring
```

---

**Dernière mise à jour** : $(date)
**Version du guide** : 1.0.0
**Mainteneur** : Équipe DevOps Veza 