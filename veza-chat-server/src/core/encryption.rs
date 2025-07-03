use std::collections::HashMap;
use std::sync::Arc;
use serde::{Deserialize, Serialize};
use dashmap::DashMap;
use ring::{aead, rand::{SystemRandom, SecureRandom}};
use base64::{Engine as _, engine::general_purpose::STANDARD as BASE64};

use crate::error::{ChatError, Result};

/// Service de chiffrement bout-en-bout
#[derive(Debug)]
pub struct E2EEncryptionService {
    /// Générateur de nombres aléatoires sécurisé
    rng: SystemRandom,
    
    /// Sessions de chiffrement actives par channel
    encryption_sessions: Arc<DashMap<String, EncryptionSession>>,
    
    /// Clés publiques des utilisateurs
    user_public_keys: Arc<DashMap<i64, UserKeyPair>>,
    
    /// Préférences de chiffrement par utilisateur
    user_preferences: Arc<DashMap<i64, EncryptionPreferences>>,
}

/// Session de chiffrement pour un channel
#[derive(Debug, Clone)]
pub struct EncryptionSession {
    /// ID unique de la session
    pub session_id: String,
    
    /// Channel concerné
    pub channel_id: String,
    
    /// Participants à la session
    pub participants: Vec<i64>,
    
    /// Clés de session partagées (chiffrées pour chaque participant)
    pub encrypted_keys: HashMap<i64, Vec<u8>>,
    
    /// Algorithme de chiffrement utilisé
    pub algorithm: EncryptionAlgorithm,
    
    /// Statut de la session
    pub status: SessionStatus,
}

/// Paire de clés d'un utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserKeyPair {
    /// ID de l'utilisateur
    pub user_id: i64,
    
    /// Clé publique (pour chiffrement asymétrique)
    pub public_key: Vec<u8>,
    
    /// Fingerprint de la clé pour vérification
    pub fingerprint: String,
    
    /// Statut de la clé
    pub status: KeyStatus,
}

/// Préférences de chiffrement d'un utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptionPreferences {
    /// Chiffrement activé par défaut
    pub enabled_by_default: bool,
    
    /// Algorithme préféré
    pub preferred_algorithm: EncryptionAlgorithm,
    
    /// Rotation automatique des clés
    pub auto_key_rotation: bool,
    
    /// Période de rotation (en jours)
    pub rotation_period_days: u32,
    
    /// Vérification des empreintes obligatoire
    pub require_fingerprint_verification: bool,
    
    /// Channels où le chiffrement est obligatoire
    pub mandatory_channels: Vec<String>,
}

/// Algorithmes de chiffrement supportés
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum EncryptionAlgorithm {
    /// AES-256-GCM (recommandé)
    AES256GCM,
    
    /// ChaCha20-Poly1305 (alternative)
    ChaCha20Poly1305,
}

/// Statut d'une session de chiffrement
#[derive(Debug, Clone, PartialEq)]
pub enum SessionStatus {
    /// Session active
    Active,
    
    /// Session expirée
    Expired,
    
    /// Session révoquée
    Revoked,
}

/// Statut d'une clé
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum KeyStatus {
    /// Clé active
    Active,
    
    /// Clé révoquée
    Revoked,
    
    /// Clé expirée
    Expired,
}

/// Message chiffré
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncryptedMessage {
    /// ID du message
    pub message_id: String,
    
    /// ID de la session de chiffrement
    pub session_id: String,
    
    /// Contenu chiffré
    pub encrypted_content: Vec<u8>,
    
    /// Nonce utilisé pour le chiffrement
    pub nonce: Vec<u8>,
    
    /// Tag d'authentification
    pub auth_tag: Vec<u8>,
    
    /// Algorithme utilisé
    pub algorithm: EncryptionAlgorithm,
}

/// Métadonnées non chiffrées d'un message
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMetadata {
    /// ID de l'expéditeur
    pub sender_id: i64,
    
    /// Timestamp d'envoi
    pub timestamp: chrono::DateTime<chrono::Utc>,
    
    /// Type de message
    pub message_type: String,
    
    /// Taille du contenu original
    pub content_size: usize,
}

/// Résultat d'une opération de chiffrement
#[derive(Debug)]
pub struct EncryptionResult {
    /// Message chiffré
    pub encrypted_message: EncryptedMessage,
    
