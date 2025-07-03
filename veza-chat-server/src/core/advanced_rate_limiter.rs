use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant};
use std::net::IpAddr;
use serde::{Deserialize, Serialize};
use dashmap::DashMap;
use tokio::sync::RwLock;
use parking_lot::Mutex;

use crate::error::Result;

/// Service de rate limiting avancé anti-DDoS
#[derive(Debug)]
pub struct AdvancedRateLimiter {
    /// Limiteurs par IP
    ip_limiters: Arc<DashMap<IpAddr, IpRateLimiter>>,
    
    /// Limiteurs par utilisateur
    user_limiters: Arc<DashMap<i64, UserRateLimiter>>,
    
    /// Limiteurs par canal
    channel_limiters: Arc<DashMap<String, ChannelRateLimiter>>,
    
    /// Patterns d'attaque détectés
    attack_patterns: Arc<DashMap<String, AttackPattern>>,
    
    /// Liste noire temporaire
    blacklist: Arc<DashMap<IpAddr, BlacklistEntry>>,
    
    /// Configuration globale
    config: Arc<RwLock<RateLimitConfig>>,
    
    /// Métriques de performance
    metrics: Arc<RateLimitMetrics>,
}

/// Limiteur par IP avec détection de patterns
#[derive(Debug)]
pub struct IpRateLimiter {
    pub ip: IpAddr,
    pub buckets: HashMap<LimitType, TokenBucket>,
    pub last_activity: Instant,
    pub violation_count: u32,
    pub trust_score: f32,
    pub request_patterns: VecDeque<RequestEvent>,
    pub status: IpStatus,
}

/// Limiteur par utilisateur
#[derive(Debug)]
pub struct UserRateLimiter {
    pub user_id: i64,
    pub buckets: HashMap<LimitType, TokenBucket>,
    pub last_activity: Instant,
    pub violation_count: u32,
    pub reputation: UserReputation,
    pub daily_limits: DailyLimits,
}

/// Limiteur par canal
#[derive(Debug)]
pub struct ChannelRateLimiter {
    pub channel_id: String,
    pub message_bucket: TokenBucket,
    pub concurrent_users: u32,
    pub last_activity: Instant,
    pub spam_threshold: f32,
    pub moderation_level: ModerationLevel,
}

/// Implémentation du Token Bucket
#[derive(Debug, Clone)]
pub struct TokenBucket {
    pub capacity: u32,
    pub tokens: u32,
    pub refill_rate: f32, // tokens per second
    pub last_refill: Instant,
    pub burst_allowance: u32,
}

/// Types de limitations
#[derive(Debug, Clone, Hash, PartialEq, Eq)]
pub enum LimitType {
    /// Messages par minute
    MessagesPerMinute,
    /// Connexions par heure
    ConnectionsPerHour,
    /// Tentatives d'authentification
    AuthAttempts,
    /// Requêtes API
    ApiRequests,
    /// Upload de fichiers
    FileUploads,
    /// Création de channels
    ChannelCreation,
    /// Invitations envoyées
    Invitations,
    /// Réactions ajoutées
    Reactions,
}

/// Statut d'une IP
#[derive(Debug, Clone, PartialEq)]
pub enum IpStatus {
    /// IP normale
    Normal,
    /// IP suspecte (surveillance accrue)
    Suspicious,
    /// IP en liste noire temporaire
    Blacklisted,
    /// IP bloquée définitivement
    Banned,
    /// IP de confiance (VPN/Proxy autorisé)
    Trusted,
}

/// Réputation d'un utilisateur
#[derive(Debug, Clone)]
pub struct UserReputation {
    pub score: f32, // 0.0 - 1.0
    pub level: ReputationLevel,
    pub violations_today: u32,
    pub positive_actions: u32,
    pub last_violation: Option<Instant>,
}

/// Niveau de réputation
#[derive(Debug, Clone, PartialEq)]
pub enum ReputationLevel {
    NewUser,    // Nouvels utilisateurs (restrictions strictes)
    Normal,     // Utilisateurs normaux
    Trusted,    // Utilisateurs de confiance
    VIP,        // Utilisateurs VIP (modérateurs, abonnés)
    System,     // Comptes système (bots officiels)
}

