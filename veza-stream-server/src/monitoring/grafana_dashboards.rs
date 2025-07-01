/// Module Grafana pour dashboards production

use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, debug, error};
use serde::{Serialize, Deserialize};
use serde_json::{json, Value};

use crate::error::AppError;

/// Gestionnaire de dashboards Grafana
#[derive(Debug)]
pub struct GrafanaManager {
    dashboards: Arc<RwLock<Vec<GrafanaDashboard>>>,
    config: GrafanaConfig,
}

/// Configuration Grafana
#[derive(Debug, Clone)]
pub struct GrafanaConfig {
    pub grafana_url: String,
    pub api_key: Option<String>,
    pub org_id: u32,
    pub datasource_uid: String,
}

/// Dashboard Grafana
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GrafanaDashboard {
    pub id: String,
    pub title: String,
    pub description: String,
    pub tags: Vec<String>,
    pub panels: Vec<GrafanaPanel>,
    pub template_variables: Vec<TemplateVariable>,
    pub time_range: TimeRange,
    pub refresh_interval: String,
}

/// Panel Grafana
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GrafanaPanel {
    pub id: u32,
    pub title: String,
    pub panel_type: PanelType,
    pub targets: Vec<PrometheusQuery>,
    pub position: PanelPosition,
    pub options: PanelOptions,
    pub thresholds: Option<Vec<Threshold>>,
}

/// Types de panels
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum PanelType {
    Graph,
    Stat,
    Table,
    Heatmap,
    Gauge,
    BarGauge,
    Logs,
    NodeGraph,
}

/// Requête Prometheus
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrometheusQuery {
    pub expr: String,
    pub legend: String,
    pub interval: Option<String>,
    pub instant: bool,
}

/// Position du panel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PanelPosition {
    pub x: u32,
    pub y: u32,
    pub width: u32,
    pub height: u32,
}

/// Options du panel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PanelOptions {
    pub unit: Option<String>,
    pub decimals: Option<u32>,
    pub min: Option<f64>,
    pub max: Option<f64>,
    pub color_mode: Option<String>,
}

/// Seuil d'alerte
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Threshold {
    pub value: f64,
    pub color: String,
    pub op: String, // "gt", "lt", etc.
}

/// Variable de template
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TemplateVariable {
    pub name: String,
    pub label: String,
    pub var_type: String,
    pub query: String,
    pub multi: bool,
}

/// Plage de temps
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TimeRange {
    pub from: String,
    pub to: String,
}

impl Default for GrafanaConfig {
    fn default() -> Self {
        Self {
            grafana_url: "http://localhost:3000".to_string(),
            api_key: None,
            org_id: 1,
            datasource_uid: "prometheus".to_string(),
        }
    }
}

impl GrafanaManager {
    /// Crée un nouveau gestionnaire Grafana
    pub async fn new() -> Result<Self, AppError> {
        info!("📊 Initialisation Grafana Manager");
        
        let mut manager = Self {
            dashboards: Arc::new(RwLock::new(Vec::new())),
            config: GrafanaConfig::default(),
        };
        
        // Créer dashboards par défaut
        manager.create_default_dashboards().await?;
        
        Ok(manager)
    }
    
    /// Crée les dashboards par défaut
    async fn create_default_dashboards(&mut self) -> Result<(), AppError> {
        info!("🎨 Création dashboards Grafana par défaut");
        
        let dashboards = vec![
            self.create_system_overview_dashboard(),
            self.create_application_metrics_dashboard(),
            self.create_business_metrics_dashboard(),
            self.create_alerts_dashboard(),
            self.create_performance_dashboard(),
        ];
        
        let mut dashboard_store = self.dashboards.write().await;
        *dashboard_store = dashboards;
        
        info!("✅ {} dashboards créés", dashboard_store.len());
        Ok(())
    }
    
