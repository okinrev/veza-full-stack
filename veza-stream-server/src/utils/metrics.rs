use std::{
    sync::{
        atomic::{AtomicU64, Ordering},
        Arc,
    },
    time::{Duration, Instant, SystemTime},
};
use tokio::sync::RwLock;
use serde::Serialize;
use std::collections::HashMap;
use crate::Config;


#[derive(Debug, Clone, Serialize)]
pub struct ServerMetrics {
    // Compteurs de requêtes
    pub total_requests: u64,
    pub successful_requests: u64,
    pub failed_requests: u64,
    pub rate_limited_requests: u64,
    pub invalid_signature_requests: u64,
    pub not_found_requests: u64,
    
    // Métriques de performance
    pub total_bytes_served: u64,
    pub average_response_time_ms: f64,
    pub cache_hits: u64,
    pub cache_misses: u64,
    
    // Statistiques temporelles
    pub uptime_seconds: u64,
    pub requests_per_second: f64,
    pub bytes_per_second: f64,
    
    // Top des fichiers les plus demandés
    pub top_files: Vec<FileStats>,
    
    // Statistiques des clients
    pub unique_ips: usize,
    pub top_ips: Vec<IpStats>,
}

#[derive(Debug, Clone, Serialize)]
pub struct FileStats {
    pub filename: String,
    pub requests: u64,
    pub bytes_served: u64,
    pub last_accessed: SystemTime,
}

#[derive(Debug, Clone, Serialize)]
pub struct IpStats {
    pub ip: String,
    pub requests: u64,
    pub bytes_served: u64,
    pub last_seen: SystemTime,
    pub rate_limited_count: u64,
}

#[derive(Debug, Clone)]
pub struct Metrics {
    _config: Arc<Config>,
    counters: Arc<MetricsCounters>,
    start_time: Instant,
}

impl Metrics {
    pub fn new(config: Arc<Config>) -> Self {
        Self {
            _config: config,
            counters: Arc::new(MetricsCounters::new()),
            start_time: Instant::now(),
        }
    }

    pub async fn start_collection(&self) {
        let metrics = self.clone();
        tokio::spawn(async move {
            let mut interval = tokio::time::interval(Duration::from_secs(60));
            loop {
                interval.tick().await;
                metrics.collect_system_metrics().await;
            }
        });
    }