/// Limites quotidiennes
#[derive(Debug, Clone)]
pub struct DailyLimits {
    pub messages_sent: u32,
    pub max_messages: u32,
    pub files_uploaded: u32,
    pub max_files: u32,
    pub reset_time: Instant,
}

/// Niveau de modération d'un canal
#[derive(Debug, Clone, PartialEq)]
pub enum ModerationLevel {
    Low,      // Modération allégée
    Normal,   // Modération standard
    High,     // Modération stricte
    Lockdown, // Canal en verrouillage
}

/// Pattern d'attaque détecté
#[derive(Debug, Clone)]
pub struct AttackPattern {
    pub pattern_id: String,
    pub pattern_type: AttackType,
    pub source_ips: Vec<IpAddr>,
    pub detection_time: Instant,
    pub severity: f32,
    pub requests_count: u32,
    pub geographic_spread: f32,
    pub user_agents: Vec<String>,
}

/// Types d'attaques détectées
#[derive(Debug, Clone, PartialEq)]
pub enum AttackType {
    /// Attaque DDoS classique
    DDoS,
    /// Spam de messages
    MessageSpam,
    /// Brute force sur l'authentification
    BruteForce,
    /// Scraping de données
    Scraping,
    /// Attaque par déni de service applicatif
    SlowLoris,
    /// Comportement suspect automatisé
    BotActivity,
}

/// Entrée de liste noire
#[derive(Debug, Clone)]
pub struct BlacklistEntry {
    pub ip: IpAddr,
    pub reason: String,
    pub blocked_at: Instant,
    pub expires_at: Option<Instant>,
    pub violation_count: u32,
    pub auto_generated: bool,
}

/// Événement de requête pour la détection de patterns
#[derive(Debug, Clone)]
pub struct RequestEvent {
    pub timestamp: Instant,
    pub request_type: String,
    pub path: String,
    pub user_agent: Option<String>,
    pub response_time: Duration,
    pub status_code: u16,
}

/// Configuration du rate limiting
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    /// Messages par minute par utilisateur
    pub messages_per_minute: u32,
    /// Connexions par heure par IP
    pub connections_per_hour: u32,
    /// Tentatives d'auth par IP
    pub auth_attempts_per_minute: u32,
    /// Taille maximale des buckets
    pub max_bucket_capacity: u32,
    /// Seuil de détection d'attaque
    pub attack_detection_threshold: f32,
    /// Durée de blacklist automatique
    pub auto_blacklist_duration: Duration,
    /// Activar la géolocalisation
    pub enable_geolocation: bool,
    /// IPs de confiance (CDN, proxies autorisés)
    pub trusted_ips: Vec<IpAddr>,
}

/// Métriques du rate limiting
#[derive(Debug, Default)]
pub struct RateLimitMetrics {
    pub requests_processed: Arc<std::sync::atomic::AtomicU64>,
    pub requests_blocked: Arc<std::sync::atomic::AtomicU64>,
    pub attacks_detected: Arc<std::sync::atomic::AtomicU32>,
    pub false_positives: Arc<std::sync::atomic::AtomicU32>,
    pub avg_response_time: Arc<Mutex<Duration>>,
}

/// Résultat d'une vérification de rate limit
#[derive(Debug, Clone)]
pub struct RateLimitResult {
    pub allowed: bool,
    pub reason: Option<String>,
    pub retry_after: Option<Duration>,
    pub remaining_tokens: u32,
    pub burst_remaining: u32,
    pub reputation_impact: f32,
}

impl AdvancedRateLimiter {
    /// Crée un nouveau rate limiter avancé
    pub fn new(config: RateLimitConfig) -> Self {
        Self {
            ip_limiters: Arc::new(DashMap::new()),
            user_limiters: Arc::new(DashMap::new()),
            channel_limiters: Arc::new(DashMap::new()),
            attack_patterns: Arc::new(DashMap::new()),
            blacklist: Arc::new(DashMap::new()),
            config: Arc::new(RwLock::new(config)),
            metrics: Arc::new(RateLimitMetrics::default()),
        }
    }
    
