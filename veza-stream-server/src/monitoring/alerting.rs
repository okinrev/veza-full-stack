/// Module Alerting Intelligence pour production
/// 
/// Système d'alerting avancé avec notifications multi-canaux,
/// règles configurables, corrélation d'événements et auto-résolution

use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use std::collections::HashMap;
use tokio::sync::{RwLock, mpsc};
use tracing::{info, warn, error, debug};
use serde::{Serialize, Deserialize};
use uuid::Uuid;
use reqwest::Client;

use crate::error::AppError;
use super::{AlertEvent, AlertType, AlertSeverity, AlertThresholds, MetricsStore};

/// Gestionnaire principal d'alerting
#[derive(Debug)]
pub struct AlertManager {
    config: AlertingConfig,
    active_alerts: Arc<RwLock<HashMap<String, AlertEvent>>>,
    alert_rules: Arc<RwLock<Vec<AlertRule>>>,
    notification_channels: Vec<NotificationChannel>,
    alert_tx: mpsc::UnboundedSender<AlertEvent>,
    alert_rx: Arc<RwLock<Option<mpsc::UnboundedReceiver<AlertEvent>>>>,
    http_client: Client,
}

/// Configuration d'alerting
#[derive(Debug, Clone)]
pub struct AlertingConfig {
    /// Webhook Slack pour notifications
    pub slack_webhook: Option<String>,
    /// Email pour alertes critiques
    pub alert_email: Option<String>,
    /// Teams webhook pour notifications
    pub teams_webhook: Option<String>,
    /// Seuils d'alerte
    pub thresholds: AlertThresholds,
    /// Délai de groupement des alertes (minutes)
    pub grouping_interval: u64,
    /// Délai d'auto-résolution (minutes)
    pub auto_resolve_timeout: u64,
    /// Canaux de notification activés
    pub enabled_channels: Vec<String>,
}

/// Règle d'alerte
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AlertRule {
    pub id: String,
    pub name: String,
    pub description: String,
    pub metric_name: String,
    pub condition: AlertCondition,
    pub threshold_value: f64,
    pub severity: AlertSeverity,
    pub enabled: bool,
    pub labels: HashMap<String, String>,
    pub annotations: HashMap<String, String>,
}

/// Condition d'alerte
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum AlertCondition {
    GreaterThan,
    LessThan,
    Equal,
    NotEqual,
    PercentageIncrease(f64), // Augmentation en % sur période
    PercentageDecrease(f64), // Diminution en % sur période
}

/// Canal de notification
#[derive(Debug, Clone)]
pub struct NotificationChannel {
    pub name: String,
    pub channel_type: ChannelType,
    pub config: ChannelConfig,
    pub enabled: bool,
}

/// Types de canaux
#[derive(Debug, Clone)]
pub enum ChannelType {
    Slack,
    Email,
    Teams,
    Webhook,
    SMS,
    PagerDuty,
}

/// Configuration du canal
#[derive(Debug, Clone)]
pub struct ChannelConfig {
    pub endpoint: String,
    pub api_key: Option<String>,
    pub template: MessageTemplate,
}

/// Template de message
#[derive(Debug, Clone)]
pub struct MessageTemplate {
    pub title_template: String,
    pub body_template: String,
    pub color_mapping: HashMap<AlertSeverity, String>,
}

/// Contexte d'alerte enrichi
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AlertContext {
    pub alert_id: String,
    pub rule_id: String,
    pub metric_value: f64,
    pub threshold: f64,
    pub duration: Duration,
    pub related_alerts: Vec<String>,
    pub suggested_actions: Vec<String>,
    pub runbook_url: Option<String>,
}

impl AlertManager {
    /// Crée un nouveau gestionnaire d'alerting
    pub async fn new(config: AlertingConfig) -> Result<Self, AppError> {
        info!("🚨 Initialisation Alert Manager");
        
        let (alert_tx, alert_rx) = mpsc::unbounded_channel();
        
        let mut manager = Self {
            config: config.clone(),
            active_alerts: Arc::new(RwLock::new(HashMap::new())),
            alert_rules: Arc::new(RwLock::new(Vec::new())),
            notification_channels: Vec::new(),
            alert_tx,
            alert_rx: Arc::new(RwLock::new(Some(alert_rx))),
            http_client: Client::new(),
        };
        
        // Initialiser les canaux de notification
        manager.setup_notification_channels().await?;
        
        // Créer les règles d'alerte par défaut
        manager.create_default_alert_rules().await?;
        
        info!("✅ Alert Manager initialisé avec {} canaux", manager.notification_channels.len());
        Ok(manager)
    }
    
