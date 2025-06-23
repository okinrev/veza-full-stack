# Health Module Documentation

Le module health fournit un système complet de monitoring de santé, de diagnostics et d'alertes pour le serveur de streaming audio.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Architecture de monitoring](#architecture-de-monitoring)
- [Types de checks](#types-de-checks)
- [Métriques de performance](#métriques-de-performance)
- [Système d'alertes](#système-dalertes)
- [Types et Structures](#types-et-structures)
- [HealthMonitor](#healthmonitor)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le système de monitoring de santé comprend :
- **Checks de santé** : Vérifications système automatiques et périodiques
- **Métriques de performance** : CPU, mémoire, disque, réseau
- **Monitoring des dépendances** : Base de données, Redis, services externes
- **Système d'alertes** : Notifications automatiques en cas de problème
- **Diagnostics avancés** : Analyse de tendances et prédiction de pannes
- **APIs de santé** : Endpoints pour monitoring externe

## Architecture de monitoring

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   System        │    │   Health         │    │   Alerting      │
│   Resources     │    │   Monitor        │    │   System        │
│                 │───►│                  │───►│                 │
│ - CPU Usage     │    │ - Health Checks  │    │ - Notifications │
│ - Memory        │    │ - Performance    │    │ - Escalation    │
│ - Disk Space    │    │ - Dependencies   │    │ - Recovery      │
│ - Network       │    │ - Trending       │    │ - Reporting     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                    ┌──────────────────────┐
                    │   External           │
                    │   Dependencies       │
                    │                      │
                    │ - Database           │
                    │ - Redis Cache        │
                    │ - Jaeger Tracing     │
                    │ - File System        │
                    │ - WebSocket Conns    │
                    └──────────────────────┘
```

## Types de checks

### Check Categories

Le système effectue plusieurs types de vérifications :

#### 1. Checks système (System Health)
- **CPU Usage** : Utilisation processeur
- **Memory Usage** : Consommation mémoire
- **Disk Space** : Espace disque disponible
- **Network Connections** : Connexions réseau actives

#### 2. Checks de services (Service Health)
- **Database Connectivity** : Connexion à SQLite
- **Cache Health** : État du cache fichiers
- **WebSocket Health** : Connexions WebSocket actives
- **Audio Directory** : Accessibilité des fichiers audio

#### 3. Checks externes (External Dependencies)
- **Redis Connectivity** : Connexion Redis (si activé)
- **Jaeger Connectivity** : Service de tracing (si activé)
- **External APIs** : Services tiers intégrés

### ServiceStatus

```rust
pub enum ServiceStatus {
    Healthy,    // Tout va bien
    Degraded,   // Performance dégradée mais fonctionnel
    Unhealthy,  // Problème nécessitant attention
    Critical,   // Panne critique nécessitant intervention immédiate
}
```

### CheckStatus

```rust
pub enum CheckStatus {
    Pass,   // Check réussi
    Warn,   // Avertissement (seuil dépassé)
    Fail,   // Échec du check
}
```

## Métriques de performance

### PerformanceMetrics

```rust
pub struct PerformanceMetrics {
    pub cpu_usage_percent: f32,             // Utilisation CPU (%)
    pub memory_usage_mb: u64,               // Mémoire utilisée (MB)
    pub memory_usage_percent: f32,          // Utilisation mémoire (%)
    pub disk_usage_percent: f32,            // Utilisation disque (%)
    pub network_connections: u32,           // Connexions réseau actives
    pub response_times: ResponseTimeMetrics, // Temps de réponse
    pub error_rates: ErrorRateMetrics,      // Taux d'erreur
}
```

### ResponseTimeMetrics

```rust
pub struct ResponseTimeMetrics {
    pub p50_ms: f64,        // Percentile 50 (médiane)
    pub p95_ms: f64,        // Percentile 95
    pub p99_ms: f64,        // Percentile 99
    pub average_ms: f64,    // Temps moyen
}
```

### ErrorRateMetrics

```rust
pub struct ErrorRateMetrics {
    pub rate_1min: f32,         // Taux d'erreur dernière minute
    pub rate_5min: f32,         // Taux d'erreur 5 dernières minutes
    pub rate_15min: f32,        // Taux d'erreur 15 dernières minutes
    pub total_errors_24h: u64,  // Total erreurs 24h
}
```

## Système d'alertes

### HealthAlert

```rust
pub struct HealthAlert {
    pub id: String,             // ID unique de l'alerte
    pub severity: AlertSeverity, // Niveau de gravité
    pub message: String,        // Message d'alerte
    pub component: String,      // Composant concerné
    pub timestamp: u64,         // Timestamp de création
    pub resolved: bool,         // Alerte résolue
}
```

### AlertSeverity

```rust
pub enum AlertSeverity {
    Info,       // Information
    Warning,    // Avertissement
    Critical,   // Critique
    Emergency,  // Urgence
}
```

### Règles d'alerte

Le système génère automatiquement des alertes selon des seuils :

| Métrique | Warning | Critical | Emergency |
|----------|---------|----------|-----------|
| CPU Usage | > 70% | > 85% | > 95% |
| Memory Usage | > 80% | > 90% | > 95% |
| Disk Usage | > 85% | > 95% | > 98% |
| Error Rate | > 5% | > 10% | > 25% |
| Response Time P95 | > 1s | > 3s | > 10s |

## Types et Structures

### HealthStatus (Structure principale)

```rust
pub struct HealthStatus {
    pub status: ServiceStatus,                      // État global du service
    pub timestamp: u64,                            // Timestamp du check
    pub service: String,                           // Nom du service
    pub version: String,                           // Version du service
    pub uptime_seconds: u64,                       // Durée de fonctionnement
    pub checks: HashMap<String, HealthCheck>,      // Détail des checks
    pub alerts: Vec<HealthAlert>,                  // Alertes actives
    pub performance: PerformanceMetrics,           // Métriques de performance
}
```

### HealthCheck

```rust
pub struct HealthCheck {
    pub name: String,                   // Nom du check
    pub status: CheckStatus,            // Statut du check
    pub message: String,                // Message descriptif
    pub duration_ms: u64,               // Durée d'exécution du check
    pub last_success: Option<u64>,      // Dernier succès
    pub last_failure: Option<u64>,      // Dernier échec
    pub failure_count: u32,             // Nombre d'échecs consécutifs
    pub threshold: HealthThreshold,     // Seuils configurés
}
```

### HealthThreshold

```rust
pub struct HealthThreshold {
    pub max_response_time_ms: u64,      // Temps de réponse max
    pub max_failure_rate: f32,          // Taux d'échec max
    pub max_consecutive_failures: u32,   // Échecs consécutifs max
}
```

## HealthMonitor

### Structure principale

```rust
pub struct HealthMonitor {
    config: Arc<Config>,                                    // Configuration
    start_time: SystemTime,                                 // Heure de démarrage
    checks: Arc<RwLock<HashMap<String, HealthCheck>>>,     // Checks actifs
    alerts: Arc<RwLock<Vec<HealthAlert>>>,                 // Alertes actives
    performance_history: Arc<RwLock<Vec<PerformanceMetrics>>>, // Historique perf
}
```

### Cycle d'exécution

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Check          │───►│  Analysis        │───►│  Action         │
│  Execution      │    │                  │    │                 │
│                 │    │ - Thresholds     │    │ - Alerting      │
│ - System Checks │    │ - Trending       │    │ - Logging       │
│ - Service Checks│    │ - Pattern        │    │ - Recovery      │
│ - Dependency    │    │   Detection      │    │ - Notification  │
│   Checks        │    │ - Health Score   │    │ - Reporting     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## API Reference

### HealthMonitor Methods

#### `new(config: Arc<Config>) -> Self`
Crée une nouvelle instance du monitor de santé.

#### `start_monitoring(&self)`
Démarre le monitoring automatique en arrière-plan.

#### `get_health_status(&self) -> HealthStatus`
Récupère l'état de santé actuel complet.

#### `run_health_checks(&self)`
Exécute tous les checks de santé manuellement.

#### `analyze_and_alert(&self)`
Analyse les métriques et génère des alertes si nécessaire.

### Checks implémentés

#### `check_system_resources(&self)`
Vérifie l'utilisation des ressources système (CPU, mémoire, disque).

#### `check_database_connectivity(&self)`
Teste la connectivité à la base de données SQLite.

#### `check_audio_directory(&self)`
Vérifie l'accessibilité du répertoire audio et les permissions.

#### `check_cache_health(&self)`
Contrôle la santé du système de cache.

#### `check_websocket_health(&self)`
Monitore les connexions WebSocket actives.

#### `check_redis_connectivity(&self)`
Teste la connectivité Redis (si configuré).

#### `check_jaeger_connectivity(&self, endpoint: &str)`
Vérifie la connectivité au service Jaeger (si configuré).

### Endpoints API

#### `GET /health`
Retourne un check de santé basique.

```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "uptime_seconds": 3600
}
```

#### `GET /health/detailed`
Retourne un rapport de santé détaillé.

```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "service": "stream-server",
  "version": "0.2.0",
  "uptime_seconds": 3600,
  "checks": {
    "database": {
      "status": "pass",
      "message": "Database connection healthy",
      "duration_ms": 5
    },
    "disk_space": {
      "status": "warn",
      "message": "Disk usage at 87%",
      "duration_ms": 2
    }
  },
  "performance": {
    "cpu_usage_percent": 45.2,
    "memory_usage_mb": 256,
    "memory_usage_percent": 32.1
  },
  "alerts": []
}
```

## Exemples d'utilisation

### Monitoring de base

```rust
use stream_server::health::{HealthMonitor, ServiceStatus};

async fn example_basic_monitoring() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let health_monitor = HealthMonitor::new(config);
    
    // Démarrer le monitoring automatique
    health_monitor.start_monitoring().await;
    
    // Attendre un peu pour collecter des données
    tokio::time::sleep(Duration::from_secs(30)).await;
    
    // Récupérer l'état de santé
    let health_status = health_monitor.get_health_status().await;
    
    println!("🏥 État de santé du service:");
    println!("  Status: {:?}", health_status.status);
    println!("  Service: {}", health_status.service);
    println!("  Version: {}", health_status.version);
    println!("  Uptime: {}s", health_status.uptime_seconds);
    
    // Afficher les métriques de performance
    let perf = &health_status.performance;
    println!("\n📊 Métriques de performance:");
    println!("  CPU: {:.1}%", perf.cpu_usage_percent);
    println!("  Mémoire: {} MB ({:.1}%)", perf.memory_usage_mb, perf.memory_usage_percent);
    println!("  Disque: {:.1}%", perf.disk_usage_percent);
    println!("  Connexions réseau: {}", perf.network_connections);
    
    // Temps de réponse
    println!("\n⏱️  Temps de réponse:");
    println!("  P50: {:.1}ms", perf.response_times.p50_ms);
    println!("  P95: {:.1}ms", perf.response_times.p95_ms);
    println!("  P99: {:.1}ms", perf.response_times.p99_ms);
    println!("  Moyenne: {:.1}ms", perf.response_times.average_ms);
    
    // Taux d'erreur
    println!("\n❌ Taux d'erreur:");
    println!("  1 min: {:.2}%", perf.error_rates.rate_1min);
    println!("  5 min: {:.2}%", perf.error_rates.rate_5min);
    println!("  15 min: {:.2}%", perf.error_rates.rate_15min);
    
    Ok(())
}
```

### Checks détaillés

```rust
async fn example_detailed_checks() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let health_monitor = HealthMonitor::new(config);
    
    // Exécuter tous les checks manuellement
    health_monitor.run_health_checks().await;
    
    let health_status = health_monitor.get_health_status().await;
    
    println!("🔍 Détail des checks de santé:");
    
    for (check_name, check) in &health_status.checks {
        let status_icon = match check.status {
            CheckStatus::Pass => "✅",
            CheckStatus::Warn => "⚠️",
            CheckStatus::Fail => "❌",
        };
        
        println!("\n{} {} ({:?})", status_icon, check_name, check.status);
        println!("  Message: {}", check.message);
        println!("  Durée: {}ms", check.duration_ms);
        
        if let Some(last_success) = check.last_success {
            let success_time = SystemTime::UNIX_EPOCH + Duration::from_secs(last_success);
            println!("  Dernier succès: {:?}", success_time);
        }
        
        if let Some(last_failure) = check.last_failure {
            let failure_time = SystemTime::UNIX_EPOCH + Duration::from_secs(last_failure);
            println!("  Dernier échec: {:?}", failure_time);
        }
        
        if check.failure_count > 0 {
            println!("  Échecs consécutifs: {}", check.failure_count);
        }
        
        // Seuils configurés
        println!("  Seuils:");
        println!("    Temps max: {}ms", check.threshold.max_response_time_ms);
        println!("    Taux échec max: {:.1}%", check.threshold.max_failure_rate * 100.0);
        println!("    Échecs consécutifs max: {}", check.threshold.max_consecutive_failures);
    }
    
    Ok(())
}
```

### Gestion des alertes

```rust
async fn example_alert_management() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let health_monitor = HealthMonitor::new(config);
    
    // Démarrer le monitoring
    health_monitor.start_monitoring().await;
    
    // Simuler une charge élevée pour déclencher des alertes
    println!("🔥 Simulation de charge élevée...");
    
    // Boucle de monitoring avec gestion d'alertes
    for i in 0..10 {
        tokio::time::sleep(Duration::from_secs(30)).await;
        
        let health_status = health_monitor.get_health_status().await;
        
        println!("\n📋 Check #{} - Status: {:?}", i + 1, health_status.status);
        
        // Vérifier les nouvelles alertes
        if !health_status.alerts.is_empty() {
            println!("🚨 Alertes actives:");
            
            for alert in &health_status.alerts {
                let severity_icon = match alert.severity {
                    AlertSeverity::Info => "ℹ️",
                    AlertSeverity::Warning => "⚠️",
                    AlertSeverity::Critical => "🔴",
                    AlertSeverity::Emergency => "🚨",
                };
                
                let status_icon = if alert.resolved { "✅" } else { "❌" };
                
                println!("  {} {} [{}] {}: {}", 
                    severity_icon, 
                    status_icon,
                    alert.component,
                    alert.id,
                    alert.message
                );
                
                let alert_time = SystemTime::UNIX_EPOCH + Duration::from_secs(alert.timestamp);
                println!("    Créée: {:?}", alert_time);
            }
        } else {
            println!("✅ Aucune alerte active");
        }
        
        // Actions basées sur le statut
        match health_status.status {
            ServiceStatus::Critical | ServiceStatus::Unhealthy => {
                println!("🚨 État critique détecté - Actions correctives recommandées!");
            }
            ServiceStatus::Degraded => {
                println!("⚠️ Performance dégradée - Surveillance accrue");
            }
            ServiceStatus::Healthy => {
                println!("✅ Service en bonne santé");
            }
        }
    }
    
    Ok(())
}
```

### Monitoring de performance en continu

```rust
async fn example_performance_monitoring() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let health_monitor = HealthMonitor::new(config);
    
    health_monitor.start_monitoring().await;
    
    println!("📊 Monitoring de performance en temps réel");
    println!("Appuyez sur Ctrl+C pour arrêter\n");
    
    let mut previous_metrics: Option<PerformanceMetrics> = None;
    
    loop {
        let health_status = health_monitor.get_health_status().await;
        let current_metrics = &health_status.performance;
        
        // Afficher les métriques actuelles
        print!("\r🔄 CPU: {:5.1}% | RAM: {:6} MB ({:5.1}%) | Disk: {:5.1}% | Conn: {:4} | Errors: {:5.2}%",
            current_metrics.cpu_usage_percent,
            current_metrics.memory_usage_mb,
            current_metrics.memory_usage_percent,
            current_metrics.disk_usage_percent,
            current_metrics.network_connections,
            current_metrics.error_rates.rate_1min
        );
        
        // Calculer les tendances si nous avons des métriques précédentes
        if let Some(ref prev) = previous_metrics {
            let cpu_trend = current_metrics.cpu_usage_percent - prev.cpu_usage_percent;
            let memory_trend = current_metrics.memory_usage_percent - prev.memory_usage_percent;
            
            if cpu_trend.abs() > 5.0 || memory_trend.abs() > 5.0 {
                println!("\n📈 Tendances:");
                if cpu_trend.abs() > 5.0 {
                    let trend_icon = if cpu_trend > 0.0 { "📈" } else { "📉" };
                    println!("  {} CPU: {:+.1}%", trend_icon, cpu_trend);
                }
                if memory_trend.abs() > 5.0 {
                    let trend_icon = if memory_trend > 0.0 { "📈" } else { "📉" };
                    println!("  {} Mémoire: {:+.1}%", trend_icon, memory_trend);
                }
            }
        }
        
        // Détecter les anomalies
        if current_metrics.cpu_usage_percent > 80.0 {
            println!("\n🔥 ATTENTION: CPU usage élevé ({:.1}%)", current_metrics.cpu_usage_percent);
        }
        
        if current_metrics.memory_usage_percent > 85.0 {
            println!("\n💾 ATTENTION: Mémoire élevée ({:.1}%)", current_metrics.memory_usage_percent);
        }
        
        if current_metrics.error_rates.rate_1min > 5.0 {
            println!("\n❌ ATTENTION: Taux d'erreur élevé ({:.2}%)", current_metrics.error_rates.rate_1min);
        }
        
        previous_metrics = Some(current_metrics.clone());
        
        tokio::time::sleep(Duration::from_secs(5)).await;
    }
}
```

### Check personnalisé

```rust
async fn example_custom_health_check() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let health_monitor = HealthMonitor::new(config);
    
    // Ajouter un check personnalisé pour l'API externe
    async fn check_external_api() -> HealthCheck {
        let start_time = SystemTime::now();
        
        // Tenter de contacter une API externe
        let result = reqwest::Client::new()
            .get("https://api.external-service.com/health")
            .timeout(Duration::from_secs(5))
            .send()
            .await;
        
        let duration = start_time.elapsed().unwrap_or_default();
        
        match result {
            Ok(response) if response.status().is_success() => {
                HealthCheck {
                    name: "external_api".to_string(),
                    status: CheckStatus::Pass,
                    message: format!("API externe accessible ({})", response.status()),
                    duration_ms: duration.as_millis() as u64,
                    last_success: Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()),
                    last_failure: None,
                    failure_count: 0,
                    threshold: HealthThreshold {
                        max_response_time_ms: 5000,
                        max_failure_rate: 0.1,
                        max_consecutive_failures: 3,
                    },
                }
            }
            Ok(response) => {
                HealthCheck {
                    name: "external_api".to_string(),
                    status: CheckStatus::Warn,
                    message: format!("API externe retourne {}", response.status()),
                    duration_ms: duration.as_millis() as u64,
                    last_success: None,
                    last_failure: Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()),
                    failure_count: 1,
                    threshold: HealthThreshold {
                        max_response_time_ms: 5000,
                        max_failure_rate: 0.1,
                        max_consecutive_failures: 3,
                    },
                }
            }
            Err(e) => {
                HealthCheck {
                    name: "external_api".to_string(),
                    status: CheckStatus::Fail,
                    message: format!("Erreur API externe: {}", e),
                    duration_ms: duration.as_millis() as u64,
                    last_success: None,
                    last_failure: Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()),
                    failure_count: 1,
                    threshold: HealthThreshold {
                        max_response_time_ms: 5000,
                        max_failure_rate: 0.1,
                        max_consecutive_failures: 3,
                    },
                }
            }
        }
    }
    
    // Exécuter le check personnalisé
    let custom_check = check_external_api().await;
    println!("🔧 Check personnalisé:");
    println!("  Status: {:?}", custom_check.status);
    println!("  Message: {}", custom_check.message);
    println!("  Durée: {}ms", custom_check.duration_ms);
    
    Ok(())
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    
    // Créer le monitor de santé
    let health_monitor = Arc::new(HealthMonitor::new(config.clone()));
    
    // Démarrer le monitoring automatique
    health_monitor.start_monitoring().await;
    
    // Router avec endpoints de santé
    let app = Router::new()
        // Endpoints de santé
        .route("/health", get(health_check))
        .route("/health/detailed", get(detailed_health_check))
        .route("/health/live", get(liveness_check))
        .route("/health/ready", get(readiness_check))
        
        // Métriques pour Prometheus
        .route("/metrics", get(metrics_endpoint))
        
        .with_state(AppState {
            health_monitor,
            // ... autres composants
        });
    
    // Handlers de santé
    async fn health_check() -> Json<serde_json::Value> {
        Json(json!({
            "status": "healthy",
            "timestamp": SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            "service": "stream-server",
            "version": env!("CARGO_PKG_VERSION")
        }))
    }
    
    async fn detailed_health_check(
        State(state): State<AppState>,
    ) -> Json<HealthStatus> {
        Json(state.health_monitor.get_health_status().await)
    }
    
    // Démarrer le serveur
    let listener = tokio::net::TcpListener::bind("0.0.0.0:8082").await?;
    axum::serve(listener, app).await?;
    
    Ok(())
}
```

### Avec Docker Health Check

```dockerfile
# Dans le Dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1
```

### Avec Kubernetes

```yaml
# kubernetes-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stream-server
spec:
  template:
    spec:
      containers:
      - name: stream-server
        image: stream-server:latest
        ports:
        - containerPort: 8082
        
        # Liveness probe
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        # Readiness probe
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8082
          initialDelaySeconds: 15
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2
        
        # Startup probe
        startupProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30