    /// Vérifie si une requête est autorisée (point d'entrée principal)
    pub async fn check_rate_limit(
        &self,
        ip: IpAddr,
        user_id: Option<i64>,
        channel_id: Option<String>,
        _limit_type: LimitType,
        request_info: RequestInfo,
    ) -> Result<RateLimitResult> {
        let start_time = Instant::now();
        
        // Incrémenter les métriques
        self.metrics.requests_processed.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        
        // 1. Vérifier la liste noire d'abord
        if let Some(blacklist_entry) = self.blacklist.get(&ip) {
            if blacklist_entry.expires_at.map_or(true, |exp| exp > Instant::now()) {
                self.metrics.requests_blocked.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
                return Ok(RateLimitResult {
                    allowed: false,
                    reason: Some(format!("IP blacklisted: {}", blacklist_entry.reason)),
                    retry_after: blacklist_entry.expires_at.map(|exp| exp.duration_since(Instant::now())),
                    remaining_tokens: 0,
                    burst_remaining: 0,
                    reputation_impact: -0.1,
                });
            } else {
                // Entrée expirée, la supprimer
                self.blacklist.remove(&ip);
            }
        }
        
        // 2. Vérifier le rate limiting par IP
        let ip_result = self.check_ip_rate_limit(ip, &_limit_type, &request_info).await?;
        if !ip_result.allowed {
            return Ok(ip_result);
        }
        
        // 3. Vérifier le rate limiting par utilisateur si applicable
        if let Some(uid) = user_id {
            let user_result = self.check_user_rate_limit(uid, &_limit_type).await?;
            if !user_result.allowed {
                return Ok(user_result);
            }
        }
        
        // 4. Vérifier le rate limiting par canal si applicable
        if let Some(cid) = channel_id {
            let channel_result = self.check_channel_rate_limit(&cid, &_limit_type).await?;
            if !channel_result.allowed {
                return Ok(channel_result);
            }
        }
        
        // 5. Analyser les patterns d'attaque
        self.analyze_request_pattern(ip, &request_info).await?;
        
        // 6. Mettre à jour les métriques de performance
        let elapsed = start_time.elapsed();
        *self.metrics.avg_response_time.lock() = elapsed;
        
        Ok(RateLimitResult {
            allowed: true,
            reason: None,
            retry_after: None,
            remaining_tokens: ip_result.remaining_tokens,
            burst_remaining: ip_result.burst_remaining,
            reputation_impact: 0.0,
        })
    }
    
