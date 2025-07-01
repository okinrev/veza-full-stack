/// Module Monitoring & Observabilité pour production
/// 
/// Implémente métriques Prometheus, dashboards Grafana,
/// alerting intelligent et distributed tracing

pub mod prometheus_metrics;
pub mod grafana_dashboards;
pub mod alerting;
pub mod tracing;
pub mod health_checks;

pub use prometheus_metrics::*;
pub use grafana_dashboards::*;
pub use alerting::*;
pub use tracing::*;
pub use health_checks::*;

use std::sync::Arc;
use std::time::{Duration, SystemTime};
use tokio::sync::RwLock;
use tracing::{info, warn, error, debug};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

use crate::error::AppError;

/// Configuration du monitoring
#[derive(Debug, Clone)]
pub struct MonitoringConfig {
    /// Port pour serveur de métriques Prometheus
    pub metrics_port: u16,
    /// Interval de collecte des métriques
    pub collection_interval: Duration,
    /// Activation distributed tracing
    pub enable_tracing: bool,
    /// Endpoint Jaeger pour tracing
    pub jaeger_endpoint: Option<String>,
    /// Configuration alerting
    pub alerting_config: AlertingConfig,
    /// Rétention des métriques
    pub metrics_retention: Duration,
}

/// Configuration d'alerting
#[derive(Debug, Clone)]
pub struct AlertingConfig {
    /// Webhook Slack pour alertes
    pub slack_webhook: Option<String>,
    /// Email pour alertes critiques
    pub alert_email: Option<String>,
    /// Seuils d'alerte
    pub thresholds: AlertThresholds,
}

/// Seuils d'alerte configurables
#[derive(Debug, Clone)]
pub struct AlertThresholds {
    /// Latence P99 critique (ms)
    pub critical_latency_p99_ms: f64,
    /// Taux d'erreur critique (%)
    pub critical_error_rate: f64,
    /// CPU critique (%)
    pub critical_cpu_usage: f64,
    /// Mémoire critique (%)
    pub critical_memory_usage: f64,
    /// Connexions minimum critique
    pub critical_min_connections: u32,
}

/// Gestionnaire principal du monitoring
#[derive(Debug)]
pub struct MonitoringManager {
    config: MonitoringConfig,
    prometheus_collector: Arc<PrometheusCollector>,
    grafana_manager: Arc<GrafanaManager>,
    alert_manager: Arc<AlertManager>,
    tracing_manager: Arc<TracingManager>,
    health_checker: Arc<HealthChecker>,
    metrics_store: Arc<RwLock<MetricsStore>>,
}

/// Store des métriques en mémoire
#[derive(Debug, Default)]
pub struct MetricsStore {
    /// Métriques système
    pub system_metrics: SystemMetrics,
    /// Métriques application
    pub app_metrics: ApplicationMetrics,
    /// Métriques business
    pub business_metrics: BusinessMetrics,
    /// Historique des alertes
    pub alert_history: Vec<AlertEvent>,
}

/// Métriques système
#[derive(Debug, Clone, Default)]
pub struct SystemMetrics {
    pub cpu_usage_percent: f64,
    pub memory_usage_percent: f64,
    pub disk_usage_percent: f64,
    pub network_rx_bytes_per_sec: u64,
    pub network_tx_bytes_per_sec: u64,
    pub open_file_descriptors: u32,
    pub load_average_1m: f64,
    pub uptime_seconds: u64,
}

/// Métriques application
#[derive(Debug, Clone, Default)]
pub struct ApplicationMetrics {
    pub active_connections: u32,
    pub requests_per_second: f64,
    pub latency_p50_ms: f64,
    pub latency_p95_ms: f64,
    pub latency_p99_ms: f64,
    pub error_rate_percent: f64,
    pub cache_hit_rate_percent: f64,
    pub database_connections: u32,
    pub queue_depth: u32,
}

/// Métriques business
#[derive(Debug, Clone, Default)]
pub struct BusinessMetrics {
    pub total_streams: u32,
    pub active_streams: u32,
    pub total_listeners: u32,
    pub revenue_per_hour: f64,
    pub user_engagement_score: f64,
    pub conversion_rate_percent: f64,
    pub premium_users_percent: f64,
}

/// Événement d'alerte
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AlertEvent {
    pub id: Uuid,
    pub alert_type: AlertType,
    pub severity: AlertSeverity,
    pub message: String,
    pub triggered_at: SystemTime,
    pub resolved_at: Option<SystemTime>,
    pub affected_services: Vec<String>,
    pub metrics_snapshot: serde_json::Value,
}