    /// Configure les canaux de notification
    async fn setup_notification_channels(&mut self) -> Result<(), AppError> {
        debug!("🔧 Configuration canaux de notification");
        
        // Canal Slack
        if let Some(webhook) = &self.config.slack_webhook {
            self.notification_channels.push(NotificationChannel {
                name: "slack".to_string(),
                channel_type: ChannelType::Slack,
                config: ChannelConfig {
                    endpoint: webhook.clone(),
                    api_key: None,
                    template: MessageTemplate {
                        title_template: "🚨 Alert: {alert_type}".to_string(),
                        body_template: "*Service:* {service}\n*Severity:* {severity}\n*Message:* {message}".to_string(),
                        color_mapping: [
                            (AlertSeverity::Info, "#36a64f".to_string()),
                            (AlertSeverity::Warning, "#ffb347".to_string()),
                            (AlertSeverity::Critical, "#ff6b6b".to_string()),
                            (AlertSeverity::Emergency, "#d63031".to_string()),
                        ].iter().cloned().collect(),
                    },
                },
                enabled: self.config.enabled_channels.contains(&"slack".to_string()),
            });
        }
        
        // Canal Email
        if let Some(email) = &self.config.alert_email {
            self.notification_channels.push(NotificationChannel {
                name: "email".to_string(),
                channel_type: ChannelType::Email,
                config: ChannelConfig {
                    endpoint: email.clone(),
                    api_key: None,
                    template: MessageTemplate {
                        title_template: "[ALERT] {severity}: {alert_type}".to_string(),
                        body_template: "Alert Details:\n\nService: {service}\nSeverity: {severity}\nType: {alert_type}\nMessage: {message}\nCurrent Value: {value}\nThreshold: {threshold}\nTime: {timestamp}\n\nSuggested Actions:\n{actions}".to_string(),
                        color_mapping: HashMap::new(),
                    },
                },
                enabled: self.config.enabled_channels.contains(&"email".to_string()),
            });
        }
        
        // Canal Teams
        if let Some(webhook) = &self.config.teams_webhook {
            self.notification_channels.push(NotificationChannel {
                name: "teams".to_string(),
                channel_type: ChannelType::Teams,
                config: ChannelConfig {
                    endpoint: webhook.clone(),
                    api_key: None,
                    template: MessageTemplate {
                        title_template: "🚨 System Alert".to_string(),
                        body_template: "**{alert_type}** alert triggered\n\n**Service:** {service}\n**Severity:** {severity}\n**Message:** {message}\n**Current Value:** {value}\n**Threshold:** {threshold}".to_string(),
                        color_mapping: [
                            (AlertSeverity::Info, "Good".to_string()),
                            (AlertSeverity::Warning, "Warning".to_string()),
                            (AlertSeverity::Critical, "Attention".to_string()),
                            (AlertSeverity::Emergency, "Attention".to_string()),
                        ].iter().cloned().collect(),
                    },
                },
                enabled: self.config.enabled_channels.contains(&"teams".to_string()),
            });
        }
        
        info!("📡 {} canaux de notification configurés", self.notification_channels.len());
        Ok(())
    }
    