    /// Vérifie le rate limiting par IP
    async fn check_ip_rate_limit(
        &self,
        ip: IpAddr,
        _limit_type: &LimitType,
        request_info: &RequestInfo,
    ) -> Result<RateLimitResult> {
        let config = self.config.read().await;
        
        // Récupérer ou créer le limiteur IP
        let mut ip_limiter = self.ip_limiters.entry(ip).or_insert_with(|| {
            IpRateLimiter::new(ip, &config)
        });
        
        // Vérifier le statut de l'IP
        match ip_limiter.status {
            IpStatus::Banned => {
                return Ok(RateLimitResult {
                    allowed: false,
                    reason: Some("IP permanently banned".to_string()),
                    retry_after: None,
                    remaining_tokens: 0,
                    burst_remaining: 0,
                    reputation_impact: -0.2,
                });
            }
            IpStatus::Blacklisted => {
                return Ok(RateLimitResult {
                    allowed: false,
                    reason: Some("IP temporarily blacklisted".to_string()),
                    retry_after: Some(Duration::from_secs(300)), // 5 minutes
                    remaining_tokens: 0,
                    burst_remaining: 0,
                    reputation_impact: -0.1,
                });
            }
            _ => {}
        }
        
        // Appliquer le rate limiting avec token bucket
        let remaining_tokens = {
            let bucket = ip_limiter.buckets.get_mut(_limit_type).unwrap();
            bucket.refill();
            
            if bucket.tokens > 0 {
                bucket.tokens -= 1;
                bucket.tokens
            } else {
                0
            }
        };
        
        if remaining_tokens > 0 {
            ip_limiter.last_activity = Instant::now();
            
            // Enregistrer l'événement pour l'analyse de patterns
            ip_limiter.request_patterns.push_back(RequestEvent {
                timestamp: Instant::now(),
                request_type: format!("{:?}", _limit_type),
                path: request_info.path.clone(),
                user_agent: request_info.user_agent.clone(),
                response_time: Duration::from_millis(0),
                status_code: 200,
            });
            
            // Garder seulement les 100 derniers événements
            if ip_limiter.request_patterns.len() > 100 {
                ip_limiter.request_patterns.pop_front();
            }
            
            Ok(RateLimitResult {
                allowed: true,
                reason: None,
                retry_after: None,
                remaining_tokens,
                burst_remaining: remaining_tokens,
                reputation_impact: 0.0,
            })
        } else {
            // Rate limit dépassé
            ip_limiter.violation_count += 1;
            
            // Escalade automatique si trop de violations
            if ip_limiter.violation_count >= 5 {
                ip_limiter.status = IpStatus::Suspicious;
            }
            if ip_limiter.violation_count >= 10 {
                self.auto_blacklist_ip(ip, "Too many violations".to_string()).await?;
            }
            
            Ok(RateLimitResult {
                allowed: false,
                reason: Some(format!("Rate limit exceeded for {:?}", _limit_type)),
                retry_after: Some(Duration::from_secs(60)),
                remaining_tokens: 0,
                burst_remaining: 0,
                reputation_impact: 0.0,
            })
        }
    }
    
    /// Vérifie le rate limiting par utilisateur
    async fn check_user_rate_limit(&self, user_id: i64, _limit_type: &LimitType) -> Result<RateLimitResult> {
        let config = self.config.read().await;
        
        let mut user_limiter = self.user_limiters.entry(user_id).or_insert_with(|| {
            UserRateLimiter::new(user_id, &config)
        });
        
        // Vérifier la réputation d'abord
        let capacity_multiplier = match user_limiter.reputation.level {
            ReputationLevel::NewUser => 0.5,
            ReputationLevel::Normal => 1.0,
            ReputationLevel::Trusted => 1.5,
            ReputationLevel::VIP => 2.0,
            ReputationLevel::System => 5.0,
        };

        // Puis accéder au bucket avec la capacité ajustée
        let remaining_tokens = {
            let bucket = user_limiter.buckets.get_mut(_limit_type).unwrap();
            bucket.capacity = (bucket.capacity as f32 * capacity_multiplier) as u32;
            bucket.refill();
            
            if bucket.tokens > 0 {
                bucket.tokens -= 1;
                bucket.tokens
            } else {
                0
            }
        };
        
        if remaining_tokens > 0 {
            user_limiter.last_activity = Instant::now();
            
            Ok(RateLimitResult {
                allowed: true,
                reason: None,
                retry_after: None,
                remaining_tokens,
                burst_remaining: remaining_tokens,
                reputation_impact: 0.0,
            })
        } else {
            user_limiter.violation_count += 1;
            user_limiter.reputation.violations_today += 1;
            user_limiter.reputation.score = (user_limiter.reputation.score - 0.05).max(0.0);
            
            Ok(RateLimitResult {
                allowed: false,
                reason: Some(format!("User rate limit exceeded for {:?}", _limit_type)),
                retry_after: Some(Duration::from_secs(30)),
                remaining_tokens: 0,
                burst_remaining: 0,
                reputation_impact: 0.0,
            })
        }
    }
    