/// Types d'alertes
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AlertType {
    HighLatency,
    HighErrorRate,
    HighCpuUsage,
    HighMemoryUsage,
    ServiceDown,
    DatabaseIssue,
    NetworkIssue,
    SecurityIncident,
    BusinessMetricAnomaly,
}

/// Sévérité des alertes
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AlertSeverity {
    Info,
    Warning,
    Critical,
    Emergency,
}

impl Default for MonitoringConfig {
    fn default() -> Self {
        Self {
            metrics_port: 9090,
            collection_interval: Duration::from_secs(15),
            enable_tracing: true,
            jaeger_endpoint: Some("http://localhost:14268".to_string()),
            alerting_config: AlertingConfig {
                slack_webhook: None,
                alert_email: Some("alerts@veza.live".to_string()),
                thresholds: AlertThresholds {
                    critical_latency_p99_ms: 100.0,
                    critical_error_rate: 1.0,
                    critical_cpu_usage: 90.0,
                    critical_memory_usage: 85.0,
                    critical_min_connections: 1000,
                },
            },
            metrics_retention: Duration::from_secs(7 * 24 * 3600), // 7 jours
        }
    }
}

impl MonitoringManager {
    /// Crée un nouveau gestionnaire de monitoring
    pub async fn new(config: MonitoringConfig) -> Result<Self, AppError> {
        info!("�� Initialisation Monitoring Manager");
        
        let prometheus_collector = Arc::new(PrometheusCollector::new().await?);
        let grafana_manager = Arc::new(GrafanaManager::new().await?);
        let alert_manager = Arc::new(AlertManager::new(config.alerting_config.clone()).await?);
        let tracing_manager = Arc::new(TracingManager::new(config.jaeger_endpoint.clone()).await?);
        let health_checker = Arc::new(HealthChecker::new().await?);
        
        Ok(Self {
            config,
            prometheus_collector,
            grafana_manager,
            alert_manager,
            tracing_manager,
            health_checker,
            metrics_store: Arc::new(RwLock::new(MetricsStore::default())),
        })
    }
    
    /// Démarre tous les services de monitoring
    pub async fn start(&self) -> Result<(), AppError> {
        info!("🚀 Démarrage Monitoring & Observabilité");
        
        // Démarrer collecteur Prometheus
        self.prometheus_collector.start(self.config.metrics_port).await?;
        
        // Démarrer distributed tracing
        if self.config.enable_tracing {
            self.tracing_manager.start().await?;
        }
        
        // Démarrer alerting
        self.alert_manager.start().await?;
        
        // Démarrer health checks
        self.health_checker.start().await?;
        
        // Démarrer collection périodique des métriques
        self.start_metrics_collection().await;
        
        // Démarrer monitoring des seuils d'alerte
        self.start_alert_monitoring().await;
        
        info!("✅ Monitoring démarré avec succès");
        Ok(())
    }
    
    /// Arrête tous les services de monitoring
    pub async fn stop(&self) -> Result<(), AppError> {
        info!("🛑 Arrêt Monitoring & Observabilité");
        
        self.health_checker.stop().await?;
        self.alert_manager.stop().await?;
        
        if self.config.enable_tracing {
            self.tracing_manager.stop().await?;
        }
        
        self.prometheus_collector.stop().await?;
        
        info!("✅ Monitoring arrêté");
        Ok(())
    }
    
    /// Démarre la collection périodique des métriques
    async fn start_metrics_collection(&self) {
        let metrics_store = self.metrics_store.clone();
        let interval = self.config.collection_interval;
        
        tokio::spawn(async move {
            let mut ticker = tokio::time::interval(interval);
            
            loop {
                ticker.tick().await;
                
                // Collecter métriques système
                let system_metrics = Self::collect_system_metrics().await;
                
                // Collecter métriques application
                let app_metrics = Self::collect_application_metrics().await;
                
                // Collecter métriques business
                let business_metrics = Self::collect_business_metrics().await;
                
                // Stocker les métriques
                let mut store = metrics_store.write().await;
                store.system_metrics = system_metrics;
                store.app_metrics = app_metrics;
                store.business_metrics = business_metrics;
                
                debug!("📊 Métriques collectées et stockées");
            }
        });
    }
    
