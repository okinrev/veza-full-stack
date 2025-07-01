/// Module Chaos Testing pour tests de résilience

use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::RwLock;
use tracing::{info, warn, error, debug};
use uuid::Uuid;
use rand::Rng;

use crate::error::AppError;
use crate::core::StreamManager;
use super::{ChaosConfig, TestEvent, TestEventType, EventSeverity};

/// Chaos Tester pour validation de la résilience
#[derive(Debug)]
pub struct ChaosTester {
    config: ChaosConfig,
    chaos_events: Arc<RwLock<Vec<ChaosEvent>>>,
    active_failures: Arc<RwLock<Vec<ActiveFailure>>>,
}

/// Événement de chaos injecté
#[derive(Debug, Clone)]
pub struct ChaosEvent {
    pub id: Uuid,
    pub event_type: ChaosEventType,
    pub started_at: Instant,
    pub duration: Duration,
    pub impact_level: ImpactLevel,
    pub recovery_time: Option<Duration>,
}

/// Types d'événements de chaos
#[derive(Debug, Clone)]
pub enum ChaosEventType {
    NetworkPartition,
    HighLatency,
    PacketLoss,
    ServiceCrash,
    ResourceExhaustion,
    DatabaseFailure,
    DiskFull,
    CpuSpike,
    MemoryLeak,
}

/// Niveau d'impact des événements
#[derive(Debug, Clone)]
pub enum ImpactLevel {
    Low,    // 1-10% des connexions affectées
    Medium, // 10-30% des connexions affectées
    High,   // 30-70% des connexions affectées
    Severe, // 70%+ des connexions affectées
}

/// Panne active en cours
#[derive(Debug, Clone)]
pub struct ActiveFailure {
    pub id: Uuid,
    pub failure_type: FailureType,
    pub started_at: Instant,
    pub affected_services: Vec<String>,
    pub mitigation_status: MitigationStatus,
}

/// Types de pannes
#[derive(Debug, Clone)]
pub enum FailureType {
    NetworkFailure { latency_ms: u64, packet_loss: f64 },
    ServiceUnavailable { service_name: String },
    ResourceOverload { resource_type: String, usage_percent: f64 },
    DataCorruption { affected_data: String },
}

/// Status de mitigation
#[derive(Debug, Clone)]
pub enum MitigationStatus {
    Detected,
    Mitigating,
    Recovering,
    Resolved,
}