    /// Dashboard Vue d'ensemble système
    fn create_system_overview_dashboard(&self) -> GrafanaDashboard {
        GrafanaDashboard {
            id: "system-overview".to_string(),
            title: "🖥️ System Overview".to_string(),
            description: "Vue d'ensemble des métriques système".to_string(),
            tags: vec!["system".to_string(), "overview".to_string()],
            panels: vec![
                // CPU Usage
                GrafanaPanel {
                    id: 1,
                    title: "CPU Usage".to_string(),
                    panel_type: PanelType::Gauge,
                    targets: vec![PrometheusQuery {
                        expr: "system_cpu_usage_percent".to_string(),
                        legend: "CPU %".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 0, y: 0, width: 6, height: 4 },
                    options: PanelOptions {
                        unit: Some("percent".to_string()),
                        decimals: Some(1),
                        min: Some(0.0),
                        max: Some(100.0),
                        color_mode: Some("thresholds".to_string()),
                    },
                    thresholds: Some(vec![
                        Threshold { value: 70.0, color: "yellow".to_string(), op: "gt".to_string() },
                        Threshold { value: 85.0, color: "red".to_string(), op: "gt".to_string() },
                    ]),
                },
                // Memory Usage
                GrafanaPanel {
                    id: 2,
                    title: "Memory Usage".to_string(),
                    panel_type: PanelType::Gauge,
                    targets: vec![PrometheusQuery {
                        expr: "system_memory_usage_bytes / system_memory_total_bytes * 100".to_string(),
                        legend: "Memory %".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 6, y: 0, width: 6, height: 4 },
                    options: PanelOptions {
                        unit: Some("percent".to_string()),
                        decimals: Some(1),
                        min: Some(0.0),
                        max: Some(100.0),
                        color_mode: Some("thresholds".to_string()),
                    },
                    thresholds: Some(vec![
                        Threshold { value: 75.0, color: "yellow".to_string(), op: "gt".to_string() },
                        Threshold { value: 90.0, color: "red".to_string(), op: "gt".to_string() },
                    ]),
                },
                // Network I/O
                GrafanaPanel {
                    id: 3,
                    title: "Network I/O".to_string(),
                    panel_type: PanelType::Graph,
                    targets: vec![
                        PrometheusQuery {
                            expr: "rate(system_network_rx_bytes_total[5m])".to_string(),
                            legend: "RX".to_string(),
                            interval: Some("30s".to_string()),
                            instant: false,
                        },
                        PrometheusQuery {
                            expr: "rate(system_network_tx_bytes_total[5m])".to_string(),
                            legend: "TX".to_string(),
                            interval: Some("30s".to_string()),
                            instant: false,
                        },
                    ],
                    position: PanelPosition { x: 0, y: 4, width: 12, height: 6 },
                    options: PanelOptions {
                        unit: Some("bytes".to_string()),
                        decimals: Some(2),
                        min: None,
                        max: None,
                        color_mode: None,
                    },
                    thresholds: None,
                },
            ],
            template_variables: vec![],
            time_range: TimeRange {
                from: "now-1h".to_string(),
                to: "now".to_string(),
            },
            refresh_interval: "30s".to_string(),
        }
    }
    
    /// Dashboard Métriques Application
    fn create_application_metrics_dashboard(&self) -> GrafanaDashboard {
        GrafanaDashboard {
            id: "application-metrics".to_string(),
            title: "🚀 Application Metrics".to_string(),
            description: "Métriques de performance de l'application".to_string(),
            tags: vec!["application".to_string(), "performance".to_string()],
            panels: vec![
                // Active Connections
                GrafanaPanel {
                    id: 1,
                    title: "Active Connections".to_string(),
                    panel_type: PanelType::Stat,
                    targets: vec![PrometheusQuery {
                        expr: "stream_connections_active".to_string(),
                        legend: "Connections".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 0, y: 0, width: 3, height: 3 },
                    options: PanelOptions {
                        unit: Some("short".to_string()),
                        decimals: Some(0),
                        min: None,
                        max: None,
                        color_mode: Some("value".to_string()),
                    },
                    thresholds: Some(vec![
                        Threshold { value: 50000.0, color: "green".to_string(), op: "gt".to_string() },
                        Threshold { value: 80000.0, color: "yellow".to_string(), op: "gt".to_string() },
                        Threshold { value: 95000.0, color: "red".to_string(), op: "gt".to_string() },
                    ]),
                },
                // Requests per Second
                GrafanaPanel {
                    id: 2,
                    title: "Requests/sec".to_string(),
                    panel_type: PanelType::Stat,
                    targets: vec![PrometheusQuery {
                        expr: "rate(http_requests_total[1m])".to_string(),
                        legend: "RPS".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 3, y: 0, width: 3, height: 3 },
                    options: PanelOptions {
                        unit: Some("reqps".to_string()),
                        decimals: Some(1),
                        min: None,
                        max: None,
                        color_mode: Some("value".to_string()),
                    },
                    thresholds: None,
                },
                // Response Time Percentiles
                GrafanaPanel {
                    id: 3,
                    title: "Response Time Percentiles".to_string(),
                    panel_type: PanelType::Graph,
                    targets: vec![
                        PrometheusQuery {
                            expr: "histogram_quantile(0.50, http_request_duration_seconds)".to_string(),
                            legend: "P50".to_string(),
                            interval: Some("30s".to_string()),
                            instant: false,
                        },
                        PrometheusQuery {
                            expr: "histogram_quantile(0.95, http_request_duration_seconds)".to_string(),
                            legend: "P95".to_string(),
                            interval: Some("30s".to_string()),
                            instant: false,
                        },
                        PrometheusQuery {
                            expr: "histogram_quantile(0.99, http_request_duration_seconds)".to_string(),
                            legend: "P99".to_string(),
                            interval: Some("30s".to_string()),
                            instant: false,
                        },
                    ],
                    position: PanelPosition { x: 6, y: 0, width: 6, height: 6 },
                    options: PanelOptions {
                        unit: Some("ms".to_string()),
                        decimals: Some(2),
                        min: Some(0.0),
                        max: None,
                        color_mode: None,
                    },
                    thresholds: Some(vec![
                        Threshold { value: 50.0, color: "red".to_string(), op: "gt".to_string() },
                    ]),
                },
                // Error Rate
                GrafanaPanel {
                    id: 4,
                    title: "Error Rate".to_string(),
                    panel_type: PanelType::Graph,
                    targets: vec![PrometheusQuery {
                        expr: "rate(stream_errors_total[5m]) / rate(http_requests_total[5m]) * 100".to_string(),
                        legend: "Error %".to_string(),
                        interval: Some("1m".to_string()),
                        instant: false,
                    }],
                    position: PanelPosition { x: 0, y: 6, width: 6, height: 4 },
                    options: PanelOptions {
                        unit: Some("percent".to_string()),
                        decimals: Some(3),
                        min: Some(0.0),
                        max: None,
                        color_mode: None,
                    },
                    thresholds: Some(vec![
                        Threshold { value: 0.1, color: "yellow".to_string(), op: "gt".to_string() },
                        Threshold { value: 1.0, color: "red".to_string(), op: "gt".to_string() },
                    ]),
                },
            ],
            template_variables: vec![],
            time_range: TimeRange {
                from: "now-6h".to_string(),
                to: "now".to_string(),
            },
            refresh_interval: "15s".to_string(),
        }
    }
    
    /// Dashboard Métriques Business
    fn create_business_metrics_dashboard(&self) -> GrafanaDashboard {
        GrafanaDashboard {
            id: "business-metrics".to_string(),
            title: "💼 Business Metrics".to_string(),
            description: "Métriques business et engagement utilisateur".to_string(),
            tags: vec!["business".to_string(), "revenue".to_string()],
            panels: vec![
                // Active Users
                GrafanaPanel {
                    id: 1,
                    title: "Active Users".to_string(),
                    panel_type: PanelType::Stat,
                    targets: vec![PrometheusQuery {
                        expr: "business_active_users".to_string(),
                        legend: "Users".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 0, y: 0, width: 4, height: 4 },
                    options: PanelOptions {
                        unit: Some("short".to_string()),
                        decimals: Some(0),
                        min: None,
                        max: None,
                        color_mode: Some("value".to_string()),
                    },
                    thresholds: None,
                },
                // Revenue
                GrafanaPanel {
                    id: 2,
                    title: "Total Revenue".to_string(),
                    panel_type: PanelType::Stat,
                    targets: vec![PrometheusQuery {
                        expr: "business_revenue_total".to_string(),
                        legend: "Revenue".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 4, y: 0, width: 4, height: 4 },
                    options: PanelOptions {
                        unit: Some("currencyUSD".to_string()),
                        decimals: Some(2),
                        min: None,
                        max: None,
                        color_mode: Some("value".to_string()),
                    },
                    thresholds: None,
                },
                // Premium Subscriptions
                GrafanaPanel {
                    id: 3,
                    title: "Premium Subscriptions".to_string(),
                    panel_type: PanelType::Stat,
                    targets: vec![PrometheusQuery {
                        expr: "business_premium_subscriptions".to_string(),
                        legend: "Premium".to_string(),
                        interval: None,
                        instant: true,
                    }],
                    position: PanelPosition { x: 8, y: 0, width: 4, height: 4 },
                    options: PanelOptions {
                        unit: Some("short".to_string()),
                        decimals: Some(0),
                        min: None,
                        max: None,
                        color_mode: Some("value".to_string()),
                    },
                    thresholds: None,
                },
            ],
            template_variables: vec![],
            time_range: TimeRange {
                from: "now-24h".to_string(),
                to: "now".to_string(),
            },
            refresh_interval: "1m".to_string(),
        }
    }
    
    /// Dashboard Alertes
    fn create_alerts_dashboard(&self) -> GrafanaDashboard {
        GrafanaDashboard {
            id: "alerts".to_string(),
            title: "🚨 Alerts & Incidents".to_string(),
            description: "Monitoring des alertes et incidents".to_string(),
            tags: vec!["alerts".to_string(), "incidents".to_string()],
            panels: vec![],
            template_variables: vec![],
            time_range: TimeRange {
                from: "now-24h".to_string(),
                to: "now".to_string(),
            },
            refresh_interval: "1m".to_string(),
        }
    }
    
    /// Dashboard Performance
    fn create_performance_dashboard(&self) -> GrafanaDashboard {
        GrafanaDashboard {
            id: "performance".to_string(),
            title: "⚡ Performance Deep Dive".to_string(),
            description: "Analyse détaillée des performances".to_string(),
            tags: vec!["performance".to_string(), "deep-dive".to_string()],
            panels: vec![],
            template_variables: vec![],
            time_range: TimeRange {
                from: "now-1h".to_string(),
                to: "now".to_string(),
            },
            refresh_interval: "10s".to_string(),
        }
    }
    
    /// Exporte un dashboard au format JSON Grafana
    pub async fn export_dashboard(&self, dashboard_id: &str) -> Result<Value, AppError> {
        let dashboards = self.dashboards.read().await;
        
        if let Some(dashboard) = dashboards.iter().find(|d| d.id == dashboard_id) {
            Ok(json!({
                "dashboard": dashboard,
                "folderId": 0,
                "overwrite": true
            }))
        } else {
            Err(AppError::InvalidData { 
                message: format!("Dashboard not found: {}", dashboard_id) 
            })
        }
    }
    
    /// Liste tous les dashboards disponibles
    pub async fn list_dashboards(&self) -> Vec<String> {
        let dashboards = self.dashboards.read().await;
        dashboards.iter().map(|d| d.id.clone()).collect()
    }
}