    /// Vérifie le rate limiting par canal
    async fn check_channel_rate_limit(&self, channel_id: &str, _limit_type: &LimitType) -> Result<RateLimitResult> {
        tracing::debug!(?_limit_type, "Vérification du rate limit canal");
        let config = self.config.read().await;
        
        let mut channel_limiter = self.channel_limiters.entry(channel_id.to_string()).or_insert_with(|| {
            ChannelRateLimiter::new(channel_id.to_string(), &config)
        });
        
        // Appliquer la modération selon le niveau du canal et le type de limite
        let rate_multiplier = match (channel_limiter.moderation_level.clone(), _limit_type) {
            (ModerationLevel::Low, LimitType::MessagesPerMinute) => 1.5,
            (ModerationLevel::Normal, LimitType::MessagesPerMinute) => 1.0,
            (ModerationLevel::High, LimitType::MessagesPerMinute) => 0.5,
            (ModerationLevel::Lockdown, LimitType::MessagesPerMinute) => 0.1,
            (_, _) => 1.0, // Valeur par défaut pour autres types
        };
        
        channel_limiter.message_bucket.refill();
        let tokens_needed = (1.0 / rate_multiplier) as u32;
        
        if channel_limiter.message_bucket.tokens >= tokens_needed {
            channel_limiter.message_bucket.tokens -= tokens_needed;
            channel_limiter.last_activity = Instant::now();
            
            Ok(RateLimitResult {
                allowed: true,
                reason: None,
                retry_after: None,
                remaining_tokens: channel_limiter.message_bucket.tokens,
                burst_remaining: channel_limiter.message_bucket.burst_allowance,
                reputation_impact: 0.0,
            })
        } else {
            Ok(RateLimitResult {
                allowed: false,
                reason: Some(format!("Channel rate limit exceeded (moderation: {:?})", channel_limiter.moderation_level)),
                retry_after: Some(Duration::from_secs(10)),
                remaining_tokens: 0,
                burst_remaining: 0,
                reputation_impact: 0.0,
            })
        }
    }
    
    /// Analyse les patterns de requêtes pour détecter les attaques
    async fn analyze_request_pattern(&self, ip: IpAddr, request_info: &RequestInfo) -> Result<()> {
        // Récupérer l'historique des requêtes pour cette IP
        if let Some(ip_limiter) = self.ip_limiters.get(&ip) {
            let recent_requests: Vec<_> = ip_limiter.request_patterns.iter()
                .filter(|event| event.timestamp.elapsed() < Duration::from_secs(60))
                .collect();
            
            // Détecter différents types d'attaques
            
            // 1. DDoS - Trop de requêtes dans un court laps de temps
            if recent_requests.len() > 100 {
                self.detect_ddos_attack(ip, &recent_requests).await?;
            }
            
            // 2. Brute Force - Tentatives répétées sur l'authentification
            if request_info.path.contains("/auth") || request_info.path.contains("/login") {
                let auth_attempts = recent_requests.iter()
                    .filter(|event| event.path.contains("/auth"))
                    .count();
                
                if auth_attempts > 10 {
                    self.detect_brute_force_attack(ip).await?;
                }
            }
            
            // 3. Bot Activity - Patterns suspects (même User-Agent, timing régulier)
            if let Some(user_agent) = &request_info.user_agent {
                let same_ua_count = recent_requests.iter()
                    .filter(|event| event.user_agent.as_ref() == Some(user_agent))
                    .count();
                
                if same_ua_count > 50 && self.is_suspicious_user_agent(user_agent) {
                    self.detect_bot_activity(ip, user_agent.clone()).await?;
                }
            }
        }
        
        Ok(())
    }
    
    /// Détecte une attaque DDoS
    async fn detect_ddos_attack(&self, ip: IpAddr, recent_requests: &[&RequestEvent]) -> Result<()> {
        let pattern = AttackPattern {
            pattern_id: format!("ddos_{}_{}",ip, Instant::now().elapsed().as_secs()),
            pattern_type: AttackType::DDoS,
            source_ips: vec![ip],
            detection_time: Instant::now(),
            severity: 0.9,
            requests_count: recent_requests.len() as u32,
            geographic_spread: 0.0, // Single IP
            user_agents: recent_requests.iter()
                .filter_map(|event| event.user_agent.clone())
                .collect::<std::collections::HashSet<_>>()
                .into_iter()
                .collect(),
        };
        
        self.attack_patterns.insert(pattern.pattern_id.clone(), pattern);
        self.auto_blacklist_ip(ip, "DDoS attack detected".to_string()).await?;
        
        self.metrics.attacks_detected.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        tracing::warn!("DDoS attack detected from IP: {}", ip);
        
        Ok(())
    }
    
