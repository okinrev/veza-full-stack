//! Intégration de la Modération IA dans le Core
//! 
//! Ce module connecte l'AdvancedModerationEngine avec :
//! - Le système de messages en temps réel
//! - Les actions automatiques (mute, ban, delete)
//! - Les notifications de modération
//! - Les métriques de sécurité

use std::sync::Arc;
use std::time::Duration;
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use tokio::sync::mpsc;

use crate::advanced_moderation::{
    AdvancedModerationEngine, 
    AdvancedModerationConfig,
    ViolationType,
    UserBehaviorProfile
};
use crate::moderation::{SanctionType, SanctionReason};
use crate::monitoring::ChatMetrics;
use crate::permissions::{Permission, UserPermissions};
use crate::core::{ConnectionManager, RichMessage, RichMessageManager};
use crate::error::{ChatError, Result};

/// Service d'intégration de modération IA
#[derive(Debug)]
pub struct ModerationIntegrationService {
    /// Engine de modération IA
    moderation_engine: Arc<AdvancedModerationEngine>,
    
    /// Gestionnaire de connexions pour actions en temps réel
    connection_manager: Arc<ConnectionManager>,
    
    /// Gestionnaire de messages riches
    message_manager: Arc<RichMessageManager>,
    
    /// Channel pour les actions de modération
    action_sender: mpsc::UnboundedSender<ModerationAction>,
    
    /// Historique des sanctions
    sanction_history: Arc<DashMap<i64, Vec<SanctionRecord>>>,
    
    /// Whitelist d'utilisateurs de confiance
    trusted_users: Arc<DashMap<i64, TrustLevel>>,
    
    /// Métriques de modération
    metrics: Arc<ModerationMetrics>,
}

/// Action de modération à exécuter
#[derive(Debug, Clone)]
pub enum ModerationAction {
    /// Supprimer un message
    DeleteMessage {
        message_id: String,
        channel_id: String,
        reason: String,
    },
    
    /// Muter un utilisateur
    MuteUser {
        user_id: i64,
        duration: Duration,
        reason: String,
    },
    
    /// Bannir un utilisateur
    BanUser {
        user_id: i64,
        duration: Option<Duration>,
        reason: String,
    },
    
    /// Avertir un utilisateur
    WarnUser {
        user_id: i64,
        reason: String,
        violation_count: u32,
    },
    
    /// Alerter les modérateurs
    AlertModerators {
        user_id: i64,
        violations: Vec<ViolationType>,
        confidence: f32,
        urgent: bool,
    },
    
    /// Shadowban (restrictions invisibles)
    ShadowBan {
        user_id: i64,
        restrictions: ShadowBanRestrictions,
        duration: Duration,
    },
}

/// Sévérité d'une violation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ViolationSeverity {
    Low,     // Warning
    Medium,  // Temporary restrictions
    High,    // Temporary ban
    Critical, // Permanent ban
}

/// Restrictions de shadowban
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ShadowBanRestrictions {
    pub message_delay: Option<Duration>,
    pub limited_channels: bool,
    pub no_mentions: bool,
    pub no_reactions: bool,
    pub reduced_visibility: bool,
}

/// Enregistrement d'une sanction
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SanctionRecord {
    pub id: String,
    pub user_id: i64,
    pub reason: String,
    pub applied_at: DateTime<Utc>,
}

/// Niveau de confiance d'un utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TrustLevel {
    /// Utilisateur nouveau (surveillance accrue)
    New,
    /// Utilisateur normal
    Normal,
    /// Utilisateur de confiance (modération allégée)
    Trusted,
    /// Modérateur/VIP (bypass certaines vérifications)
    Privileged,
}

/// Métriques de modération
#[derive(Debug, Default)]
pub struct ModerationMetrics {
    pub messages_analyzed: Arc<std::sync::atomic::AtomicU64>,
    pub violations_detected: Arc<std::sync::atomic::AtomicU64>,
    pub auto_actions_taken: Arc<std::sync::atomic::AtomicU64>,
    pub false_positives: Arc<std::sync::atomic::AtomicU64>,
    pub manual_overrides: Arc<std::sync::atomic::AtomicU64>,
}

impl ModerationIntegrationService {
    pub fn new(connection_manager: Arc<ConnectionManager>) -> Result<Self> {
        let moderation_config = crate::advanced_moderation::AdvancedModerationConfig::default();
        let metrics = Arc::new(crate::monitoring::ChatMetrics::new());
        let moderation_engine = Arc::new(crate::advanced_moderation::AdvancedModerationEngine::new(moderation_config, metrics)?);
        
        Ok(Self {
            moderation_engine,
            connection_manager,
            message_manager: Arc::new(RichMessageManager::default()),
            action_sender: mpsc::unbounded_channel().0,
            sanction_history: Arc::new(DashMap::new()),
            trusted_users: Arc::new(DashMap::new()),
            metrics: Arc::new(ModerationMetrics::default()),
        })
    }
    