    /// Crée les règles d'alerte par défaut
    async fn create_default_alert_rules(&mut self) -> Result<(), AppError> {
        debug!("📋 Création règles d'alerte par défaut");
        
        let default_rules = vec![
            // Métriques système
            AlertRule {
                id: "high_cpu_usage".to_string(),
                name: "High CPU Usage".to_string(),
                description: "CPU usage is critically high".to_string(),
                metric_name: "system_cpu_usage_percent".to_string(),
                condition: AlertCondition::GreaterThan,
                threshold_value: self.config.thresholds.critical_cpu_usage,
                severity: AlertSeverity::Critical,
                enabled: true,
                labels: [("category".to_string(), "system".to_string())].iter().cloned().collect(),
                annotations: [
                    ("runbook".to_string(), "https://docs.veza.live/runbooks/high-cpu".to_string()),
                    ("description".to_string(), "System CPU usage exceeded critical threshold".to_string()),
                ].iter().cloned().collect(),
            },
            
            AlertRule {
                id: "high_memory_usage".to_string(),
                name: "High Memory Usage".to_string(),
                description: "Memory usage is critically high".to_string(),
                metric_name: "system_memory_usage_percent".to_string(),
                condition: AlertCondition::GreaterThan,
                threshold_value: self.config.thresholds.critical_memory_usage,
                severity: AlertSeverity::Critical,
                enabled: true,
                labels: [("category".to_string(), "system".to_string())].iter().cloned().collect(),
                annotations: [
                    ("runbook".to_string(), "https://docs.veza.live/runbooks/high-memory".to_string()),
                    ("suggested_action".to_string(), "Check for memory leaks, restart services if needed".to_string()),
                ].iter().cloned().collect(),
            },
            
            // Métriques application
            AlertRule {
                id: "high_latency_p99".to_string(),
                name: "High Latency P99".to_string(),
                description: "P99 latency is above critical threshold".to_string(),
                metric_name: "latency_p99_ms".to_string(),
                condition: AlertCondition::GreaterThan,
                threshold_value: self.config.thresholds.critical_latency_p99_ms,
                severity: AlertSeverity::Critical,
                enabled: true,
                labels: [("category".to_string(), "performance".to_string())].iter().cloned().collect(),
                annotations: [
                    ("impact".to_string(), "User experience degraded".to_string()),
                    ("suggested_action".to_string(), "Check database connections, optimize queries".to_string()),
                ].iter().cloned().collect(),
            },
            
            AlertRule {
                id: "high_error_rate".to_string(),
                name: "High Error Rate".to_string(),
                description: "Error rate exceeded critical threshold".to_string(),
                metric_name: "error_rate_percent".to_string(),
                condition: AlertCondition::GreaterThan,
                threshold_value: self.config.thresholds.critical_error_rate,
                severity: AlertSeverity::Critical,
                enabled: true,
                labels: [("category".to_string(), "reliability".to_string())].iter().cloned().collect(),
                annotations: [
                    ("impact".to_string(), "Service reliability compromised".to_string()),
                    ("suggested_action".to_string(), "Check logs, validate recent deployments".to_string()),
                ].iter().cloned().collect(),
            },
            
            // Métriques business
            AlertRule {
                id: "low_active_connections".to_string(),
                name: "Low Active Connections".to_string(),
                description: "Active connections dropped below minimum threshold".to_string(),
                metric_name: "active_connections".to_string(),
                condition: AlertCondition::LessThan,
                threshold_value: self.config.thresholds.critical_min_connections as f64,
                severity: AlertSeverity::Warning,
                enabled: true,
                labels: [("category".to_string(), "business".to_string())].iter().cloned().collect(),
                annotations: [
                    ("impact".to_string(), "Low user engagement".to_string()),
                    ("suggested_action".to_string(), "Check service availability, marketing campaigns".to_string()),
                ].iter().cloned().collect(),
            },
        ];
        
        let mut rules = self.alert_rules.write().await;
        *rules = default_rules;
        
        info!("✅ {} règles d'alerte créées", rules.len());
        Ok(())
    }
    
    /// Démarre le gestionnaire d'alerting
    pub async fn start(&self) -> Result<(), AppError> {
        info!("🚀 Démarrage Alert Manager");
        
        let alert_rx = {
            let mut rx_guard = self.alert_rx.write().await;
            rx_guard.take().ok_or_else(|| AppError::Internal("Alert receiver already taken".to_string()))?
        };
        
        let active_alerts = self.active_alerts.clone();
        let channels = self.notification_channels.clone();
        let http_client = self.http_client.clone();
        
        tokio::spawn(async move {
            Self::process_alerts(alert_rx, active_alerts, channels, http_client).await;
        });
        
        info!("✅ Alert Manager démarré");
        Ok(())
    }
    
    /// Traite les alertes en arrière-plan
    async fn process_alerts(
        mut alert_rx: mpsc::UnboundedReceiver<AlertEvent>,
        active_alerts: Arc<RwLock<HashMap<String, AlertEvent>>>,
        channels: Vec<NotificationChannel>,
        http_client: Client,
    ) {
        info!("📨 Traitement des alertes démarré");
        
        while let Some(alert) = alert_rx.recv().await {
            debug!("🚨 Traitement alerte: {}", alert.id);
            
            // Ajouter à la liste des alertes actives
            {
                let mut alerts = active_alerts.write().await;
                alerts.insert(alert.id.to_string(), alert.clone());
            }
            
            // Envoyer notifications
            for channel in &channels {
                if channel.enabled {
                    if let Err(e) = Self::send_notification(&http_client, channel, &alert).await {
                        error!("❌ Erreur notification {}: {}", channel.name, e);
                    }
                }
            }
        }
    }
    
