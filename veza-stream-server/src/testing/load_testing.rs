/// Module Load Testing pour 100k+ connexions simultanées

use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{RwLock, Semaphore};
use tracing::{info, debug};
use uuid::Uuid;
use futures::future::join_all;

use crate::error::AppError;
use crate::core::StreamManager;
use super::{TestConnection, ConnectionStatus, ConnectionMetrics};

/// Load Tester pour simulation de charge massive
#[derive(Debug)]
pub struct LoadTester {
    max_connections: u32,
    ramp_up_duration: Duration,
    test_duration: Duration,
    connection_semaphore: Arc<Semaphore>,
}

impl LoadTester {
    pub fn new(
        max_connections: u32,
        ramp_up_duration: Duration,
        test_duration: Duration,
    ) -> Self {
        Self {
            max_connections,
            ramp_up_duration,
            test_duration,
            connection_semaphore: Arc::new(Semaphore::new(max_connections as usize)),
        }
    }

    /// Execute le test de charge
    pub async fn execute(
        &self,
        stream_manager: Arc<StreamManager>,
        active_connections: Arc<RwLock<Vec<TestConnection>>>,
    ) -> Result<(), AppError> {
        info!("🚀 Démarrage Load Test: {} connexions", self.max_connections);
        
        let start_time = Instant::now();
        
        // Phase 1: Ramp-up progressif
        self.ramp_up_phase(stream_manager.clone(), active_connections.clone()).await?;
        
        // Phase 2: Maintien de la charge
        self.sustain_phase(active_connections.clone()).await?;
        
        // Phase 3: Ramp-down
        self.ramp_down_phase(active_connections.clone()).await?;
        
        let total_duration = start_time.elapsed();
        info!("✅ Load Test terminé en {:?}", total_duration);
        
        Ok(())
    }

    /// Phase de montée en charge progressive
    async fn ramp_up_phase(
        &self,
        stream_manager: Arc<StreamManager>,
        active_connections: Arc<RwLock<Vec<TestConnection>>>,
    ) -> Result<(), AppError> {
        info!("📈 Phase Ramp-up: {} connexions en {:?}", 
              self.max_connections, self.ramp_up_duration);
        
        let connections_per_second = self.max_connections as f64 / self.ramp_up_duration.as_secs_f64();
        let interval_ms = (1000.0 / connections_per_second) as u64;
        
        let mut interval = tokio::time::interval(Duration::from_millis(interval_ms));
        let mut created_connections = 0;
        
        let _stream_manager = stream_manager.clone();
        
        while created_connections < self.max_connections {
            interval.tick().await;
            
            // Créer batch de connexions
            let batch_size = (connections_per_second / 10.0).ceil() as u32;
            let mut batch_futures = Vec::new();
            
            for _ in 0..batch_size.min(self.max_connections - created_connections) {
                let permit = self.connection_semaphore.clone().acquire_owned().await
                                         .map_err(|e| AppError::NetworkError { message: format!("Semaphore error: {}", e) })?;
                
                let connection_future = self.create_test_connection(permit);
                batch_futures.push(connection_future);
                created_connections += 1;
            }
            
            let new_connections = join_all(batch_futures).await;
            let mut connections_guard = active_connections.write().await;
            
            for connection_result in new_connections {
                if let Ok(connection) = connection_result {
                    connections_guard.push(connection);
                }
            }
            
            if created_connections % 10_000 == 0 {
                info!("📊 Connexions créées: {}/{}", created_connections, self.max_connections);
            }
        }
        
        info!("✅ Ramp-up terminé: {} connexions actives", created_connections);
        Ok(())
    }

    /// Phase de maintien de la charge
    async fn sustain_phase(
        &self,
        active_connections: Arc<RwLock<Vec<TestConnection>>>,
    ) -> Result<(), AppError> {
        info!("⚡ Phase Sustain: maintien charge pendant {:?}", self.test_duration);
        
        let mut interval = tokio::time::interval(Duration::from_secs(10));
        let sustain_end = Instant::now() + self.test_duration;
        
        while Instant::now() < sustain_end {
            interval.tick().await;
            
            // Simuler activité sur les connexions
            self.simulate_connection_activity(active_connections.clone()).await;
            
            // Monitoring des métriques
            let active_count = active_connections.read().await.len();
            debug!("📈 Connexions actives: {}", active_count);
        }
        
        info!("✅ Phase sustain terminée");
        Ok(())
    }