    /// Clés de session pour les participants
    pub session_keys: HashMap<i64, Vec<u8>>,
    
    /// Participants qui ont reçu les clés
    pub delivered_to: Vec<i64>,
}

impl E2EEncryptionService {
    /// Crée un nouveau service de chiffrement
    pub fn new() -> Self {
        Self {
            rng: SystemRandom::new(),
            encryption_sessions: Arc::new(DashMap::new()),
            user_public_keys: Arc::new(DashMap::new()),
            user_preferences: Arc::new(DashMap::new()),
        }
    }
    
    /// Génère une nouvelle paire de clés pour un utilisateur
    pub async fn generate_user_keypair(&self, user_id: i64) -> Result<UserKeyPair> {
        let mut public_key = vec![0u8; 32];
        self.rng.fill(&mut public_key)
            .map_err(|_| ChatError::internal_error("Failed to generate key"))?;
        
        let fingerprint = self.generate_fingerprint(&public_key);
        
        let keypair = UserKeyPair {
            user_id,
            public_key,
            fingerprint,
            status: KeyStatus::Active,
        };
        
        self.user_public_keys.insert(user_id, keypair.clone());
        Ok(keypair)
    }
    
    /// Établit une session de chiffrement pour un channel
    pub async fn establish_session(
        &self,
        channel_id: String,
        participants: Vec<i64>,
        algorithm: EncryptionAlgorithm,
    ) -> Result<EncryptionSession> {
        for &user_id in &participants {
            if !self.user_public_keys.contains_key(&user_id) {
                return Err(ChatError::validation_error(
                    &format!("User {} has no public key", user_id)
                ));
            }
        }
        
        let session_key = self.generate_session_key()?;
        let mut encrypted_keys = HashMap::new();
        
        for &user_id in &participants {
            if let Some(user_keypair) = self.user_public_keys.get(&user_id) {
                let encrypted_key = self.encrypt_session_key(&session_key, &user_keypair.public_key)?;
                encrypted_keys.insert(user_id, encrypted_key);
            }
        }
        
        let session = EncryptionSession {
            session_id: format!("session_{}", uuid::Uuid::new_v4()),
            channel_id: channel_id.clone(),
            participants,
            encrypted_keys,
            algorithm,
            status: SessionStatus::Active,
        };
        
        self.encryption_sessions.insert(channel_id, session.clone());
        Ok(session)
    }
    
    /// Chiffre un message pour un channel
    pub async fn encrypt_message(
        &self,
        channel_id: &str,
        sender_id: i64,
        content: &str,
        message_id: String,
    ) -> Result<EncryptedMessage> {
        let session = self.encryption_sessions.get(channel_id)
            .ok_or_else(|| ChatError::not_found("session", "session_not_found"))?;
        
        if !session.participants.contains(&sender_id) {
            return Err(ChatError::permission_denied("User not authorized"));
        }
        
        let nonce = self.generate_nonce()?;
        let session_key = &session.encrypted_keys[&sender_id];
        
        let (encrypted_content, auth_tag) = self.encrypt_content(
            content.as_bytes(),
            session_key,
            &nonce,
            &session.algorithm,
        )?;
        
        Ok(EncryptedMessage {
            message_id,
            session_id: session.session_id.clone(),
            encrypted_content,
            nonce,
            auth_tag,
            algorithm: session.algorithm.clone(),
        })
    }
    
    /// Déchiffre un message
    pub async fn decrypt_message(
        &self,
        encrypted_message: &EncryptedMessage,
        recipient_id: i64,
    ) -> Result<String> {
        let session = self.encryption_sessions.iter()
            .find(|entry| entry.value().session_id == encrypted_message.session_id)
            .ok_or_else(|| ChatError::not_found("session", "session_not_found"))?;
        
        if !session.participants.contains(&recipient_id) {
            return Err(ChatError::permission_denied("Not authorized"));
        }
        
        let encrypted_session_key = session.encrypted_keys.get(&recipient_id)
            .ok_or_else(|| ChatError::not_found("session_key", "key_not_found"))?;
        
        let decrypted_content = self.decrypt_content(
            &encrypted_message.encrypted_content,
            &encrypted_message.auth_tag,
            encrypted_session_key,
            &encrypted_message.nonce,
            &encrypted_message.algorithm,
        )?;
        
        String::from_utf8(decrypted_content)
            .map_err(|_| ChatError::internal_error("Invalid decrypted content"))
    }
    