    /// Détecte une attaque de brute force
    async fn detect_brute_force_attack(&self, ip: IpAddr) -> Result<()> {
        let pattern = AttackPattern {
            pattern_id: format!("bruteforce_{}_{}", ip, Instant::now().elapsed().as_secs()),
            pattern_type: AttackType::BruteForce,
            source_ips: vec![ip],
            detection_time: Instant::now(),
            severity: 0.8,
            requests_count: 10,
            geographic_spread: 0.0,
            user_agents: vec![],
        };
        
        self.attack_patterns.insert(pattern.pattern_id.clone(), pattern);
        self.auto_blacklist_ip(ip, "Brute force attack detected".to_string()).await?;
        
        tracing::warn!("Brute force attack detected from IP: {}", ip);
        Ok(())
    }
    
    /// Détecte une activité de bot
    async fn detect_bot_activity(&self, ip: IpAddr, user_agent: String) -> Result<()> {
        let pattern = AttackPattern {
            pattern_id: format!("bot_{}_{}", ip, Instant::now().elapsed().as_secs()),
            pattern_type: AttackType::BotActivity,
            source_ips: vec![ip],
            detection_time: Instant::now(),
            severity: 0.6,
            requests_count: 50,
            geographic_spread: 0.0,
            user_agents: vec![user_agent],
        };
        
        self.attack_patterns.insert(pattern.pattern_id.clone(), pattern);
        
        // Marquer l'IP comme suspecte plutôt que de la blacklister immédiatement
        if let Some(mut ip_limiter) = self.ip_limiters.get_mut(&ip) {
            ip_limiter.status = IpStatus::Suspicious;
            ip_limiter.trust_score = (ip_limiter.trust_score - 0.3).max(0.0);
        }
        
        tracing::info!("Bot activity detected from IP: {}", ip);
        Ok(())
    }
    
    /// Vérifie si un User-Agent est suspect
    fn is_suspicious_user_agent(&self, user_agent: &str) -> bool {
        let suspicious_patterns = [
            "bot", "crawler", "spider", "scraper", "curl", "wget", "python", "java", 
            "headless", "selenium", "phantom", "automated"
        ];
        
        let ua_lower = user_agent.to_lowercase();
        suspicious_patterns.iter().any(|pattern| ua_lower.contains(pattern))
    }
    
    /// Ajoute automatiquement une IP à la liste noire
    async fn auto_blacklist_ip(&self, ip: IpAddr, reason: String) -> Result<()> {
        let reason_clone = reason.clone();
        let config = self.config.read().await;
        
        let blacklist_entry = BlacklistEntry {
            ip,
            reason,
            blocked_at: Instant::now(),
            expires_at: Some(Instant::now() + config.auto_blacklist_duration),
            violation_count: 1,
            auto_generated: true,
        };
        
        self.blacklist.insert(ip, blacklist_entry);
        self.metrics.attacks_detected.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
        
        tracing::warn!("IP {} automatically blacklisted: {}", ip, reason_clone);
        Ok(())
    }
    
    /// Nettoie les entrées expirées
    pub async fn cleanup_expired_entries(&self) {
        let now = Instant::now();
        
        // Nettoyer les blacklists expirées
        self.blacklist.retain(|_, entry| {
            entry.expires_at.map_or(true, |exp| exp > now)
        });
        
        // Nettoyer les limiteurs inactifs (plus de 1 heure)
        self.ip_limiters.retain(|_, limiter| {
            now.duration_since(limiter.last_activity) < Duration::from_secs(3600)
        });
        
        self.user_limiters.retain(|_, limiter| {
            now.duration_since(limiter.last_activity) < Duration::from_secs(3600)
        });
        
        self.channel_limiters.retain(|_, limiter| {
            now.duration_since(limiter.last_activity) < Duration::from_secs(3600)
        });
    }
}

/// Informations sur une requête
#[derive(Debug, Clone)]
pub struct RequestInfo {
    pub path: String,
    pub user_agent: Option<String>,
    pub method: String,
    pub content_length: Option<usize>,
}