```

### Avec Prometheus monitoring

```rust
// Métriques Prometheus
use prometheus::{Counter, Gauge, Histogram, Registry};

pub struct HealthMetrics {
    pub check_duration: Histogram,
    pub check_failures: Counter,
    pub cpu_usage: Gauge,
    pub memory_usage: Gauge,
    pub disk_usage: Gauge,
    pub active_connections: Gauge,
}

impl HealthMetrics {
    pub fn new(registry: &Registry) -> Result<Self, prometheus::Error> {
        let check_duration = Histogram::new(
            prometheus::HistogramOpts::new(
                "health_check_duration_seconds",
                "Duration of health checks in seconds"
            ).buckets(vec![0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0])
        )?;
        
        let check_failures = Counter::new(
            "health_check_failures_total",
            "Total number of health check failures"
        )?;
        
        let cpu_usage = Gauge::new(
            "system_cpu_usage_percent",
            "CPU usage percentage"
        )?;
        
        let memory_usage = Gauge::new(
            "system_memory_usage_percent", 
            "Memory usage percentage"
        )?;
        
        let disk_usage = Gauge::new(
            "system_disk_usage_percent",
            "Disk usage percentage"
        )?;
        
        let active_connections = Gauge::new(
            "network_connections_active",
            "Number of active network connections"
        )?;
        
        // Enregistrer les métriques
        registry.register(Box::new(check_duration.clone()))?;
        registry.register(Box::new(check_failures.clone()))?;
        registry.register(Box::new(cpu_usage.clone()))?;
        registry.register(Box::new(memory_usage.clone()))?;
        registry.register(Box::new(disk_usage.clone()))?;
        registry.register(Box::new(active_connections.clone()))?;
        
        Ok(Self {
            check_duration,
            check_failures,
            cpu_usage,
            memory_usage,
            disk_usage,
            active_connections,
        })
    }
    