impl ChaosTester {
    pub fn new(config: ChaosConfig) -> Self {
        Self {
            config,
            chaos_events: Arc::new(RwLock::new(Vec::new())),
            active_failures: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Execute les tests de chaos
    pub async fn execute(
        &self,
        stream_manager: Arc<StreamManager>,
        test_duration: Duration,
    ) -> Result<(), AppError> {
        info!("🌪️  Démarrage Chaos Testing pendant {:?}", test_duration);
        
        let start_time = Instant::now();
        let end_time = start_time + test_duration;
        
        // Lancer les générateurs de chaos en parallèle
        let tasks = vec![
            self.network_chaos_generator(),
            self.service_chaos_generator(),
            self.resource_chaos_generator(),
            self.monitoring_chaos_events(),
        ];
        
        let (_, _, _, _) = tokio::join!(
            tasks[0],
            tasks[1], 
            tasks[2],
            tasks[3]
        );
        
        // Attendre la fin du test
        while Instant::now() < end_time {
            tokio::time::sleep(Duration::from_secs(1)).await;
        }
        
        // Arrêter toutes les pannes actives
        self.stop_all_failures().await?;
        
        let total_events = self.chaos_events.read().await.len();
        info!("✅ Chaos Testing terminé: {} événements injectés", total_events);
        
        Ok(())
    }

    /// Générateur de chaos réseau
    async fn network_chaos_generator(&self) {
        info!("📡 Démarrage générateur chaos réseau");
        
        let mut interval = tokio::time::interval(Duration::from_secs(30));
        
        loop {
            interval.tick().await;
            
            if rand::random::<f64>() < self.config.network_failure_rate {
                self.inject_network_failure().await;
            }
            
            if rand::random::<f64>() < 0.3 { // 30% chance de latence élevée
                self.inject_high_latency().await;
            }
            
            if rand::random::<f64>() < 0.2 { // 20% chance de perte paquets
                self.inject_packet_loss().await;
            }
        }
    }

    /// Générateur de chaos de services
    async fn service_chaos_generator(&self) {
        info!("🔧 Démarrage générateur chaos services");
        
        let mut interval = tokio::time::interval(self.config.service_restart_interval);
        
        loop {
            interval.tick().await;
            
            if rand::random::<f64>() < 0.1 { // 10% chance de crash service
                self.inject_service_crash().await;
            }
        }
    }

    /// Générateur de chaos de ressources
    async fn resource_chaos_generator(&self) {
        info!("💾 Démarrage générateur chaos ressources");
        
        let mut interval = tokio::time::interval(Duration::from_secs(60));
        
        loop {
            interval.tick().await;
            
            let chaos_type = rand::thread_rng().gen_range(0..4);
            match chaos_type {
                0 => self.inject_cpu_spike().await,
                1 => self.inject_memory_pressure().await,
                2 => self.inject_disk_full().await,
                3 => self.inject_database_slowdown().await,
                _ => {}
            }
        }
    }

    /// Monitoring des événements de chaos
    async fn monitoring_chaos_events(&self) {
        info!("📊 Démarrage monitoring événements chaos");
        
        let mut interval = tokio::time::interval(Duration::from_secs(5));
        
        loop {
            interval.tick().await;
            
            // Vérifier les pannes actives
            let active_failures = self.active_failures.read().await;
            if !active_failures.is_empty() {
                debug!("⚠️  Pannes actives: {}", active_failures.len());
                
                for failure in active_failures.iter() {
                    let duration = failure.started_at.elapsed();
                    if duration > Duration::from_secs(300) { // Plus de 5 minutes
                        warn!("🚨 Panne prolongée détectée: {:?}", failure.failure_type);
                    }
                }
            }
            
            // Nettoyer les pannes résolues
            self.cleanup_resolved_failures().await;
        }
    }

    /// Injecte une panne réseau
    async fn inject_network_failure(&self) {
        let latency = rand::thread_rng().gen_range(100..2000); // 100-2000ms
        let packet_loss = rand::thread_rng().gen_range(0.01..0.1); // 1-10%
        
        warn!("📡 Injection panne réseau: latence +{}ms, perte {}%", latency, packet_loss * 100.0);
        
        let failure = ActiveFailure {
            id: Uuid::new_v4(),
            failure_type: FailureType::NetworkFailure {
                latency_ms: latency,
                packet_loss,
            },
            started_at: Instant::now(),
            affected_services: vec!["stream_server".to_string(), "grpc_service".to_string()],
            mitigation_status: MitigationStatus::Detected,
        };
        
        self.active_failures.write().await.push(failure);
        
        // Simuler la panne pendant 30-120 secondes
        let duration = Duration::from_secs(rand::thread_rng().gen_range(30..120));
        
        tokio::spawn({
            let active_failures = self.active_failures.clone();
            let failure_id = failure.id;
            
            async move {
                tokio::time::sleep(duration).await;
                
                let mut failures = active_failures.write().await;
                failures.retain(|f| f.id != failure_id);
                
                info!("✅ Panne réseau résolue après {:?}", duration);
            }
        });
    }

    /// Injecte une latence élevée
    async fn inject_high_latency(&self) {
        let extra_latency = self.config.artificial_latency_ms;
        debug!("🐌 Injection latence élevée: +{}ms", extra_latency);
        
        // Simulation d'ajout de latence
        tokio::time::sleep(Duration::from_millis(extra_latency)).await;
    }

    /// Injecte une perte de paquets
    async fn inject_packet_loss(&self) {
        let loss_rate = self.config.packet_loss_rate;
        debug!("📦 Injection perte paquets: {:.1}%", loss_rate * 100.0);
        
        // Simulation de perte de paquets
        if rand::random::<f64>() < loss_rate {
            // Simuler la perte en ne traitant pas la requête
            return;
        }
    }

    /// Injecte un crash de service
    async fn inject_service_crash(&self) {
        let services = vec!["auth_service", "stream_processor", "event_bus"];
        let service = services[rand::thread_rng().gen_range(0..services.len())];
        
        warn!("💥 Injection crash service: {}", service);
        
        let failure = ActiveFailure {
            id: Uuid::new_v4(),
            failure_type: FailureType::ServiceUnavailable {
                service_name: service.to_string(),
            },
            started_at: Instant::now(),
            affected_services: vec![service.to_string()],
            mitigation_status: MitigationStatus::Detected,
        };
        
        self.active_failures.write().await.push(failure);
        
        // Simuler redémarrage du service après 10-60 secondes
        let restart_time = Duration::from_secs(rand::thread_rng().gen_range(10..60));
        
        tokio::spawn({
            let active_failures = self.active_failures.clone();
            let failure_id = failure.id;
            
            async move {
                tokio::time::sleep(restart_time).await;
                
                let mut failures = active_failures.write().await;
                failures.retain(|f| f.id != failure_id);
                
                info!("🔄 Service redémarré: {} après {:?}", service, restart_time);
            }
        });
    }

    /// Injecte un pic CPU
    async fn inject_cpu_spike(&self) {
        debug!("🔥 Injection pic CPU");
        
        // Simuler charge CPU intensive
        let start = Instant::now();
        while start.elapsed() < Duration::from_millis(100) {
            // Calculs intensifs pour simuler la charge
            let _ = (0..10000).map(|x| x * x).sum::<i32>();
        }
    }

    /// Injecte une pression mémoire
    async fn inject_memory_pressure(&self) {
        debug!("🧠 Injection pression mémoire");
        
        // Simuler allocation mémoire importante (temporaire)
        let _temp_memory: Vec<u8> = vec![0; 10_000_000]; // 10MB
        tokio::time::sleep(Duration::from_secs(5)).await;
        // La mémoire sera libérée à la fin de la scope
    }

    /// Injecte un disque plein
    async fn inject_disk_full(&self) {
        debug!("💽 Injection disque plein (simulation)");
        
        // En production, cela simulerait un disque plein
        // Ici on simule juste l'événement
        let failure = ActiveFailure {
            id: Uuid::new_v4(),
            failure_type: FailureType::ResourceOverload {
                resource_type: "disk".to_string(),
                usage_percent: 98.5,
            },
            started_at: Instant::now(),
            affected_services: vec!["stream_server".to_string()],
            mitigation_status: MitigationStatus::Detected,
        };
        
        self.active_failures.write().await.push(failure);
    }

    /// Injecte un ralentissement base de données
    async fn inject_database_slowdown(&self) {
        debug!("🗄️  Injection ralentissement BDD");
        
        // Simuler requêtes BDD lentes
        tokio::time::sleep(Duration::from_millis(500)).await;
    }

    /// Nettoie les pannes résolues
    async fn cleanup_resolved_failures(&self) {
        let mut failures = self.active_failures.write().await;
        let initial_count = failures.len();
        
        failures.retain(|failure| {
            // Garder seulement les pannes récentes (< 10 minutes)
            failure.started_at.elapsed() < Duration::from_secs(600)
        });
        
        let cleaned = initial_count - failures.len();
        if cleaned > 0 {
            debug!("🧹 Nettoyé {} pannes expirées", cleaned);
        }
    }

    /// Arrête toutes les pannes actives
    async fn stop_all_failures(&self) -> Result<(), AppError> {
        let mut failures = self.active_failures.write().await;
        let count = failures.len();
        failures.clear();
        
        info!("🛑 Arrêt de {} pannes actives", count);
        Ok(())
    }

    /// Obtient les statistiques de chaos
    pub async fn get_chaos_stats(&self) -> ChaosStats {
        let events = self.chaos_events.read().await;
        let active_failures = self.active_failures.read().await;
        
        ChaosStats {
            total_chaos_events: events.len(),
            active_failures_count: active_failures.len(),
            network_failures: events.iter().filter(|e| matches!(e.event_type, ChaosEventType::NetworkPartition)).count(),
            service_crashes: events.iter().filter(|e| matches!(e.event_type, ChaosEventType::ServiceCrash)).count(),
            resource_exhaustions: events.iter().filter(|e| matches!(e.event_type, ChaosEventType::ResourceExhaustion)).count(),
        }
    }
}

/// Statistiques des tests de chaos
#[derive(Debug, Clone)]
pub struct ChaosStats {
    pub total_chaos_events: usize,
    pub active_failures_count: usize,
    pub network_failures: usize,
    pub service_crashes: usize,
    pub resource_exhaustions: usize,
}