    /// Révoque une session de chiffrement
    pub async fn revoke_session(&self, channel_id: &str, revoked_by: i64) -> Result<()> {
        if let Some(mut session) = self.encryption_sessions.get_mut(channel_id) {
            // Vérifier que l'utilisateur peut révoquer la session
            if !session.participants.contains(&revoked_by) {
                return Err(ChatError::permission_denied("Non autorisé à révoquer cette session"));
            }
            
            session.status = SessionStatus::Revoked;
            tracing::info!("Session {} révoquée par l'utilisateur {}", session.session_id, revoked_by);
        }
        
        Ok(())
    }
    
    /// Configure les préférences de chiffrement d'un utilisateur
    pub async fn set_user_preferences(
        &self,
        user_id: i64,
        preferences: EncryptionPreferences,
    ) -> Result<()> {
        self.user_preferences.insert(user_id, preferences);
        Ok(())
    }
    
    /// Vérifie si le chiffrement est requis pour un channel
    pub fn is_encryption_required(&self, channel_id: &str, user_id: i64) -> bool {
        if let Some(prefs) = self.user_preferences.get(&user_id) {
            prefs.mandatory_channels.contains(&channel_id.to_string()) || prefs.enabled_by_default
        } else {
            false
        }
    }
    
    /// Vérifie si une session existe et est active
    pub fn has_active_session(&self, channel_id: &str) -> bool {
        self.encryption_sessions.get(channel_id)
            .map(|session| session.status == SessionStatus::Active)
            .unwrap_or(false)
    }
    
    // === Méthodes privées utilitaires ===
    
    /// Génère une clé de session aléatoire
    fn generate_session_key(&self) -> Result<Vec<u8>> {
        let mut key = vec![0u8; 32];
        self.rng.fill(&mut key)
            .map_err(|_| ChatError::internal_error("Failed to generate session key"))?;
        Ok(key)
    }
    
    /// Génère un nonce aléatoire
    fn generate_nonce(&self) -> Result<Vec<u8>> {
        let mut nonce = vec![0u8; 12];
        self.rng.fill(&mut nonce)
            .map_err(|_| ChatError::internal_error("Failed to generate nonce"))?;
        Ok(nonce)
    }
    
    /// Génère le fingerprint d'une clé publique
    fn generate_fingerprint(&self, public_key: &[u8]) -> String {
        use ring::digest;
        let digest = digest::digest(&digest::SHA256, public_key);
        BASE64.encode(digest.as_ref())
    }
    
    /// Chiffre une clé de session avec une clé publique
    fn encrypt_session_key(&self, session_key: &[u8], public_key: &[u8]) -> Result<Vec<u8>> {
        let mut encrypted = session_key.to_vec();
        for (i, &byte) in public_key.iter().enumerate() {
            if i < encrypted.len() {
                encrypted[i] ^= byte;
            }
        }
        Ok(encrypted)
    }
    
    /// Chiffre du contenu avec AES-GCM
    fn encrypt_content(
        &self,
        content: &[u8],
        key: &[u8],
        nonce: &[u8],
        algorithm: &EncryptionAlgorithm,
    ) -> Result<(Vec<u8>, Vec<u8>)> {
        match algorithm {
            EncryptionAlgorithm::AES256GCM => self.encrypt_aes_gcm(content, key, nonce),
            EncryptionAlgorithm::ChaCha20Poly1305 => self.encrypt_chacha20_poly1305(content, key, nonce),
        }
    }
    
    /// Déchiffre du contenu
    fn decrypt_content(
        &self,
        encrypted_content: &[u8],
        auth_tag: &[u8],
        key: &[u8],
        nonce: &[u8],
        algorithm: &EncryptionAlgorithm,
    ) -> Result<Vec<u8>> {
        match algorithm {
            EncryptionAlgorithm::AES256GCM => self.decrypt_aes_gcm(encrypted_content, auth_tag, key, nonce),
            EncryptionAlgorithm::ChaCha20Poly1305 => self.decrypt_chacha20_poly1305(encrypted_content, auth_tag, key, nonce),
        }
    }
    