    /// Phase de descente de charge
    async fn ramp_down_phase(
        &self,
        active_connections: Arc<RwLock<Vec<TestConnection>>>,
    ) -> Result<(), AppError> {
        info!("📉 Phase Ramp-down: fermeture progressive des connexions");
        
        let initial_count = active_connections.read().await.len();
        let connections_per_second = initial_count as f64 / Duration::from_secs(120).as_secs_f64(); // 2 minutes
        let interval_ms = (1000.0 / connections_per_second) as u64;
        
        let mut interval = tokio::time::interval(Duration::from_millis(interval_ms));
        
        while !active_connections.read().await.is_empty() {
            interval.tick().await;
            
            // Fermer batch de connexions
            let batch_size = (connections_per_second / 10.0).ceil() as usize;
            let mut connections_guard = active_connections.write().await;
            
            let current_len = connections_guard.len();
            let to_remove = batch_size.min(current_len);
            connections_guard.truncate(current_len - to_remove);
            
            if connections_guard.len() % 10_000 == 0 {
                info!("📊 Connexions restantes: {}", connections_guard.len());
            }
        }
        
        info!("✅ Ramp-down terminé: toutes les connexions fermées");
        Ok(())
    }

    /// Crée une connexion de test
    async fn create_test_connection(
        &self,
        _permit: tokio::sync::OwnedSemaphorePermit,
    ) -> Result<TestConnection, AppError> {
        let connection = TestConnection {
            id: Uuid::new_v4(),
            created_at: Instant::now(),
            last_activity: Instant::now(),
            status: ConnectionStatus::Connected,
            metrics: ConnectionMetrics::default(),
        };
        
        // Simulation d'établissement de connexion
        tokio::time::sleep(Duration::from_millis(1)).await;
        
        Ok(connection)
    }

    /// Simule l'activité sur les connexions
    async fn simulate_connection_activity(
        &self,
        active_connections: Arc<RwLock<Vec<TestConnection>>>,
    ) {
        let mut connections_guard = active_connections.write().await;
        
        for connection in connections_guard.iter_mut() {
            // Simuler activité
            connection.last_activity = Instant::now();
            connection.metrics.requests_sent += 1;
            connection.metrics.responses_received += 1;
            connection.metrics.bytes_sent += 1024;
            connection.metrics.bytes_received += 2048;
            
            // Simuler quelques erreurs occasionnelles
            if rand::random::<f64>() < 0.001 { // 0.1% d'erreurs
                connection.metrics.errors += 1;
                connection.status = ConnectionStatus::Error { 
                    message: "Simulated error".to_string() 
                };
            } else {
                connection.status = ConnectionStatus::Active;
            }
        }
    }
}

/// Configuration spécialisée pour différents patterns de charge
#[derive(Debug, Clone)]
pub enum LoadPattern {
    /// Montée linéaire constante
    Linear,
    /// Montée exponentielle
    Exponential,
    /// Pics de charge soudains
    Spike { spike_interval: Duration },
    /// Charge en vagues
    Wave { wave_period: Duration },
}

/// Générateur de charges spécialisées
pub struct LoadPatternGenerator {
    pattern: LoadPattern,
    max_connections: u32,
}

impl LoadPatternGenerator {
    pub fn new(pattern: LoadPattern, max_connections: u32) -> Self {
        Self { pattern, max_connections }
    }

    /// Calcule le nombre de connexions à créer à un instant donné
    pub fn connections_at_time(&self, elapsed: Duration, total_duration: Duration) -> u32 {
        let progress = elapsed.as_secs_f64() / total_duration.as_secs_f64();
        
        match &self.pattern {
            LoadPattern::Linear => {
                (progress * self.max_connections as f64) as u32
            },
            LoadPattern::Exponential => {
                let exp_progress = progress * progress * progress;
                (exp_progress * self.max_connections as f64) as u32
            },
            LoadPattern::Spike { spike_interval } => {
                let spike_cycle = elapsed.as_secs_f64() % spike_interval.as_secs_f64();
                if spike_cycle < spike_interval.as_secs_f64() * 0.1 { // 10% du cycle = spike
                    self.max_connections
                } else {
                    (self.max_connections as f64 * 0.3) as u32 // 30% baseline
                }
            },
            LoadPattern::Wave { wave_period } => {
                let wave_progress = (elapsed.as_secs_f64() / wave_period.as_secs_f64() * 2.0 * std::f64::consts::PI).sin();
                let normalized = (wave_progress + 1.0) / 2.0; // Normaliser entre 0 et 1
                (normalized * self.max_connections as f64) as u32
            },
        }
    }
}
