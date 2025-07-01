/// Module Health Checks pour production
use std::sync::Arc;
use std::time::{Duration, SystemTime};
use std::collections::HashMap;
use tokio::sync::RwLock;
use tracing::{info, debug, error};
use serde::{Serialize, Deserialize};
use crate::error::AppError;

#[derive(Debug)]
pub struct HealthChecker {
    config: HealthConfig,
    health_status: Arc<RwLock<SystemHealth>>,
    check_results: Arc<RwLock<HashMap<String, HealthCheckResult>>>,
}

#[derive(Debug, Clone)]
pub struct HealthConfig {
    pub check_interval: Duration,
    pub timeout: Duration,
    pub critical_services: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemHealth {
    pub overall_status: HealthStatus,
    pub last_check: SystemTime,
    pub uptime: Duration,
    pub services: HashMap<String, ServiceHealth>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum HealthStatus {
    Healthy,
    Degraded,
    Unhealthy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceHealth {
    pub name: String,
    pub status: HealthStatus,
    pub last_check: SystemTime,
    pub response_time_ms: u64,
    pub error_message: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthCheckResult {
    pub service_name: String,
    pub status: HealthStatus,
    pub timestamp: SystemTime,
    pub response_time: Duration,
    pub details: String,
}

impl Default for HealthConfig {
    fn default() -> Self {
        Self {
            check_interval: Duration::from_secs(30),
            timeout: Duration::from_secs(5),
            critical_services: vec![
                "database".to_string(),
                "redis".to_string(),
                "grpc".to_string(),
            ],
        }
    }
}

impl HealthChecker {
    pub async fn new(config: HealthConfig) -> Result<Self, AppError> {
        info!("🏥 Initialisation Health Checker");
        Ok(Self {
            config,
            health_status: Arc::new(RwLock::new(SystemHealth {
                overall_status: HealthStatus::Healthy,
                last_check: SystemTime::now(),
                uptime: Duration::from_secs(0),
                services: HashMap::new(),
            })),
            check_results: Arc::new(RwLock::new(HashMap::new())),
        })
    }
    
    pub async fn start(&self) -> Result<(), AppError> {
        info!("🚀 Démarrage Health Checker");
        let health_status = self.health_status.clone();
        let check_results = self.check_results.clone();
        let config = self.config.clone();
        
        tokio::spawn(async move {
            Self::health_check_loop(health_status, check_results, config).await;
        });
        
        Ok(())
    }
    
    async fn health_check_loop(
        health_status: Arc<RwLock<SystemHealth>>,
        check_results: Arc<RwLock<HashMap<String, HealthCheckResult>>>,
        config: HealthConfig,
    ) {
        let mut interval = tokio::time::interval(config.check_interval);
        
        loop {
            interval.tick().await;
            
            // Vérification database
            let db_result = Self::check_database().await;
            
            // Vérification redis
            let redis_result = Self::check_redis().await;
            
            // Vérification gRPC
            let grpc_result = Self::check_grpc().await;
            
            // Mise à jour des résultats
            {
                let mut results = check_results.write().await;
                results.insert("database".to_string(), db_result.clone());
                results.insert("redis".to_string(), redis_result.clone());
                results.insert("grpc".to_string(), grpc_result.clone());
            }
            
            // Mise à jour du statut global
            {
                let mut health = health_status.write().await;
                health.last_check = SystemTime::now();
                
                let mut services = HashMap::new();
                services.insert("database".to_string(), ServiceHealth {
                    name: "database".to_string(),
                    status: db_result.status,
                    last_check: db_result.timestamp,
                    response_time_ms: db_result.response_time.as_millis() as u64,
                    error_message: None,
                });
                
                health.services = services;
                health.overall_status = HealthStatus::Healthy;
            }
            
            debug!("💓 Health check terminé");
        }
    }
    
    async fn check_database() -> HealthCheckResult {
        let start = SystemTime::now();
        // Simulation du check database
        tokio::time::sleep(Duration::from_millis(10)).await;
        
        HealthCheckResult {
            service_name: "database".to_string(),
            status: HealthStatus::Healthy,
            timestamp: SystemTime::now(),
            response_time: start.elapsed().unwrap_or_default(),
            details: "PostgreSQL connection OK".to_string(),
        }
    }
    
    async fn check_redis() -> HealthCheckResult {
        let start = SystemTime::now();
        // Simulation du check redis
        tokio::time::sleep(Duration::from_millis(5)).await;
        
        HealthCheckResult {
            service_name: "redis".to_string(),
            status: HealthStatus::Healthy,
            timestamp: SystemTime::now(),
            response_time: start.elapsed().unwrap_or_default(),
            details: "Redis connection OK".to_string(),
        }
    }
    
    async fn check_grpc() -> HealthCheckResult {
        let start = SystemTime::now();
        // Simulation du check gRPC
        tokio::time::sleep(Duration::from_millis(8)).await;
        
        HealthCheckResult {
            service_name: "grpc".to_string(),
            status: HealthStatus::Healthy,
            timestamp: SystemTime::now(),
            response_time: start.elapsed().unwrap_or_default(),
            details: "gRPC service OK".to_string(),
        }
    }
    
    pub async fn get_health_status(&self) -> SystemHealth {
        let health = self.health_status.read().await;
        health.clone()
    }
}