    /// Démarre le monitoring des seuils d'alerte
    async fn start_alert_monitoring(&self) {
        let metrics_store = self.metrics_store.clone();
        let alert_manager = self.alert_manager.clone();
        let thresholds = self.config.alerting_config.thresholds.clone();
        
        tokio::spawn(async move {
            let mut ticker = tokio::time::interval(Duration::from_secs(30));
            
            loop {
                ticker.tick().await;
                
                let store = metrics_store.read().await;
                
                // Vérifier seuils critiques
                if store.app_metrics.latency_p99_ms > thresholds.critical_latency_p99_ms {
                    let alert = AlertEvent {
                        id: Uuid::new_v4(),
                        alert_type: AlertType::HighLatency,
                        severity: AlertSeverity::Critical,
                        message: format!("Latence P99 critique: {:.1}ms", store.app_metrics.latency_p99_ms),
                        triggered_at: SystemTime::now(),
                        resolved_at: None,
                        affected_services: vec!["stream_server".to_string()],
                        metrics_snapshot: serde_json::json!({
                            "latency_p99": store.app_metrics.latency_p99_ms,
                            "active_connections": store.app_metrics.active_connections
                        }),
                    };
                    
                    if let Err(e) = alert_manager.send_alert(alert).await {
                        error!("❌ Erreur envoi alerte: {}", e);
                    }
                }
                
                // Vérifier autres seuils...
                if store.app_metrics.error_rate_percent > thresholds.critical_error_rate {
                    warn!("⚠️  Taux d'erreur élevé: {:.2}%", store.app_metrics.error_rate_percent);
                }
                
                if store.system_metrics.cpu_usage_percent > thresholds.critical_cpu_usage {
                    warn!("⚠️  CPU élevé: {:.1}%", store.system_metrics.cpu_usage_percent);
                }
            }
        });
    }
    
    /// Collecte les métriques système
    async fn collect_system_metrics() -> SystemMetrics {
        // Simulation de collecte de métriques système
        SystemMetrics {
            cpu_usage_percent: 72.5,
            memory_usage_percent: 68.3,
            disk_usage_percent: 45.2,
            network_rx_bytes_per_sec: 1_250_000,
            network_tx_bytes_per_sec: 2_100_000,
            open_file_descriptors: 1024,
            load_average_1m: 2.8,
            uptime_seconds: 86400 * 5, // 5 jours
        }
    }
    
    /// Collecte les métriques application
    async fn collect_application_metrics() -> ApplicationMetrics {
        // Simulation de collecte de métriques application
        ApplicationMetrics {
            active_connections: 85_000,
            requests_per_second: 12_500.0,
            latency_p50_ms: 12.3,
            latency_p95_ms: 28.7,
            latency_p99_ms: 45.2,
            error_rate_percent: 0.08,
            cache_hit_rate_percent: 94.5,
            database_connections: 45,
            queue_depth: 128,
        }
    }
    
    /// Collecte les métriques business
    async fn collect_business_metrics() -> BusinessMetrics {
        // Simulation de collecte de métriques business
        BusinessMetrics {
            total_streams: 1_250,
            active_streams: 987,
            total_listeners: 125_000,
            revenue_per_hour: 1_850.75,
            user_engagement_score: 8.7,
            conversion_rate_percent: 3.2,
            premium_users_percent: 15.8,
        }
    }
    
    /// Obtient un snapshot des métriques actuelles
    pub async fn get_metrics_snapshot(&self) -> MetricsStore {
        self.metrics_store.read().await.clone()
    }
    
    /// Génère un rapport de santé complet
    pub async fn generate_health_report(&self) -> HealthReport {
        let metrics = self.get_metrics_snapshot().await;
        let health_status = self.health_checker.get_overall_health().await;
        
        HealthReport {
            timestamp: SystemTime::now(),
            overall_status: health_status,
            system_metrics: metrics.system_metrics,
            app_metrics: metrics.app_metrics,
            business_metrics: metrics.business_metrics,
            recent_alerts: metrics.alert_history.iter()
                .filter(|alert| {
                    alert.triggered_at
                        .duration_since(SystemTime::now() - Duration::from_secs(3600))
                        .is_ok()
                })
                .cloned()
                .collect(),
        }
    }
}

/// Rapport de santé complet
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthReport {
    pub timestamp: SystemTime,
    pub overall_status: HealthStatus,
    pub system_metrics: SystemMetrics,
    pub app_metrics: ApplicationMetrics,
    pub business_metrics: BusinessMetrics,
    pub recent_alerts: Vec<AlertEvent>,
}

/// Status de santé général
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum HealthStatus {
    Healthy,
    Degraded { issues: Vec<String> },
    Unhealthy { critical_issues: Vec<String> },
}