    pub fn update_from_health_status(&self, health_status: &HealthStatus) {
        // Mettre à jour les métriques Prometheus
        self.cpu_usage.set(health_status.performance.cpu_usage_percent as f64);
        self.memory_usage.set(health_status.performance.memory_usage_percent as f64);
        self.disk_usage.set(health_status.performance.disk_usage_percent as f64);
        self.active_connections.set(health_status.performance.network_connections as f64);
        
        // Compter les échecs de checks
        for check in health_status.checks.values() {
            if matches!(check.status, CheckStatus::Fail) {
                self.check_failures.inc();
            }
        }
    }
}
```

### Avec l'API Go

```go
// Client de monitoring Go
type HealthClient struct {
    baseURL    string
    httpClient *http.Client
}

type HealthStatus struct {
    Status      string                 `json:"status"`
    Timestamp   int64                  `json:"timestamp"`
    Service     string                 `json:"service"`
    Version     string                 `json:"version"`
    UptimeSeconds int64                `json:"uptime_seconds"`
    Checks      map[string]HealthCheck `json:"checks"`
    Performance PerformanceMetrics     `json:"performance"`
    Alerts      []HealthAlert          `json:"alerts"`
}

func (c *HealthClient) GetHealthStatus() (*HealthStatus, error) {
    resp, err := c.httpClient.Get(fmt.Sprintf("%s/health/detailed", c.baseURL))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var health HealthStatus
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        return nil, err
    }
    
    return &health, nil
}

func (c *HealthClient) IsHealthy() bool {
    resp, err := c.httpClient.Get(fmt.Sprintf("%s/health", c.baseURL))
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    
    return resp.StatusCode == http.StatusOK
}

// Monitor périodique
func (c *HealthClient) StartMonitoring(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        health, err := c.GetHealthStatus()
        if err != nil {
            log.Printf("Erreur health check: %v", err)
            continue
        }
        
        // Loguer les alertes critiques
        for _, alert := range health.Alerts {
            if alert.Severity == "Critical" || alert.Severity == "Emergency" {
                log.Printf("ALERT [%s]: %s - %s", alert.Severity, alert.Component, alert.Message)
            }
        }
        
        // Loguer les métriques importantes
        if health.Performance.CPUUsagePercent > 80 {
            log.Printf("WARNING: CPU usage élevé: %.1f%%", health.Performance.CPUUsagePercent)
        }
    }
}
```

Cette documentation complète du module health vous permet d'implémenter un système de monitoring robuste et proactif pour votre serveur de streaming audio. 