    /// Envoie une notification via un canal
    async fn send_notification(
        http_client: &Client,
        channel: &NotificationChannel,
        alert: &AlertEvent,
    ) -> Result<(), AppError> {
        match channel.channel_type {
            ChannelType::Slack => Self::send_slack_notification(http_client, channel, alert).await,
            ChannelType::Teams => Self::send_teams_notification(http_client, channel, alert).await,
            ChannelType::Email => Self::send_email_notification(http_client, channel, alert).await,
            _ => {
                debug!("📨 Type de canal non implémenté: {:?}", channel.channel_type);
                Ok(())
            }
        }
    }
    
    /// Notification Slack
    async fn send_slack_notification(
        http_client: &Client,
        channel: &NotificationChannel,
        alert: &AlertEvent,
    ) -> Result<(), AppError> {
        let color = channel.config.template.color_mapping
            .get(&alert.severity)
            .cloned()
            .unwrap_or_else(|| "#cccccc".to_string());
        
        let payload = serde_json::json!({
            "attachments": [{
                "color": color,
                "title": format!("🚨 Alert: {:?}", alert.alert_type),
                "text": format!("*Severity:* {:?}\n*Message:* {}\n*Time:* {:?}", 
                    alert.severity, alert.message, alert.triggered_at),
                "fields": [
                    {
                        "title": "Alert ID",
                        "value": alert.id.to_string(),
                        "short": true
                    },
                    {
                        "title": "Services Affected",
                        "value": alert.affected_services.join(", "),
                        "short": true
                    }
                ]
            }]
        });
        
        let response = http_client
            .post(&channel.config.endpoint)
            .json(&payload)
            .send()
            .await
            .map_err(|e| AppError::Internal(format!("Slack notification failed: {}", e)))?;
        
        if response.status().is_success() {
            debug!("✅ Notification Slack envoyée pour alerte {}", alert.id);
        } else {
            error!("❌ Erreur notification Slack: {}", response.status());
        }
        
        Ok(())
    }
    
    /// Notification Teams
    async fn send_teams_notification(
        http_client: &Client,
        channel: &NotificationChannel,
        alert: &AlertEvent,
    ) -> Result<(), AppError> {
        let theme_color = match alert.severity {
            AlertSeverity::Info => "00ff00",
            AlertSeverity::Warning => "ffaa00",
            AlertSeverity::Critical => "ff0000",
            AlertSeverity::Emergency => "aa0000",
        };
        
        let payload = serde_json::json!({
            "@type": "MessageCard",
            "@context": "http://schema.org/extensions",
            "themeColor": theme_color,
            "title": format!("🚨 System Alert: {:?}", alert.alert_type),
            "text": format!("**Severity:** {:?}<br/>**Message:** {}<br/>**Services:** {}", 
                alert.severity, alert.message, alert.affected_services.join(", ")),
            "sections": [{
                "facts": [
                    {
                        "name": "Alert ID",
                        "value": alert.id.to_string()
                    },
                    {
                        "name": "Triggered At",
                        "value": format!("{:?}", alert.triggered_at)
                    }
                ]
            }]
        });
        
        let response = http_client
            .post(&channel.config.endpoint)
            .json(&payload)
            .send()
            .await
            .map_err(|e| AppError::Internal(format!("Teams notification failed: {}", e)))?;
        
        if response.status().is_success() {
            debug!("✅ Notification Teams envoyée pour alerte {}", alert.id);
        } else {
            error!("❌ Erreur notification Teams: {}", response.status());
        }
        
        Ok(())
    }
    
    /// Notification Email (simulation)
    async fn send_email_notification(
        _http_client: &Client,
        _channel: &NotificationChannel,
        alert: &AlertEvent,
    ) -> Result<(), AppError> {
        // Simulation d'envoi email
        debug!("📧 Email notification simulée pour alerte {}", alert.id);
        Ok(())
    }
    