impl TokenBucket {
    pub fn new(capacity: u32, refill_rate: f32, burst_allowance: u32) -> Self {
        Self {
            capacity,
            tokens: capacity,
            refill_rate,
            last_refill: Instant::now(),
            burst_allowance,
        }
    }
    
    pub fn refill(&mut self) {
        let now = Instant::now();
        let elapsed = now.duration_since(self.last_refill).as_secs_f32();
        let tokens_to_add = (elapsed * self.refill_rate) as u32;
        
        if tokens_to_add > 0 {
            self.tokens = (self.tokens + tokens_to_add).min(self.capacity);
            self.last_refill = now;
        }
    }
}

impl IpRateLimiter {
    pub fn new(ip: IpAddr, config: &RateLimitConfig) -> Self {
        let mut buckets = HashMap::new();
        
        // Créer les buckets pour différents types de limitations
        buckets.insert(LimitType::MessagesPerMinute, TokenBucket::new(config.messages_per_minute, config.messages_per_minute as f32 / 60.0, 10));
        buckets.insert(LimitType::ConnectionsPerHour, TokenBucket::new(config.connections_per_hour, config.connections_per_hour as f32 / 3600.0, 5));
        buckets.insert(LimitType::AuthAttempts, TokenBucket::new(config.auth_attempts_per_minute, config.auth_attempts_per_minute as f32 / 60.0, 2));
        buckets.insert(LimitType::ApiRequests, TokenBucket::new(1000, 16.67, 50)); // 1000/min
        buckets.insert(LimitType::FileUploads, TokenBucket::new(10, 0.17, 2)); // 10/min
        
        Self {
            ip,
            buckets,
            last_activity: Instant::now(),
            violation_count: 0,
            trust_score: 0.5, // Score neutre initial
            request_patterns: VecDeque::new(),
            status: IpStatus::Normal,
        }
    }
}

impl UserRateLimiter {
    pub fn new(user_id: i64, config: &RateLimitConfig) -> Self {
        let mut buckets = HashMap::new();
        
        buckets.insert(LimitType::MessagesPerMinute, TokenBucket::new(config.messages_per_minute, config.messages_per_minute as f32 / 60.0, 5));
        buckets.insert(LimitType::FileUploads, TokenBucket::new(20, 0.33, 3)); // 20/min
        buckets.insert(LimitType::ChannelCreation, TokenBucket::new(5, 0.083, 1)); // 5/min
        buckets.insert(LimitType::Invitations, TokenBucket::new(10, 0.17, 2)); // 10/min
        buckets.insert(LimitType::Reactions, TokenBucket::new(60, 1.0, 10)); // 60/min
        
        Self {
            user_id,
            buckets,
            last_activity: Instant::now(),
            violation_count: 0,
            reputation: UserReputation {
                score: 0.5,
                level: ReputationLevel::NewUser,
                violations_today: 0,
                positive_actions: 0,
                last_violation: None,
            },
            daily_limits: DailyLimits {
                messages_sent: 0,
                max_messages: 1000,
                files_uploaded: 0,
                max_files: 50,
                reset_time: Instant::now() + Duration::from_secs(86400), // 24h
            },
        }
    }
}

impl ChannelRateLimiter {
    pub fn new(channel_id: String, config: &RateLimitConfig) -> Self {
        Self {
            channel_id,
            message_bucket: TokenBucket::new(config.messages_per_minute * 10, (config.messages_per_minute * 10) as f32 / 60.0, 20),
            concurrent_users: 0,
            last_activity: Instant::now(),
            spam_threshold: 0.7,
            moderation_level: ModerationLevel::Normal,
        }
    }
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            messages_per_minute: 30,
            connections_per_hour: 100,
            auth_attempts_per_minute: 5,
            max_bucket_capacity: 1000,
            attack_detection_threshold: 0.8,
            auto_blacklist_duration: Duration::from_secs(900), // 15 minutes
            enable_geolocation: true,
            trusted_ips: vec![],
        }
    }
} 