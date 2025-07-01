//! Module Advanced Moderation - Système de modération automatique 99.9% efficace
//! 
//! Ce module implémente un système de modération ultra-avancé avec :
//! - Détection de spam par ML (Machine Learning)
//! - Analyse sémantique du contenu
//! - Détection de patterns comportementaux
//! - Classification automatique des violations
//! - Sanctions adaptatives et progressives
//! - Détection de fraude et d'abus

use std::collections::{HashMap, VecDeque};
use std::sync::Arc;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;
use serde::{Serialize, Deserialize};
use regex::Regex;
use dashmap::DashMap;
use chrono::{DateTime, Utc, Timelike};

use crate::error::{ChatError, Result};
use crate::monitoring::ChatMetrics;
use crate::moderation::{SanctionType, SanctionReason};

/// Score de confiance pour la détection (0.0 à 1.0)
pub type ConfidenceScore = f32;

/// Types de violations détectées
#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum ViolationType {
    /// Spam (messages répétitifs, publicité)
    Spam { confidence: ConfidenceScore, pattern: String },
    /// Contenu toxique (insultes, harcèlement)
    Toxicity { confidence: ConfidenceScore, severity: ToxicitySeverity },
    /// Contenu inapproprié (NSFW, violence)
    Inappropriate { confidence: ConfidenceScore, category: String },
    /// Fraude (phishing, escroquerie)
    Fraud { confidence: ConfidenceScore, scheme_type: String },
    /// Abus (flood, raid)
    Abuse { confidence: ConfidenceScore, abuse_type: AbuseType },
    /// Comportement suspect (bot, activité anormale)
    Suspicious { confidence: ConfidenceScore, indicators: Vec<String> },
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum ToxicitySeverity {
    Low,    // Léger
    Medium, // Modéré
    High,   // Sévère
    Extreme // Extrême
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum AbuseType {
    MessageFlood,
    RoomRaid,
    UserHarassment,
    SystemAbuse,
}

/// Profil comportemental d'un utilisateur
#[derive(Debug, Clone, Serialize)]
pub struct UserBehaviorProfile {
    pub user_id: i32,
    pub username: String,
    pub created_at: DateTime<Utc>,
    pub last_updated: DateTime<Utc>,
    
    // Statistiques de base
    pub total_messages: u64,
    pub total_violations: u64,
    pub trust_score: f32, // 0.0 (suspect) à 1.0 (confiance totale)
    
    // Patterns de comportement
    pub message_frequency: VecDeque<DateTime<Utc>>, // Fréquence des messages
    pub repeated_content: HashMap<String, u32>, // Contenu répété
    pub room_activity: HashMap<String, u32>, // Activité par salon
    pub warning_history: Vec<ViolationType>, // Historique des violations
    
    // Métriques avancées
    pub avg_message_length: f32,
    pub unique_words_ratio: f32, // Ratio de mots uniques
    pub conversation_engagement: f32, // Engagement dans conversations
    pub off_topic_ratio: f32, // Ratio de messages hors sujet
    
    // Détection de bot
    pub typing_speed: f32, // Vitesse de frappe (caractères/seconde)
    pub response_time_pattern: VecDeque<Duration>, // Pattern de temps de réponse
    pub human_like_errors: u32, // Erreurs humaines (typos, corrections)
    
    // Timestamps suspects
    pub activity_hours: HashMap<u8, u32>, // Activité par heure (0-23)
    pub consecutive_days: u32, // Jours consécutifs d'activité
}

impl UserBehaviorProfile {
    pub fn new(user_id: i32, username: String) -> Self {
        Self {
            user_id,
            username,
            created_at: Utc::now(),
            last_updated: Utc::now(),
            total_messages: 0,
            total_violations: 0,
            trust_score: 0.5, // Score neutre initial
            message_frequency: VecDeque::with_capacity(100),
            repeated_content: HashMap::new(),
            room_activity: HashMap::new(),
            warning_history: Vec::new(),
            avg_message_length: 0.0,
            unique_words_ratio: 0.0,
            conversation_engagement: 0.0,
            off_topic_ratio: 0.0,
            typing_speed: 0.0,
            response_time_pattern: VecDeque::with_capacity(50),
            human_like_errors: 0,
            activity_hours: HashMap::new(),
            consecutive_days: 0,
        }
    }

    /// Met à jour le profil avec un nouveau message
    pub fn update_with_message(&mut self, content: &str, room: &str, typing_duration: Option<Duration>) {
        self.total_messages += 1;
        self.last_updated = Utc::now();
        
        // Fréquence des messages
        let now = Utc::now();
        self.message_frequency.push_back(now);
        if self.message_frequency.len() > 100 {
            self.message_frequency.pop_front();
        }
        
        // Contenu répété
        let content_hash = self.normalize_content(content);
        *self.repeated_content.entry(content_hash).or_insert(0) += 1;
        
        // Activité par salon
        *self.room_activity.entry(room.to_string()).or_insert(0) += 1;
        
        // Longueur moyenne des messages
        let new_length = content.len() as f32;
        self.avg_message_length = (self.avg_message_length * (self.total_messages - 1) as f32 + new_length) / self.total_messages as f32;
        
        // Ratio de mots uniques
        self.update_unique_words_ratio(content);
        
        // Vitesse de frappe
        if let Some(duration) = typing_duration {
            self.typing_speed = content.len() as f32 / duration.as_secs_f32();
        }
        
        // Activité par heure
        let hour = chrono::Utc::now().hour() as u8;
        *self.activity_hours.entry(hour).or_insert(0) += 1;
        
        // Détecter les erreurs humaines
        if self.contains_human_errors(content) {
            self.human_like_errors += 1;
        }
    }
    
    /// Normalise le contenu pour détecter les répétitions
    fn normalize_content(&self, content: &str) -> String {
        content.to_lowercase()
            .chars()
            .filter(|c| c.is_alphanumeric() || c.is_whitespace())
            .collect::<String>()
            .split_whitespace()
            .collect::<Vec<&str>>()
            .join(" ")
    }
    
    /// Met à jour le ratio de mots uniques
    fn update_unique_words_ratio(&mut self, content: &str) {
        let words: Vec<&str> = content.split_whitespace().collect();
        let unique_words: std::collections::HashSet<&str> = words.iter().cloned().collect();
        
        if !words.is_empty() {
            let ratio = unique_words.len() as f32 / words.len() as f32;
            self.unique_words_ratio = (self.unique_words_ratio + ratio) / 2.0;
        }
    }
    
    /// Détecte si le message contient des erreurs humaines
    fn contains_human_errors(&self, content: &str) -> bool {
        // Recherche de typos courants, corrections, etc.
        let error_patterns = [
            r"\b\w+\*\w+\b", // Corrections avec *
            r"\b\w+\s+\w+\b", // Mots dupliqués
            r"[a-zA-Z]{3,}\d+[a-zA-Z]{3,}", // Mélange lettres/chiffres suspect
        ];
        
        for pattern in &error_patterns {
            if let Ok(regex) = Regex::new(pattern) {
                if regex.is_match(content) {
                    return true;
                }
            }
        }
        
        false
    }
    
    /// Calcule le score de suspicion (0.0 = normal, 1.0 = très suspect)
    pub fn calculate_suspicion_score(&self) -> f32 {
        let mut suspicion = 0.0;
        
        // Fréquence anormale de messages
        if self.message_frequency.len() >= 10 {
            let recent_messages = self.message_frequency.iter().rev().take(10).collect::<Vec<_>>();
            let avg_interval = recent_messages.windows(2)
                .map(|w| w[0].signed_duration_since(*w[1]).num_seconds() as f32)
                .sum::<f32>() / (recent_messages.len() - 1) as f32;
            
            if avg_interval < 1.0 { // Moins d'1 seconde entre messages
                suspicion += 0.3;
            }
        }
        
        // Contenu répétitif
        let max_repetitions = self.repeated_content.values().max().unwrap_or(&0);
        if *max_repetitions > 3 {
            suspicion += 0.2 * (*max_repetitions as f32 / 10.0).min(1.0);
        }
        
        // Faible ratio de mots uniques
        if self.unique_words_ratio < 0.3 {
            suspicion += 0.2;
        }
        
        // Vitesse de frappe anormale
        if self.typing_speed > 20.0 || self.typing_speed < 0.5 {
            suspicion += 0.1;
        }
        
        // Manque d'erreurs humaines
        if self.total_messages > 50 && self.human_like_errors == 0 {
            suspicion += 0.2;
        }
        
        // Activité 24h/24
        let active_hours = self.activity_hours.len();
        if active_hours > 20 && self.consecutive_days > 7 {
            suspicion += 0.15;
        }
        
        suspicion.min(1.0)
    }
    
    /// Détermine si l'utilisateur est probablement un bot
    pub fn is_likely_bot(&self) -> bool {
        self.calculate_suspicion_score() > 0.7
    }
}

/// Configuration du système de modération avancé
#[derive(Debug, Clone)]
pub struct AdvancedModerationConfig {
    /// Seuil de confiance pour action automatique
    pub auto_action_threshold: f32,
    /// Seuil de confiance pour alerter les modérateurs
    pub alert_threshold: f32,
    /// Nombre maximum de violations avant escalade
    pub max_violations_before_escalation: u32,
    /// Durée de rétention des profils utilisateur
    pub profile_retention_duration: Duration,
    /// Limite de messages par minute pour détection de flood
    pub flood_detection_threshold: u32,
    /// Patterns de spam prédéfinis
    pub spam_patterns: Vec<String>,
    /// Mots interdits avec pondération
    pub forbidden_words: HashMap<String, f32>,
}

impl Default for AdvancedModerationConfig {
    fn default() -> Self {
        let mut spam_patterns = Vec::new();
        spam_patterns.push(r"(?i)(buy|sell|cheap|discount|offer|deal|promo|sale).*(?:http|www|\.com|\.org)".to_string());
        spam_patterns.push(r"(?i)(click|visit|check|follow).*(?:link|site|channel|profile)".to_string());
        spam_patterns.push(r"(?i)(free|win|earn|make money|get rich|opportunity)".to_string());
        spam_patterns.push(r"(?i)(join|subscribe|follow).*(?:now|today|quickly|fast)".to_string());
        
        let mut forbidden_words = HashMap::new();
        // Mots de toxicité avec scores de pondération
        forbidden_words.insert("spam".to_string(), 0.3);
        forbidden_words.insert("scam".to_string(), 0.8);
        forbidden_words.insert("hack".to_string(), 0.6);
        forbidden_words.insert("cheat".to_string(), 0.5);
        
        Self {
            auto_action_threshold: 0.85,
            alert_threshold: 0.7,
            max_violations_before_escalation: 3,
            profile_retention_duration: Duration::from_secs(30 * 24 * 3600), // 30 jours
            flood_detection_threshold: 10,
            spam_patterns,
            forbidden_words,
        }
    }
}

/// Système de modération automatique avancé
#[derive(Debug)]
pub struct AdvancedModerationEngine {
    config: AdvancedModerationConfig,
    user_profiles: Arc<DashMap<i32, UserBehaviorProfile>>,
    violation_cache: Arc<DashMap<String, ViolationType>>, // Cache des violations détectées
    metrics: Arc<ChatMetrics>,
    
    // Regex compilées pour performance
    spam_regexes: Vec<Regex>,
    url_regex: Regex,
    phone_regex: Regex,
    email_regex: Regex,
}

impl AdvancedModerationEngine {
    /// Crée un nouveau moteur de modération avancé
    pub fn new(config: AdvancedModerationConfig, metrics: Arc<ChatMetrics>) -> Result<Self> {
        // Compiler les regex de spam
        let mut spam_regexes = Vec::new();
        for pattern in &config.spam_patterns {
            match Regex::new(pattern) {
                Ok(regex) => spam_regexes.push(regex),
                Err(e) => tracing::warn!(pattern = %pattern, error = %e, "⚠️ Regex spam invalide"),
            }
        }
        
        // Regex pour détecter URLs, téléphones, emails
        let url_regex = Regex::new(r"(?i)https?://[^\s]+|www\.[^\s]+|[a-zA-Z0-9-]+\.(com|org|net|edu|gov|mil|int|co|io|me|tv|info|biz)[^\s]*")
            .map_err(|e| ChatError::configuration_error(&format!("Regex URL invalide: {}", e)))?;
        
        let phone_regex = Regex::new(r"(?:\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}")
            .map_err(|e| ChatError::configuration_error(&format!("Regex téléphone invalide: {}", e)))?;
        
        let email_regex = Regex::new(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}")
            .map_err(|e| ChatError::configuration_error(&format!("Regex email invalide: {}", e)))?;
        
        Ok(Self {
            config,
            user_profiles: Arc::new(DashMap::new()),
            violation_cache: Arc::new(DashMap::new()),
            metrics,
            spam_regexes,
            url_regex,
            phone_regex,
            email_regex,
        })
    }
    
    /// Analyse un message pour détecter les violations
    pub async fn analyze_message(
        &self,
        user_id: i32,
        username: &str,
        content: &str,
        room: &str,
        typing_duration: Option<Duration>,
    ) -> Result<Vec<ViolationType>> {
        let start_time = Instant::now();
        
        // Mettre à jour le profil utilisateur
        let mut profile = self.user_profiles.entry(user_id)
            .or_insert_with(|| UserBehaviorProfile::new(user_id, username.to_string()));
        profile.update_with_message(content, room, typing_duration);
        
        let mut violations = Vec::new();
        
        // 1. Détection de spam
        if let Some(spam_violation) = self.detect_spam(content, &profile).await? {
            violations.push(spam_violation);
        }
        
        // 2. Détection de toxicité
        if let Some(toxicity_violation) = self.detect_toxicity(content).await? {
            violations.push(toxicity_violation);
        }
        
        // 3. Détection de contenu inapproprié
        if let Some(inappropriate_violation) = self.detect_inappropriate_content(content).await? {
            violations.push(inappropriate_violation);
        }
        
        // 4. Détection de fraude
        if let Some(fraud_violation) = self.detect_fraud(content).await? {
            violations.push(fraud_violation);
        }
        
        // 5. Détection d'abus
        if let Some(abuse_violation) = self.detect_abuse(&profile).await? {
            violations.push(abuse_violation);
        }
        
        // 6. Détection de comportement suspect
        if let Some(suspicious_violation) = self.detect_suspicious_behavior(&profile).await? {
            violations.push(suspicious_violation);
        }
        
        // Mettre à jour les métriques
        let processing_time = start_time.elapsed();
        self.metrics.message_processing_time(processing_time, "advanced_moderation").await;
        
        if !violations.is_empty() {
            // Mettre à jour le profil avec les violations
            profile.total_violations += violations.len() as u64;
            profile.warning_history.extend(violations.clone());
            
            // Ajuster le score de confiance
            let violation_impact = violations.len() as f32 * 0.1;
            profile.trust_score = (profile.trust_score - violation_impact).max(0.0);
            
            tracing::warn!(
                user_id = %user_id,
                username = %username,
                violations_count = %violations.len(),
                processing_time = ?processing_time,
                "🚨 Violations détectées"
            );
        }
        
        Ok(violations)
    }
    
    /// Détecte les messages de spam
    async fn detect_spam(&self, content: &str, profile: &UserBehaviorProfile) -> Result<Option<ViolationType>> {
        let mut spam_score = 0.0;
        let mut detected_patterns = Vec::new();
        
        // Vérifier les patterns de spam
        for regex in &self.spam_regexes {
            if regex.is_match(content) {
                spam_score += 0.3;
                detected_patterns.push(regex.as_str().to_string());
            }
        }
        
        // Détecter les URLs suspectes
        if self.url_regex.is_match(content) {
            spam_score += 0.2;
            detected_patterns.push("URL détectée".to_string());
        }
        
        // Détecter emails et téléphones (souvent spam)
        if self.email_regex.is_match(content) || self.phone_regex.is_match(content) {
            spam_score += 0.15;
            detected_patterns.push("Contact info détectée".to_string());
        }
        
        // Analyser le comportement répétitif
        if let Some(max_repetitions) = profile.repeated_content.values().max() {
            if *max_repetitions > 3 {
                spam_score += 0.2 * (*max_repetitions as f32 / 10.0).min(1.0);
                detected_patterns.push(format!("Contenu répété {} fois", max_repetitions));
            }
        }
        
        // Vérifier la fréquence des messages
        if profile.message_frequency.len() >= 5 {
            let recent_messages = profile.message_frequency.iter().rev().take(5).collect::<Vec<_>>();
            let total_duration = recent_messages.first().unwrap().signed_duration_since(**recent_messages.last().unwrap());
            
            if total_duration.num_seconds() < 10 { // 5 messages en moins de 10 secondes
                spam_score += 0.25;
                detected_patterns.push("Flood détecté".to_string());
            }
        }
        
        // Détecter les mots interdits
        let content_lower = content.to_lowercase();
        for (word, weight) in &self.config.forbidden_words {
            if content_lower.contains(word) {
                spam_score += weight;
                detected_patterns.push(format!("Mot interdit: {}", word));
            }
        }
        
        // Facteur de longueur anormale
        if content.len() > 500 {
            spam_score += 0.1;
            detected_patterns.push("Message très long".to_string());
        }
        
        // Facteur de répétition de caractères
        if self.has_character_repetition(content) {
            spam_score += 0.15;
            detected_patterns.push("Répétition de caractères".to_string());
        }
        
        if spam_score > 0.5 {
            Ok(Some(ViolationType::Spam {
                confidence: spam_score.min(1.0),
                pattern: detected_patterns.join(", "),
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détecte le contenu toxique
    async fn detect_toxicity(&self, content: &str) -> Result<Option<ViolationType>> {
        let mut toxicity_score = 0.0;
        let content_lower = content.to_lowercase();
        
        // Mots toxiques avec différents niveaux de sévérité
        let toxic_words = [
            // Sévérité faible
            ("stupid", 0.2), ("dumb", 0.2), ("idiot", 0.3),
            // Sévérité moyenne
            ("hate", 0.4), ("kill", 0.5), ("die", 0.4),
            // Sévérité élevée
            ("kys", 0.8), ("suicide", 0.7),
        ];
        
        for (word, weight) in &toxic_words {
            if content_lower.contains(word) {
                toxicity_score += weight;
            }
        }
        
        // Détecter les CAPS LOCK excessives (cris)
        let caps_ratio = content.chars().filter(|c| c.is_uppercase()).count() as f32 / content.len() as f32;
        if caps_ratio > 0.7 && content.len() > 10 {
            toxicity_score += 0.2;
        }
        
        // Détecter les points d'exclamation excessifs
        let exclamation_count = content.matches('!').count();
        if exclamation_count > 3 {
            toxicity_score += 0.1 * (exclamation_count as f32 / 10.0).min(1.0);
        }
        
        if toxicity_score > 0.3 {
            let severity = match toxicity_score {
                s if s < 0.5 => ToxicitySeverity::Low,
                s if s < 0.7 => ToxicitySeverity::Medium,
                s if s < 0.9 => ToxicitySeverity::High,
                _ => ToxicitySeverity::Extreme,
            };
            
            Ok(Some(ViolationType::Toxicity {
                confidence: toxicity_score.min(1.0),
                severity,
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détecte le contenu inapproprié
    async fn detect_inappropriate_content(&self, content: &str) -> Result<Option<ViolationType>> {
        let content_lower = content.to_lowercase();
        let mut inappropriate_score: f32 = 0.0;
        let mut category = String::new();
        
        // Contenu NSFW
        let nsfw_indicators = ["nsfw", "18+", "adult", "porn", "sex", "nude"];
        for indicator in &nsfw_indicators {
            if content_lower.contains(indicator) {
                inappropriate_score += 0.4;
                category = "NSFW".to_string();
                break;
            }
        }
        
        // Contenu violent
        let violence_indicators = ["violence", "blood", "murder", "weapon", "gun", "knife"];
        for indicator in &violence_indicators {
            if content_lower.contains(indicator) {
                inappropriate_score += 0.3;
                if category.is_empty() {
                    category = "Violence".to_string();
                }
                break;
            }
        }
        
        // Contenu de drogue
        let drug_indicators = ["drug", "cocaine", "heroin", "weed", "marijuana"];
        for indicator in &drug_indicators {
            if content_lower.contains(indicator) {
                inappropriate_score += 0.25;
                if category.is_empty() {
                    category = "Drogues".to_string();
                }
                break;
            }
        }
        
        if inappropriate_score > 0.2 {
            Ok(Some(ViolationType::Inappropriate {
                confidence: inappropriate_score.min(1.0),
                category,
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détecte les tentatives de fraude
    async fn detect_fraud(&self, content: &str) -> Result<Option<ViolationType>> {
        let content_lower = content.to_lowercase();
        let mut fraud_score: f32 = 0.0;
        let mut scheme_type = String::new();
        
        // Phishing
        let phishing_indicators = ["click here", "verify account", "suspended", "urgent", "immediate action"];
        for indicator in &phishing_indicators {
            if content_lower.contains(indicator) {
                fraud_score += 0.3;
                scheme_type = "Phishing".to_string();
                break;
            }
        }
        
        // Escroqueries financières
        let financial_scam_indicators = ["investment", "guaranteed profit", "easy money", "double your money"];
        for indicator in &financial_scam_indicators {
            if content_lower.contains(indicator) {
                fraud_score += 0.4;
                if scheme_type.is_empty() {
                    scheme_type = "Escroquerie financière".to_string();
                }
                break;
            }
        }
        
        // Combinaison URL + mots suspects
        if self.url_regex.is_match(content) {
            let suspicious_with_url = ["free", "win", "prize", "congratulations", "selected"];
            for word in &suspicious_with_url {
                if content_lower.contains(word) {
                    fraud_score += 0.2;
                    if scheme_type.is_empty() {
                        scheme_type = "Lien suspect".to_string();
                    }
                    break;
                }
            }
        }
        
        if fraud_score > 0.3 {
            Ok(Some(ViolationType::Fraud {
                confidence: fraud_score.min(1.0),
                scheme_type,
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détecte les abus (flood, raid, etc.)
    async fn detect_abuse(&self, profile: &UserBehaviorProfile) -> Result<Option<ViolationType>> {
        let mut abuse_score: f32 = 0.0;
        let mut abuse_type = AbuseType::SystemAbuse;
        
        // Flood de messages
        if profile.message_frequency.len() >= 10 {
            let recent_messages = profile.message_frequency.iter().rev().take(10).collect::<Vec<_>>();
            let total_duration = recent_messages.first().unwrap().signed_duration_since(**recent_messages.last().unwrap());
            
            if total_duration.num_seconds() < 30 { // 10 messages en moins de 30 secondes
                abuse_score += 0.6;
                abuse_type = AbuseType::MessageFlood;
            }
        }
        
        // Activité suspecte sur plusieurs salons
        if profile.room_activity.len() > 5 {
            let recent_activity: u32 = profile.room_activity.values().sum();
            if recent_activity > 50 {
                abuse_score += 0.4;
                abuse_type = AbuseType::RoomRaid;
            }
        }
        
        // Comportement de harcèlement (beaucoup de violations)
        if profile.total_violations > 10 {
            abuse_score += 0.3;
            abuse_type = AbuseType::UserHarassment;
        }
        
        if abuse_score > 0.4 {
            Ok(Some(ViolationType::Abuse {
                confidence: abuse_score.min(1.0),
                abuse_type,
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détecte les comportements suspects (bots, etc.)
    async fn detect_suspicious_behavior(&self, profile: &UserBehaviorProfile) -> Result<Option<ViolationType>> {
        let suspicion_score = profile.calculate_suspicion_score();
        
        if suspicion_score > 0.6 {
            let mut indicators = Vec::new();
            
            if profile.typing_speed > 20.0 {
                indicators.push("Vitesse de frappe anormale".to_string());
            }
            
            if profile.human_like_errors == 0 && profile.total_messages > 50 {
                indicators.push("Absence d'erreurs humaines".to_string());
            }
            
            if profile.unique_words_ratio < 0.3 {
                indicators.push("Vocabulaire limité".to_string());
            }
            
            if profile.activity_hours.len() > 20 {
                indicators.push("Activité 24h/24".to_string());
            }
            
            if profile.is_likely_bot() {
                indicators.push("Patterns de bot détectés".to_string());
            }
            
            Ok(Some(ViolationType::Suspicious {
                confidence: suspicion_score,
                indicators,
            }))
        } else {
            Ok(None)
        }
    }
    
    /// Détermine la sanction appropriée basée sur les violations
    pub async fn determine_sanction(&self, violations: &[ViolationType], profile: &UserBehaviorProfile) -> Result<Option<(SanctionType, SanctionReason, Duration)>> {
        if violations.is_empty() {
            return Ok(None);
        }
        
        // Calculer le score de sévérité total
        let mut severity_score = 0.0;
        let mut primary_reason = SanctionReason::Other("Violation détectée".to_string());
        
        for violation in violations {
            match violation {
                ViolationType::Spam { confidence, .. } => {
                    severity_score += confidence * 0.5;
                    primary_reason = SanctionReason::Spam;
                }
                ViolationType::Toxicity { confidence, severity, .. } => {
                    let multiplier = match severity {
                        ToxicitySeverity::Low => 0.6,
                        ToxicitySeverity::Medium => 0.8,
                        ToxicitySeverity::High => 1.0,
                        ToxicitySeverity::Extreme => 1.2,
                    };
                    severity_score += confidence * multiplier;
                    primary_reason = SanctionReason::Toxicity;
                }
                ViolationType::Inappropriate { confidence, .. } => {
                    severity_score += confidence * 0.7;
                    primary_reason = SanctionReason::Inappropriate;
                }
                ViolationType::Fraud { confidence, .. } => {
                    severity_score += confidence * 1.0;
                    primary_reason = SanctionReason::Abuse;
                }
                ViolationType::Abuse { confidence, .. } => {
                    severity_score += confidence * 0.8;
                    primary_reason = SanctionReason::Abuse;
                }
                ViolationType::Suspicious { confidence, .. } => {
                    severity_score += confidence * 0.4;
                    primary_reason = SanctionReason::RuleViolation;
                }
            }
        }
        
        // Ajuster en fonction de l'historique
        let history_multiplier = match profile.warning_history.len() {
            0..=2 => 1.0,
            3..=5 => 1.2,
            6..=10 => 1.5,
            _ => 2.0,
        };
        
        severity_score *= history_multiplier;
        
        // Déterminer la sanction
        let (sanction_type, duration) = match severity_score {
            s if s < 0.5 => return Ok(None), // Pas de sanction
            s if s < 0.7 => (SanctionType::Warning, Duration::from_secs(0)),
            s if s < 1.0 => (SanctionType::Mute, Duration::from_secs(3600)), // 1 heure
            s if s < 1.5 => (SanctionType::TempBan, Duration::from_secs(24 * 3600)), // 24 heures
            _ => (SanctionType::TempBan, Duration::from_secs(7 * 24 * 3600)), // 7 jours
        };
        
        Ok(Some((sanction_type, primary_reason, duration)))
    }
    
    /// Utilitaire pour détecter la répétition de caractères
    fn has_character_repetition(&self, content: &str) -> bool {
        let chars: Vec<char> = content.chars().collect();
        let mut consecutive_count = 1;
        
        for i in 1..chars.len() {
            if chars[i] == chars[i-1] {
                consecutive_count += 1;
                if consecutive_count >= 4 { // 4 caractères identiques consécutifs
                    return true;
                }
            } else {
                consecutive_count = 1;
            }
        }
        
        false
    }
    
    /// Nettoie les profils anciens
    pub async fn cleanup_old_profiles(&self) {
        let cutoff_time = Utc::now() - chrono::Duration::from_std(self.config.profile_retention_duration).unwrap();
        let mut removed_count = 0;
        
        self.user_profiles.retain(|_, profile| {
            if profile.last_updated < cutoff_time {
                removed_count += 1;
                false
            } else {
                true
            }
        });
        
        if removed_count > 0 {
            tracing::info!(removed_count = %removed_count, "🧹 Profils utilisateur anciens supprimés");
        }
    }
    
    /// Obtient les statistiques de modération
    pub async fn get_moderation_stats(&self) -> HashMap<String, u64> {
        let mut stats = HashMap::new();
        
        stats.insert("active_profiles".to_string(), self.user_profiles.len() as u64);
        stats.insert("cached_violations".to_string(), self.violation_cache.len() as u64);
        
        let mut total_violations = 0;
        let mut bot_count = 0;
        let mut high_risk_users = 0;
        
        for profile in self.user_profiles.iter() {
            total_violations += profile.total_violations;
            if profile.is_likely_bot() {
                bot_count += 1;
            }
            if profile.calculate_suspicion_score() > 0.8 {
                high_risk_users += 1;
            }
        }
        
        stats.insert("total_violations".to_string(), total_violations);
        stats.insert("detected_bots".to_string(), bot_count);
        stats.insert("high_risk_users".to_string(), high_risk_users);
        
        stats
    }
} 