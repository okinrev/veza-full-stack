use std::collections::HashMap;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};
use tracing::{info, debug};
use crate::Config;


#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthStatus {
    pub status: ServiceStatus,
    pub timestamp: u64,
    pub service: String,
    pub version: String,
    pub uptime_seconds: u64,
    pub checks: HashMap<String, HealthCheck>,
    pub alerts: Vec<HealthAlert>,
    pub performance: PerformanceMetrics,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum ServiceStatus {
    Healthy,
    Degraded,
    Unhealthy,
    Critical,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthCheck {
    pub name: String,
    pub status: CheckStatus,
    pub message: String,
    pub duration_ms: u64,
    pub last_success: Option<u64>,
    pub last_failure: Option<u64>,
    pub failure_count: u32,
    pub threshold: HealthThreshold,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum CheckStatus {
    Pass,
    Warn,
    Fail,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthThreshold {
    pub max_response_time_ms: u64,
    pub max_failure_rate: f32,
    pub max_consecutive_failures: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthAlert {
    pub id: String,
    pub severity: AlertSeverity,
    pub message: String,
    pub component: String,
    pub timestamp: u64,
    pub resolved: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AlertSeverity {
    Info,
    Warning,
    Critical,
    Emergency,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceMetrics {
    pub cpu_usage_percent: f32,
    pub memory_usage_mb: u64,
    pub memory_usage_percent: f32,
    pub disk_usage_percent: f32,
    pub network_connections: u32,
    pub response_times: ResponseTimeMetrics,
    pub error_rates: ErrorRateMetrics,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ResponseTimeMetrics {
    pub p50_ms: f64,
    pub p95_ms: f64,
    pub p99_ms: f64,
    pub average_ms: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorRateMetrics {
    pub rate_1min: f32,
    pub rate_5min: f32,
    pub rate_15min: f32,
    pub total_errors_24h: u64,
}

pub struct HealthMonitor {
    config: Arc<Config>,
    start_time: SystemTime,
    checks: Arc<RwLock<HashMap<String, HealthCheck>>>,
    alerts: Arc<RwLock<Vec<HealthAlert>>>,
    performance_history: Arc<RwLock<Vec<PerformanceMetrics>>>,
}

impl HealthMonitor {
    pub fn new(config: Arc<Config>) -> Self {
        Self {
            config,
            start_time: SystemTime::now(),
            checks: Arc::new(RwLock::new(HashMap::new())),
            alerts: Arc::new(RwLock::new(Vec::new())),
            performance_history: Arc::new(RwLock::new(Vec::new())),
        }
    }

    pub async fn start_monitoring(&self) {
        info!("🏥 Démarrage du monitoring de santé");
        
        let monitor = self.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(monitor.config.monitoring.health_check_interval);
            
            loop {
                interval.tick().await;
                monitor.run_health_checks().await;
            }
        });
    }

    async fn run_health_checks(&self) {
        debug!("Exécution des checks de santé");

        // Checks système
        self.check_system_resources().await;
        self.check_disk_space().await;
        self.check_database_connectivity().await;
        self.check_external_dependencies().await;
        
        // Checks applicatifs
        self.check_audio_directory().await;
        self.check_cache_health().await;
        self.check_websocket_health().await;
        
        // Analyser les alertes
        self.analyze_and_alert().await;
        
        // Nettoyer l'historique
        self.cleanup_old_data().await;
    }

    async fn check_system_resources(&self) {
        let start = SystemTime::now();
        
        let (cpu_usage, memory_info, network_connections) = tokio::join!(
            self.get_cpu_usage(),
            self.get_memory_info(),
            self.get_network_connections()
        );

        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;
        
        let status = if cpu_usage > 90.0 || memory_info.1 > 90.0 {
            CheckStatus::Fail
        } else if cpu_usage > 70.0 || memory_info.1 > 70.0 {
            CheckStatus::Warn
        } else {
            CheckStatus::Pass
        };

        let message = format!(
            "CPU: {:.1}%, Memory: {} MB ({:.1}%), Connections: {}",
            cpu_usage, memory_info.0, memory_info.1, network_connections
        );

        let check = HealthCheck {
            name: "system_resources".to_string(),
            status: status.clone(),
            message,
            duration_ms: duration,
            last_success: if status == CheckStatus::Pass { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_success("system_resources").await 
            },
            last_failure: if status == CheckStatus::Fail { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_failure("system_resources").await 
            },
            failure_count: if status == CheckStatus::Fail { 
                self.increment_failure_count("system_resources").await 
            } else { 
                0 
            },
            threshold: HealthThreshold {
                max_response_time_ms: 5000,
                max_failure_rate: 0.1,
                max_consecutive_failures: 3,
            },
        };

        self.update_check("system_resources", check).await;
    }

    async fn check_disk_space(&self) {
        let start = SystemTime::now();
        
        let disk_info = self.get_disk_space(&self.config.audio_dir).await;
        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;

        let status = if disk_info.free_percentage < 5.0 {
            CheckStatus::Fail
        } else if disk_info.free_percentage < 15.0 {
            CheckStatus::Warn
        } else {
            CheckStatus::Pass
        };

        let message = format!(
            "Espace libre: {:.1}% ({} GB disponible)",
            disk_info.free_percentage,
            disk_info.free_bytes / (1024 * 1024 * 1024)
        );

        let check = HealthCheck {
            name: "disk_space".to_string(),
            status: status.clone(),
            message,
            duration_ms: duration,
            last_success: if status == CheckStatus::Pass { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_success("disk_space").await 
            },
            last_failure: if status == CheckStatus::Fail { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_failure("disk_space").await 
            },
            failure_count: if status == CheckStatus::Fail { 
                self.increment_failure_count("disk_space").await 
            } else { 
                0 
            },
            threshold: HealthThreshold {
                max_response_time_ms: 2000,
                max_failure_rate: 0.05,
                max_consecutive_failures: 2,
            },
        };

        self.update_check("disk_space", check).await;
    }

    async fn check_database_connectivity(&self) {
        let start = SystemTime::now();
        
        // Simuler un check de base de données (à adapter selon votre implémentation)
        let db_healthy = self.test_database_connection().await;
        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;

        let status = if db_healthy {
            CheckStatus::Pass
        } else {
            CheckStatus::Fail
        };

        let message = if db_healthy {
            "Base de données accessible".to_string()
        } else {
            "Impossible de se connecter à la base de données".to_string()
        };

        let check = HealthCheck {
            name: "database".to_string(),
            status: status.clone(),
            message,
            duration_ms: duration,
            last_success: if status == CheckStatus::Pass { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_success("database").await 
            },
            last_failure: if status == CheckStatus::Fail { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_failure("database").await 
            },
            failure_count: if status == CheckStatus::Fail { 
                self.increment_failure_count("database").await 
            } else { 
                0 
            },
            threshold: HealthThreshold {
                max_response_time_ms: 1000,
                max_failure_rate: 0.01,
                max_consecutive_failures: 1,
            },
        };

        self.update_check("database", check).await;
    }

    async fn check_external_dependencies(&self) {
        // Check Redis si activé
        if self.config.redis_enabled() {
            self.check_redis_connectivity().await;
        }

        // Check des services externes si configurés
        if let Some(jaeger_endpoint) = &self.config.monitoring.jaeger_endpoint {
            self.check_jaeger_connectivity(jaeger_endpoint).await;
        }
    }

    async fn check_redis_connectivity(&self) {
        let start = SystemTime::now();
        
        // Simuler un check Redis (à implémenter selon votre client Redis)
        let redis_healthy = self.test_redis_connection().await;
        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;

        let status = if redis_healthy {
            CheckStatus::Pass
        } else {
            CheckStatus::Warn // Redis est optionnel, donc warn au lieu de fail
        };

        let message = if redis_healthy {
            "Redis accessible".to_string()
        } else {
            "Redis non accessible - fonctionnement en mode dégradé".to_string()
        };

        let check = HealthCheck {
            name: "redis".to_string(),
            status,
            message,
            duration_ms: duration,
            last_success: None,
            last_failure: None,
            failure_count: 0,
            threshold: HealthThreshold {
                max_response_time_ms: 500,
                max_failure_rate: 0.2,
                max_consecutive_failures: 5,
            },
        };

        self.update_check("redis", check).await;
    }

    async fn check_jaeger_connectivity(&self, endpoint: &str) {
        let start = SystemTime::now();
        
        let jaeger_healthy = self.test_jaeger_connection(endpoint).await;
        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;

        let status = if jaeger_healthy {
            CheckStatus::Pass
        } else {
            CheckStatus::Warn
        };

        let message = if jaeger_healthy {
            "Jaeger accessible".to_string()
        } else {
            "Jaeger non accessible - tracing désactivé".to_string()
        };

        let check = HealthCheck {
            name: "jaeger".to_string(),
            status,
            message,
            duration_ms: duration,
            last_success: None,
            last_failure: None,
            failure_count: 0,
            threshold: HealthThreshold {
                max_response_time_ms: 2000,
                max_failure_rate: 0.3,
                max_consecutive_failures: 10,
            },
        };

        self.update_check("jaeger", check).await;
    }

    async fn check_audio_directory(&self) {
        let start = SystemTime::now();
        
        let audio_dir_status = self.check_audio_directory_access().await;
        let duration = start.elapsed().unwrap_or_default().as_millis() as u64;

        let (status, message, file_count) = audio_dir_status;

        let check = HealthCheck {
            name: "audio_directory".to_string(),
            status: status.clone(),
            message: format!("{} ({} fichiers)", message, file_count),
            duration_ms: duration,
            last_success: if status == CheckStatus::Pass { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_success("audio_directory").await 
            },
            last_failure: if status == CheckStatus::Fail { 
                Some(SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()) 
            } else { 
                self.get_last_failure("audio_directory").await 
            },
            failure_count: if status == CheckStatus::Fail { 
                self.increment_failure_count("audio_directory").await 
            } else { 
                0 
            },
            threshold: HealthThreshold {
                max_response_time_ms: 3000,
                max_failure_rate: 0.01,
                max_consecutive_failures: 1,
            },
        };

        self.update_check("audio_directory", check).await;
    }

    async fn check_cache_health(&self) {
        // Simuler un check de cache (à adapter selon votre implémentation)
        let cache_stats = self.get_cache_statistics().await;
        
        let status = if cache_stats.hit_rate > 0.5 {
            CheckStatus::Pass
        } else if cache_stats.hit_rate > 0.2 {
            CheckStatus::Warn
        } else {
            CheckStatus::Fail
        };

        let message = format!(
            "Hit rate: {:.1}%, {} entrées, {} MB utilisés",
            cache_stats.hit_rate * 100.0,
            cache_stats.entries,
            cache_stats.memory_usage_mb
        );

        let check = HealthCheck {
            name: "cache".to_string(),
            status,
            message,
            duration_ms: 0,
            last_success: None,
            last_failure: None,
            failure_count: 0,
            threshold: HealthThreshold {
                max_response_time_ms: 100,
                max_failure_rate: 0.1,
                max_consecutive_failures: 5,
            },
        };

        self.update_check("cache", check).await;
    }

    async fn check_websocket_health(&self) {
        // Simuler un check WebSocket (à adapter selon votre implémentation)
        let ws_stats = self.get_websocket_statistics().await;
        
        let status = if ws_stats.active_connections < 10000 && ws_stats.error_rate < 0.05 {
            CheckStatus::Pass
        } else if ws_stats.active_connections < 15000 && ws_stats.error_rate < 0.1 {
            CheckStatus::Warn
        } else {
            CheckStatus::Fail
        };

        let message = format!(
            "{} connexions actives, {:.1}% erreurs",
            ws_stats.active_connections,
            ws_stats.error_rate * 100.0
        );

        let check = HealthCheck {
            name: "websockets".to_string(),
            status,
            message,
            duration_ms: 0,
            last_success: None,
            last_failure: None,
            failure_count: 0,
            threshold: HealthThreshold {
                max_response_time_ms: 200,
                max_failure_rate: 0.05,
                max_consecutive_failures: 3,
            },
        };

        self.update_check("websockets", check).await;
    }

    async fn analyze_and_alert(&self) {
        let checks = self.checks.read().await;
        let mut new_alerts = Vec::new();

        for (name, check) in checks.iter() {
            match check.status {
                CheckStatus::Fail => {
                    if check.failure_count >= check.threshold.max_consecutive_failures {
                        let alert = HealthAlert {
                            id: format!("{}_{}", name, SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()),
                            severity: AlertSeverity::Critical,
                            message: format!("Service {} en échec critique: {}", name, check.message),
                            component: name.clone(),
                            timestamp: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
                            resolved: false,
                        };
                        new_alerts.push(alert);
                    }
                }
                CheckStatus::Warn => {
                    let alert = HealthAlert {
                        id: format!("{}_{}", name, SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs()),
                        severity: AlertSeverity::Warning,
                        message: format!("Service {} dégradé: {}", name, check.message),
                        component: name.clone(),
                        timestamp: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
                        resolved: false,
                    };
                    new_alerts.push(alert);
                }
                _ => {}
            }
        }

        if !new_alerts.is_empty() {
            let mut alerts = self.alerts.write().await;
            alerts.extend(new_alerts);
            
            // Garder seulement les 100 dernières alertes
            if alerts.len() > 100 {
                let len = alerts.len();
                alerts.drain(0..len - 100);
            }
        }
    }

    pub async fn get_health_status(&self) -> HealthStatus {
        let checks = self.checks.read().await.clone();
        let alerts = self.alerts.read().await.clone();
        
        let overall_status = self.calculate_overall_status(&checks).await;
        let performance = self.get_current_performance_metrics().await;
        
        HealthStatus {
            status: overall_status,
            timestamp: SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs(),
            service: "stream_server".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
            uptime_seconds: self.start_time.elapsed().unwrap_or_default().as_secs(),
            checks,
            alerts,
            performance,
        }
    }

    async fn calculate_overall_status(&self, checks: &HashMap<String, HealthCheck>) -> ServiceStatus {
        let mut critical_count = 0;
        let mut warning_count = 0;
        let mut total_count = 0;

        for check in checks.values() {
            total_count += 1;
            match check.status {
                CheckStatus::Fail => critical_count += 1,
                CheckStatus::Warn => warning_count += 1,
                CheckStatus::Pass => {}
            }
        }

        if critical_count > 0 {
            if critical_count as f32 / total_count as f32 > 0.5 {
                ServiceStatus::Critical
            } else {
                ServiceStatus::Unhealthy
            }
        } else if warning_count > 0 {
            ServiceStatus::Degraded
        } else {
            ServiceStatus::Healthy
        }
    }

    // Méthodes utilitaires (à implémenter selon votre environnement)
    async fn get_cpu_usage(&self) -> f32 {
        // Implémentation simplifiée - à remplacer par une vraie mesure
        50.0
    }

    async fn get_memory_info(&self) -> (u64, f32) {
        // Retourne (memory_mb, percentage)
        (1024, 60.0)
    }

    async fn get_network_connections(&self) -> u32 {
        100
    }

    async fn get_disk_space(&self, _path: &str) -> DiskInfo {
        DiskInfo {
            total_bytes: 1024 * 1024 * 1024 * 100, // 100GB
            free_bytes: 1024 * 1024 * 1024 * 80,   // 80GB
            free_percentage: 80.0,
        }
    }

    async fn test_database_connection(&self) -> bool {
        true // Simulé
    }

    async fn test_redis_connection(&self) -> bool {
        true // Simulé
    }

    async fn test_jaeger_connection(&self, _endpoint: &str) -> bool {
        true // Simulé
    }

    async fn check_audio_directory_access(&self) -> (CheckStatus, String, u32) {
        (CheckStatus::Pass, "Répertoire audio accessible".to_string(), 42)
    }

    async fn get_cache_statistics(&self) -> CacheStats {
        CacheStats {
            hit_rate: 0.85,
            entries: 500,
            memory_usage_mb: 128,
        }
    }

    async fn get_websocket_statistics(&self) -> WebSocketStats {
        WebSocketStats {
            active_connections: 150,
            error_rate: 0.02,
        }
    }

    async fn get_current_performance_metrics(&self) -> PerformanceMetrics {
        PerformanceMetrics {
            cpu_usage_percent: 45.0,
            memory_usage_mb: 512,
            memory_usage_percent: 60.0,
            disk_usage_percent: 20.0,
            network_connections: 100,
            response_times: ResponseTimeMetrics {
                p50_ms: 25.0,
                p95_ms: 150.0,
                p99_ms: 300.0,
                average_ms: 45.0,
            },
            error_rates: ErrorRateMetrics {
                rate_1min: 0.01,
                rate_5min: 0.015,
                rate_15min: 0.02,
                total_errors_24h: 45,
            },
        }
    }

    async fn update_check(&self, name: &str, check: HealthCheck) {
        let mut checks = self.checks.write().await;
        checks.insert(name.to_string(), check);
    }

    async fn get_last_success(&self, name: &str) -> Option<u64> {
        let checks = self.checks.read().await;
        checks.get(name).and_then(|c| c.last_success)
    }

    async fn get_last_failure(&self, name: &str) -> Option<u64> {
        let checks = self.checks.read().await;
        checks.get(name).and_then(|c| c.last_failure)
    }

    async fn increment_failure_count(&self, name: &str) -> u32 {
        let checks = self.checks.read().await;
        checks.get(name).map(|c| c.failure_count + 1).unwrap_or(1)
    }

    async fn cleanup_old_data(&self) {
        let mut alerts = self.alerts.write().await;
        let cutoff = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() - (24 * 3600); // 24h
        
        alerts.retain(|alert| alert.timestamp > cutoff);
    }
}

impl Clone for HealthMonitor {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            start_time: self.start_time,
            checks: self.checks.clone(),
            alerts: self.alerts.clone(),
            performance_history: self.performance_history.clone(),
        }
    }
}

struct DiskInfo {
    total_bytes: u64,
    free_bytes: u64,
    free_percentage: f32,
}

struct CacheStats {
    hit_rate: f32,
    entries: u32,
    memory_usage_mb: u32,
}

struct WebSocketStats {
    active_connections: u32,
    error_rate: f32,
} 