    pub async fn analyze_message(&self, message: &RichMessage) -> Result<ModerationDecision> {
        let violations = self.moderation_engine.analyze_message(
            message.author_id as i32,
            &message.author_username,
            &message.content,
            &message.channel_id,
            None,
        ).await?;
        
        let decision = if violations.is_empty() {
            ModerationDecision {
                allowed: true,
                action: None,
                violations: vec![],
                confidence: 0.0,
                reason: "Aucune violation détectée".to_string(),
            }
        } else {
            self.make_decision(message, &violations).await?
        };
        
        Ok(decision)
    }
    
    async fn make_decision(&self, message: &RichMessage, violations: &[ViolationType]) -> Result<ModerationDecision> {
        let confidence = self.calculate_confidence(violations);
        
        let action = if confidence > 0.8 {
            Some(ModerationAction::BanUser {
                user_id: message.author_id,
                duration: Some(Duration::from_secs(3600)),
                reason: "Violations critiques détectées".to_string(),
            })
        } else if confidence > 0.5 {
            Some(ModerationAction::DeleteMessage {
                message_id: message.id.clone(),
                channel_id: message.channel_id.clone(),
                reason: "Contenu inapproprié".to_string(),
            })
        } else {
            None
        };
        
        Ok(ModerationDecision {
            allowed: action.is_none(),
            action,
            violations: violations.to_vec(),
            confidence,
            reason: self.generate_reason(violations),
        })
    }
    
    fn calculate_confidence(&self, violations: &[ViolationType]) -> f32 {
        violations.iter()
            .map(|v| match v {
                ViolationType::Spam { confidence, .. } => *confidence,
                ViolationType::Toxicity { confidence, .. } => *confidence,
                ViolationType::Inappropriate { confidence, .. } => *confidence,
                ViolationType::Fraud { confidence, .. } => *confidence,
                ViolationType::Abuse { confidence, .. } => *confidence,
                ViolationType::Suspicious { confidence, .. } => *confidence,
            })
            .fold(0.0, |acc, x| acc.max(x))
    }
    
    fn generate_reason(&self, violations: &[ViolationType]) -> String {
        if violations.is_empty() {
            return "Aucune violation".to_string();
        }
        
        violations.iter()
            .map(|v| match v {
                ViolationType::Spam { .. } => "Spam",
                ViolationType::Toxicity { .. } => "Toxicité",
                ViolationType::Inappropriate { .. } => "Contenu inapproprié",
                ViolationType::Fraud { .. } => "Fraude",
                ViolationType::Abuse { .. } => "Abus",
                ViolationType::Suspicious { .. } => "Suspect",
            })
            .collect::<Vec<_>>()
            .join(", ")
    }
}

/// Décision de modération pour un message
#[derive(Debug, Clone)]
pub struct ModerationDecision {
    /// Le message est-il autorisé ?
    pub allowed: bool,
    /// Action à exécuter (si any)
    pub action: Option<ModerationAction>,
    /// Violations détectées
    pub violations: Vec<ViolationType>,
    /// Score de confiance
    pub confidence: f32,
    /// Raison lisible
    pub reason: String,
}

// Implémentation de Clone pour le service (pour le worker)
impl Clone for ModerationIntegrationService {
    fn clone(&self) -> Self {
        Self {
            moderation_engine: self.moderation_engine.clone(),
            connection_manager: self.connection_manager.clone(),
            message_manager: self.message_manager.clone(),
            action_sender: self.action_sender.clone(),
            sanction_history: self.sanction_history.clone(),
            trusted_users: self.trusted_users.clone(),
            metrics: self.metrics.clone(),
        }
    }
}

// Extensions pour ConnectionManager
impl ConnectionManager {
    pub async fn mute_user(&self, user_id: i64, duration: Duration) -> Result<()> {
        // Implémentation pour muter un utilisateur
        tracing::info!("Muting user {} for {:?}", user_id, duration);
        Ok(())
    }
    
    pub async fn ban_user(&self, user_id: i64, duration: Option<Duration>) -> Result<()> {
        // Implémentation pour bannir un utilisateur
        tracing::info!("Banning user {} for {:?}", user_id, duration);
        Ok(())
    }
} 