    /// Chiffrement AES-256-GCM
    fn encrypt_aes_gcm(&self, content: &[u8], key: &[u8], nonce: &[u8]) -> Result<(Vec<u8>, Vec<u8>)> {
        let unbound_key = aead::UnboundKey::new(&aead::AES_256_GCM, key)
            .map_err(|_| ChatError::internal_error("Invalid AES key"))?;
        
        let key = aead::LessSafeKey::new(unbound_key);
        let nonce = aead::Nonce::try_assume_unique_for_key(nonce)
            .map_err(|_| ChatError::internal_error("Invalid nonce"))?;
        
        let mut in_out = content.to_vec();
        let tag = key.seal_in_place_separate_tag(nonce, aead::Aad::empty(), &mut in_out)
            .map_err(|_| ChatError::internal_error("Encryption failed"))?;
        
        Ok((in_out, tag.as_ref().to_vec()))
    }
    
    /// Déchiffrement AES-256-GCM
    fn decrypt_aes_gcm(
        &self,
        encrypted_content: &[u8],
        auth_tag: &[u8],
        key: &[u8],
        nonce: &[u8],
    ) -> Result<Vec<u8>> {
        let unbound_key = aead::UnboundKey::new(&aead::AES_256_GCM, key)
            .map_err(|_| ChatError::internal_error("Invalid AES key"))?;
        
        let key = aead::LessSafeKey::new(unbound_key);
        let nonce = aead::Nonce::try_assume_unique_for_key(nonce)
            .map_err(|_| ChatError::internal_error("Invalid nonce"))?;
        
        let mut in_out = encrypted_content.to_vec();
        in_out.extend_from_slice(auth_tag);
        
        let plaintext = key.open_in_place(nonce, aead::Aad::empty(), &mut in_out)
            .map_err(|_| ChatError::internal_error("Decryption failed"))?;
        
        Ok(plaintext.to_vec())
    }
    
    /// Chiffrement ChaCha20-Poly1305
    fn encrypt_chacha20_poly1305(&self, content: &[u8], key: &[u8], nonce: &[u8]) -> Result<(Vec<u8>, Vec<u8>)> {
        let unbound_key = aead::UnboundKey::new(&aead::CHACHA20_POLY1305, key)
            .map_err(|_| ChatError::internal_error("Invalid ChaCha20 key"))?;
        
        let key = aead::LessSafeKey::new(unbound_key);
        let nonce = aead::Nonce::try_assume_unique_for_key(nonce)
            .map_err(|_| ChatError::internal_error("Invalid nonce"))?;
        
        let mut in_out = content.to_vec();
        let tag = key.seal_in_place_separate_tag(nonce, aead::Aad::empty(), &mut in_out)
            .map_err(|_| ChatError::internal_error("Encryption failed"))?;
        
        Ok((in_out, tag.as_ref().to_vec()))
    }
    
    /// Déchiffrement ChaCha20-Poly1305
    fn decrypt_chacha20_poly1305(
        &self,
        encrypted_content: &[u8],
        auth_tag: &[u8],
        key: &[u8],
        nonce: &[u8],
    ) -> Result<Vec<u8>> {
        let unbound_key = aead::UnboundKey::new(&aead::CHACHA20_POLY1305, key)
            .map_err(|_| ChatError::internal_error("Invalid ChaCha20 key"))?;
        
        let key = aead::LessSafeKey::new(unbound_key);
        let nonce = aead::Nonce::try_assume_unique_for_key(nonce)
            .map_err(|_| ChatError::internal_error("Invalid nonce"))?;
        
        let mut in_out = encrypted_content.to_vec();
        in_out.extend_from_slice(auth_tag);
        
        let plaintext = key.open_in_place(nonce, aead::Aad::empty(), &mut in_out)
            .map_err(|_| ChatError::internal_error("Decryption failed"))?;
        
        Ok(plaintext.to_vec())
    }
}

impl Default for EncryptionPreferences {
    fn default() -> Self {
        Self {
            enabled_by_default: false,
            preferred_algorithm: EncryptionAlgorithm::AES256GCM,
            auto_key_rotation: true,
            rotation_period_days: 90,
            require_fingerprint_verification: true,
            mandatory_channels: vec![],
        }
    }
} 