    pub fn increment_requests(&self) {
        self.counters.total_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_successful_requests(&self) {
        self.counters.successful_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_failed_requests(&self) {
        self.counters.failed_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_rate_limited(&self) {
        self.counters.rate_limited_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn add_bytes_served(&self, bytes: u64) {
        self.counters.bytes_served.fetch_add(bytes, Ordering::Relaxed);
    }

    pub async fn record_request(&self, _filename: String, _client_ip: String, bytes_served: u64, _response_time: Duration) {
        self.increment_requests();
        self.add_bytes_served(bytes_served);
        
        // Enregistrer les statistiques par fichier et IP
        // Implémentation simplifiée pour la compilation
    }

    pub async fn get_metrics(&self) -> serde_json::Value {
        let uptime = self.start_time.elapsed().as_secs();
        let total_requests = self.counters.total_requests.load(Ordering::Relaxed);
        let successful_requests = self.counters.successful_requests.load(Ordering::Relaxed);
        let failed_requests = self.counters.failed_requests.load(Ordering::Relaxed);
        let rate_limited_requests = self.counters.rate_limited_requests.load(Ordering::Relaxed);
        let bytes_served = self.counters.bytes_served.load(Ordering::Relaxed);

        serde_json::json!({
            "uptime_seconds": uptime,
            "total_requests": total_requests,
            "successful_requests": successful_requests,
            "failed_requests": failed_requests,
            "rate_limited_requests": rate_limited_requests,
            "bytes_served": bytes_served,
            "requests_per_second": if uptime > 0 { total_requests as f64 / uptime as f64 } else { 0.0 },
            "bytes_per_second": if uptime > 0 { bytes_served as f64 / uptime as f64 } else { 0.0 }
        })
    }

    async fn collect_system_metrics(&self) {
        // Collecter les métriques système
        // Implémentation basique pour la compilation
    }
}

#[derive(Debug)]
pub struct MetricsCounters {
    pub total_requests: AtomicU64,
    pub successful_requests: AtomicU64,
    pub failed_requests: AtomicU64,
    pub rate_limited_requests: AtomicU64,
    pub bytes_served: AtomicU64,
}

impl MetricsCounters {
    pub fn new() -> Self {
        Self {
            total_requests: AtomicU64::new(0),
            successful_requests: AtomicU64::new(0),
            failed_requests: AtomicU64::new(0),
            rate_limited_requests: AtomicU64::new(0),
            bytes_served: AtomicU64::new(0),
        }
    }
}

#[derive(Clone)]
pub struct MetricsCollector {
    // Compteurs atomiques pour les métriques de base
    total_requests: Arc<AtomicU64>,
    successful_requests: Arc<AtomicU64>,
    failed_requests: Arc<AtomicU64>,
    rate_limited_requests: Arc<AtomicU64>,
    invalid_signature_requests: Arc<AtomicU64>,
    not_found_requests: Arc<AtomicU64>,
    total_bytes_served: Arc<AtomicU64>,
    cache_hits: Arc<AtomicU64>,
    cache_misses: Arc<AtomicU64>,
    
    // Métriques détaillées (protégées par RwLock)
    file_stats: Arc<RwLock<HashMap<String, FileStatsInternal>>>,
    ip_stats: Arc<RwLock<HashMap<String, IpStatsInternal>>>,
    response_times: Arc<RwLock<Vec<Duration>>>,
    
    // Timestamp de démarrage
    start_time: Instant,
}

#[derive(Debug, Clone)]
struct FileStatsInternal {
    requests: u64,
    bytes_served: u64,
    last_accessed: SystemTime,
}

#[derive(Debug, Clone)]
struct IpStatsInternal {
    requests: u64,
    bytes_served: u64,
    last_seen: SystemTime,
    rate_limited_count: u64,
}

impl MetricsCollector {
    pub fn new() -> Self {
        Self {
            total_requests: Arc::new(AtomicU64::new(0)),
            successful_requests: Arc::new(AtomicU64::new(0)),
            failed_requests: Arc::new(AtomicU64::new(0)),
            rate_limited_requests: Arc::new(AtomicU64::new(0)),
            invalid_signature_requests: Arc::new(AtomicU64::new(0)),
            not_found_requests: Arc::new(AtomicU64::new(0)),
            total_bytes_served: Arc::new(AtomicU64::new(0)),
            cache_hits: Arc::new(AtomicU64::new(0)),
            cache_misses: Arc::new(AtomicU64::new(0)),
            file_stats: Arc::new(RwLock::new(HashMap::new())),
            ip_stats: Arc::new(RwLock::new(HashMap::new())),
            response_times: Arc::new(RwLock::new(Vec::new())),
            start_time: Instant::now(),
        }
    }

    pub fn increment_total_requests(&self) {
        self.total_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_successful_requests(&self) {
        self.successful_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_failed_requests(&self) {
        self.failed_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_rate_limited(&self) {
        self.rate_limited_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_invalid_signature(&self) {
        self.invalid_signature_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_not_found(&self) {
        self.not_found_requests.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_cache_hits(&self) {
        self.cache_hits.fetch_add(1, Ordering::Relaxed);
    }

    pub fn increment_cache_misses(&self) {
        self.cache_misses.fetch_add(1, Ordering::Relaxed);
    }

    pub fn add_bytes_served(&self, bytes: u64) {
        self.total_bytes_served.fetch_add(bytes, Ordering::Relaxed);
    }

    pub async fn record_request(&self, 
        filename: String, 
        client_ip: String, 
        bytes_served: u64, 
        response_time: Duration,
        is_rate_limited: bool
    ) {
        let now = SystemTime::now();

        // Enregistrement des statistiques de fichier
        {
            let mut file_stats = self.file_stats.write().await;
            let stats = file_stats.entry(filename.clone()).or_insert(FileStatsInternal {
                requests: 0,
                bytes_served: 0,
                last_accessed: now,
            });
            stats.requests += 1;
            stats.bytes_served += bytes_served;
            stats.last_accessed = now;
        }

        // Enregistrement des statistiques d'IP
        {
            let mut ip_stats = self.ip_stats.write().await;
            let stats = ip_stats.entry(client_ip).or_insert(IpStatsInternal {
                requests: 0,
                bytes_served: 0,
                last_seen: now,
                rate_limited_count: 0,
            });
            stats.requests += 1;
            stats.bytes_served += bytes_served;
            stats.last_seen = now;
            if is_rate_limited {
                stats.rate_limited_count += 1;
            }
        }

        // Enregistrement du temps de réponse
        {
            let mut response_times = self.response_times.write().await;
            response_times.push(response_time);
            
            // Limitation à 10000 entrées pour éviter la croissance excessive
            if response_times.len() > 10000 {
                response_times.drain(0..1000); // Supprime les 1000 plus anciennes
            }
        }
    }

    pub async fn get_metrics(&self) -> ServerMetrics {
        let uptime = self.start_time.elapsed();
        let uptime_seconds = uptime.as_secs();

        // Calcul des métriques de base
        let total_requests = self.total_requests.load(Ordering::Relaxed);
        let total_bytes = self.total_bytes_served.load(Ordering::Relaxed);

        let requests_per_second = if uptime_seconds > 0 {
            total_requests as f64 / uptime_seconds as f64
        } else {
            0.0
        };

        let bytes_per_second = if uptime_seconds > 0 {
            total_bytes as f64 / uptime_seconds as f64
        } else {
            0.0
        };

        // Calcul du temps de réponse moyen
        let average_response_time_ms = {
            let response_times = self.response_times.read().await;
            if response_times.is_empty() {
                0.0
            } else {
                let total_ms: u64 = response_times.iter()
                    .map(|d| d.as_millis() as u64)
                    .sum();
                total_ms as f64 / response_times.len() as f64
            }
        };

        // Top des fichiers
        let top_files = {
            let file_stats = self.file_stats.read().await;
            let mut files: Vec<_> = file_stats.iter()
                .map(|(filename, stats)| FileStats {
                    filename: filename.clone(),
                    requests: stats.requests,
                    bytes_served: stats.bytes_served,
                    last_accessed: stats.last_accessed,
                })
                .collect();
            files.sort_by(|a, b| b.requests.cmp(&a.requests));
            files.into_iter().take(10).collect()
        };

        // Top des IPs
        let (unique_ips, top_ips) = {
            let ip_stats = self.ip_stats.read().await;
            let unique_count = ip_stats.len();
            let mut ips: Vec<_> = ip_stats.iter()
                .map(|(ip, stats)| IpStats {
                    ip: ip.clone(),
                    requests: stats.requests,
                    bytes_served: stats.bytes_served,
                    last_seen: stats.last_seen,
                    rate_limited_count: stats.rate_limited_count,
                })
                .collect();
            ips.sort_by(|a, b| b.requests.cmp(&a.requests));
            (unique_count, ips.into_iter().take(10).collect())
        };

        ServerMetrics {
            total_requests,
            successful_requests: self.successful_requests.load(Ordering::Relaxed),
            failed_requests: self.failed_requests.load(Ordering::Relaxed),
            rate_limited_requests: self.rate_limited_requests.load(Ordering::Relaxed),
            invalid_signature_requests: self.invalid_signature_requests.load(Ordering::Relaxed),
            not_found_requests: self.not_found_requests.load(Ordering::Relaxed),
            total_bytes_served: total_bytes,
            average_response_time_ms,
            cache_hits: self.cache_hits.load(Ordering::Relaxed),
            cache_misses: self.cache_misses.load(Ordering::Relaxed),
            uptime_seconds,
            requests_per_second,
            bytes_per_second,
            top_files,
            unique_ips,
            top_ips,
        }
    }

    pub async fn cleanup_old_stats(&self) {
        let cutoff = SystemTime::now() - Duration::from_secs(86400 * 7); // 7 jours

        // Nettoyage des statistiques de fichiers
        {
            let mut file_stats = self.file_stats.write().await;
            file_stats.retain(|_, stats| stats.last_accessed > cutoff);
        }

        // Nettoyage des statistiques d'IP
        {
            let mut ip_stats = self.ip_stats.write().await;
            ip_stats.retain(|_, stats| stats.last_seen > cutoff);
        }
    }
}

impl Default for MetricsCollector {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tokio::time::{sleep, Duration as TokioDuration};

    #[tokio::test]
    async fn test_metrics_basic_operations() {
        let metrics = MetricsCollector::new();

        metrics.increment_total_requests();
        metrics.increment_successful_requests();
        metrics.add_bytes_served(1024);

        let server_metrics = metrics.get_metrics().await;
        assert_eq!(server_metrics.total_requests, 1);
        assert_eq!(server_metrics.successful_requests, 1);
        assert_eq!(server_metrics.total_bytes_served, 1024);
    }

    #[tokio::test]
    async fn test_request_recording() {
        let metrics = MetricsCollector::new();

        metrics.record_request(
            "test.mp3".to_string(),
            "192.168.1.1".to_string(),
            2048,
            Duration::from_millis(100),
            false
        ).await;

        let server_metrics = metrics.get_metrics().await;
        assert_eq!(server_metrics.top_files.len(), 1);
        assert_eq!(server_metrics.top_files[0].filename, "test.mp3");
        assert_eq!(server_metrics.top_files[0].requests, 1);
        assert_eq!(server_metrics.top_files[0].bytes_served, 2048);
    }
} 