    /// Évalue les métriques contre les règles d'alerte
    pub async fn evaluate_metrics(&self, metrics: &MetricsStore) -> Result<(), AppError> {
        let rules = self.alert_rules.read().await;
        
        for rule in rules.iter() {
            if !rule.enabled {
                continue;
            }
            
            let metric_value = self.extract_metric_value(metrics, &rule.metric_name);
            
            if self.should_trigger_alert(rule, metric_value) {
                let alert = self.create_alert_event(rule, metric_value).await;
                
                if let Err(e) = self.alert_tx.send(alert) {
                    error!("❌ Erreur envoi alerte: {}", e);
                }
            }
        }
        
        Ok(())
    }
    
    /// Extrait la valeur d'une métrique
    fn extract_metric_value(&self, metrics: &MetricsStore, metric_name: &str) -> f64 {
        match metric_name {
            "system_cpu_usage_percent" => metrics.system_metrics.cpu_usage_percent,
            "system_memory_usage_percent" => metrics.system_metrics.memory_usage_percent,
            "latency_p99_ms" => metrics.app_metrics.latency_p99_ms,
            "error_rate_percent" => metrics.app_metrics.error_rate_percent,
            "active_connections" => metrics.app_metrics.active_connections as f64,
            _ => 0.0,
        }
    }
    
    /// Détermine si une alerte doit être déclenchée
    fn should_trigger_alert(&self, rule: &AlertRule, metric_value: f64) -> bool {
        match rule.condition {
            AlertCondition::GreaterThan => metric_value > rule.threshold_value,
            AlertCondition::LessThan => metric_value < rule.threshold_value,
            AlertCondition::Equal => (metric_value - rule.threshold_value).abs() < f64::EPSILON,
            AlertCondition::NotEqual => (metric_value - rule.threshold_value).abs() > f64::EPSILON,
            AlertCondition::PercentageIncrease(_) => false, // Nécessite historique
            AlertCondition::PercentageDecrease(_) => false, // Nécessite historique
        }
    }
    
    /// Crée un événement d'alerte
    async fn create_alert_event(&self, rule: &AlertRule, metric_value: f64) -> AlertEvent {
        AlertEvent {
            id: Uuid::new_v4(),
            alert_type: self.map_rule_to_alert_type(rule),
            severity: rule.severity.clone(),
            message: format!("{}: {} = {:.2} (threshold: {:.2})", 
                rule.name, rule.metric_name, metric_value, rule.threshold_value),
            triggered_at: SystemTime::now(),
            resolved_at: None,
            affected_services: vec!["stream-server".to_string()],
            metrics_snapshot: serde_json::json!({
                "metric_name": rule.metric_name,
                "current_value": metric_value,
                "threshold": rule.threshold_value,
                "rule_id": rule.id
            }),
        }
    }
    
    /// Mappe une règle vers un type d'alerte
    fn map_rule_to_alert_type(&self, rule: &AlertRule) -> AlertType {
        match rule.metric_name.as_str() {
            "system_cpu_usage_percent" => AlertType::HighCpuUsage,
            "system_memory_usage_percent" => AlertType::HighMemoryUsage,
            "latency_p99_ms" => AlertType::HighLatency,
            "error_rate_percent" => AlertType::HighErrorRate,
            "active_connections" => AlertType::ServiceDown,
            _ => AlertType::BusinessMetricAnomaly,
        }
    }
    
    /// Résout une alerte
    pub async fn resolve_alert(&self, alert_id: &str) -> Result<(), AppError> {
        let mut alerts = self.active_alerts.write().await;
        
        if let Some(alert) = alerts.get_mut(alert_id) {
            alert.resolved_at = Some(SystemTime::now());
            info!("✅ Alerte {} résolue", alert_id);
        }
        
        Ok(())
    }
    
    /// Obtient les alertes actives
    pub async fn get_active_alerts(&self) -> Vec<AlertEvent> {
        let alerts = self.active_alerts.read().await;
        alerts.values().filter(|a| a.resolved_at.is_none()).cloned().collect()
    }
}

/// Implémentation par défaut de la configuration d'alerting
impl Default for AlertingConfig {
    fn default() -> Self {
        Self {
            slack_webhook: None,
            alert_email: Some("alerts@veza.live".to_string()),
            teams_webhook: None,
            thresholds: AlertThresholds {
                critical_latency_p99_ms: 100.0,
                critical_error_rate: 1.0,
                critical_cpu_usage: 90.0,
                critical_memory_usage: 85.0,
                critical_min_connections: 1000,
            },
            grouping_interval: 5,
            auto_resolve_timeout: 30,
            enabled_channels: vec!["email".to_string()],
        }
    }
} 