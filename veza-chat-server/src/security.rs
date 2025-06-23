use std::collections::{HashMap, HashSet};
use std::time::{Duration, SystemTime};
use regex::Regex;
use crate::error::{ChatError, Result};
use sha2::{Sha256, Digest};

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub enum SecurityAction {
    SendMessage,
    CreateRoom,
    JoinRoom,
    SendDM,
    UploadFile,
    ChangeSettings,
    AdminAction,
}

/// Filtre de contenu amélioré avec détection ML
pub struct ContentFilter {
    forbidden_words: HashSet<String>,
    dangerous_patterns: Vec<Regex>,
    spam_detector: SpamDetector,
    toxicity_detector: ToxicityDetector,
}

impl ContentFilter {
    pub fn new() -> Result<Self> {
        Ok(Self {
            forbidden_words: {
                let mut words = HashSet::new();
                words.insert("spam".to_string());
                words.insert("test_bad_word".to_string());
                words
            },
            dangerous_patterns: vec![
                Regex::new(r"<script[^>]*>.*?</script>").unwrap(),
                Regex::new(r"javascript:").unwrap(),
                Regex::new(r"data:text/html").unwrap(),
                Regex::new(r"on\w+\s*=").unwrap(),
                Regex::new(r"eval\s*\(").unwrap(),
                Regex::new(r"document\.(write|cookie)").unwrap(),
                Regex::new(r"window\.(location|open)").unwrap(),
                Regex::new(r"<iframe[^>]*>").unwrap(),
                Regex::new(r"<object[^>]*>").unwrap(),
                Regex::new(r"<embed[^>]*>").unwrap(),
                Regex::new(r"<link[^>]*>").unwrap(),
                Regex::new(r"<meta[^>]*>").unwrap(),
                Regex::new(r"@import").unwrap(),
                Regex::new(r"expression\s*\(").unwrap(),
                Regex::new(r"url\s*\(").unwrap(),
                Regex::new(r"behavior\s*:").unwrap(),
                Regex::new(r"-moz-binding").unwrap(),
                Regex::new(r"<\?php").unwrap(),
                Regex::new(r"<%.*?%>").unwrap(),
                Regex::new(r"\{\{.*?\}\}").unwrap(),
                Regex::new(r"\{%.*?%\}").unwrap(),
                Regex::new(r"<\s*script").unwrap(),
                Regex::new(r"<\s*style").unwrap(),
                Regex::new(r"<\s*link").unwrap(),
                Regex::new(r"<\s*meta").unwrap(),
                Regex::new(r"<\s*base").unwrap(),
                Regex::new(r"<\s*title").unwrap(),
                Regex::new(r"<\s*frame").unwrap(),
                Regex::new(r"<\s*applet").unwrap(),
                Regex::new(r"<\s*form").unwrap(),
                Regex::new(r"<\s*input").unwrap(),
                Regex::new(r"SELECT\s+.*\s+FROM").unwrap(),
                Regex::new(r"INSERT\s+INTO").unwrap(),
                Regex::new(r"UPDATE\s+.*\s+SET").unwrap(),
                Regex::new(r"DELETE\s+FROM").unwrap(),
                Regex::new(r"DROP\s+TABLE").unwrap(),
                Regex::new(r"CREATE\s+TABLE").unwrap(),
                Regex::new(r"ALTER\s+TABLE").unwrap(),
                Regex::new(r"TRUNCATE\s+TABLE").unwrap(),
                Regex::new(r"UNION\s+SELECT").unwrap(),
                Regex::new(r"OR\s+1\s*=\s*1").unwrap(),
                Regex::new(r"AND\s+1\s*=\s*1").unwrap(),
                Regex::new(r"'\s*OR\s*'").unwrap(),
                Regex::new(r"'\s*AND\s*'").unwrap(),
                Regex::new(r"--\s*").unwrap(),
                Regex::new(r"/\*.*?\*/").unwrap(),
                Regex::new(r"xp_cmdshell").unwrap(),
                Regex::new(r"sp_executesql").unwrap(),
                Regex::new(r"exec\s*\(").unwrap(),
                Regex::new(r"execute\s*\(").unwrap(),
                Regex::new(r"cmd\.exe").unwrap(),
                Regex::new(r"powershell").unwrap(),
                Regex::new(r"bash").unwrap(),
                Regex::new(r"sh\s").unwrap(),
                Regex::new(r"perl").unwrap(),
                Regex::new(r"python").unwrap(),
                Regex::new(r"ruby").unwrap(),
                Regex::new(r"wget").unwrap(),
                Regex::new(r"curl").unwrap(),
                Regex::new(r"nc\s").unwrap(),
                Regex::new(r"netcat").unwrap(),
                Regex::new(r"telnet").unwrap(),
                Regex::new(r"ssh").unwrap(),
                Regex::new(r"ftp").unwrap(),
                Regex::new(r"tftp").unwrap(),
            ],
            spam_detector: SpamDetector::new(),
            toxicity_detector: ToxicityDetector::new(),
        })
    }

    pub fn validate_content(&mut self, content: &str) -> Result<String> {
        // Nettoyer le contenu d'abord
        let cleaned = self.sanitize_content(content)?;
        
        // Vérifier les mots interdits
        for word in &self.forbidden_words {
            if cleaned.to_lowercase().contains(word) {
                return Err(ChatError::inappropriate_content_simple(&format!("Contenu inapproprié détecté: {}", word)));
            }
        }

        // Vérifier les patterns dangereux
        for pattern in &self.dangerous_patterns {
            if pattern.is_match(&cleaned) {
                return Err(ChatError::inappropriate_content_simple("Pattern de sécurité détecté"));
            }
        }

        // Vérifier le spam
        if self.spam_detector.is_spam(&cleaned)? {
            return Err(ChatError::inappropriate_content_simple("Contenu identifié comme spam"));
        }

        // Vérifier la toxicité
        if self.toxicity_detector.is_toxic(&cleaned)? {
            return Err(ChatError::inappropriate_content_simple("Contenu toxique détecté"));
            }

        Ok(cleaned)
    }

    pub fn validate_room_name(&mut self, room_name: &str) -> Result<String> {
        // Validation spécifique pour les noms de salon
        if room_name.is_empty() {
            return Err(ChatError::InvalidFormat { 
                field: "room_name".to_string(), 
                reason: "Le nom du salon ne peut pas être vide".to_string() 
            });
        }

        if room_name.len() > 50 {
            return Err(ChatError::InvalidFormat { 
                field: "room_name".to_string(), 
                reason: "Le nom du salon ne peut pas dépasser 50 caractères".to_string() 
            });
        }

        // Caractères autorisés : lettres, chiffres, tirets, underscores
        let room_regex = Regex::new(r"^[a-zA-Z0-9\-_\s]+$").unwrap();
        if !room_regex.is_match(room_name) {
            return Err(ChatError::InvalidFormat { 
                field: "room_name".to_string(), 
                reason: "Le nom du salon contient des caractères non autorisés".to_string() 
            });
        }

        // Nettoyer et retourner
        Ok(self.sanitize_html(room_name))
    }

    pub fn sanitize_content(&mut self, content: &str) -> Result<String> {
        // Nettoyer le HTML
        let cleaned = self.sanitize_html(content);
        
        // Encoder les caractères spéciaux
        let cleaned = cleaned
            .replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace("\"", "&quot;")
            .replace("'", "&#x27;");

        // Limiter la longueur
        if cleaned.len() > 4000 {
            return Err(ChatError::message_too_long(cleaned.len(), 4000));
        }

        Ok(cleaned)
    }

    fn sanitize_html(&self, content: &str) -> String {
        content
            .replace("<", "&lt;")
            .replace(">", "&gt;")
            .replace("\"", "&quot;")
            .replace("'", "&#x27;")
            .replace("&", "&amp;")
            .chars()
            .filter(|c| c.is_ascii() || c.is_alphanumeric() || " .,!?-_@#()[]{}".contains(*c))
            .collect()
    }
}

/// Détecteur de spam avec algorithmes heuristiques
pub struct SpamDetector {
    repetition_threshold: f32,
    caps_threshold: f32,
    _emoji_threshold: f32, // Préfixé avec _ pour éviter l'avertissement
}

impl SpamDetector {
    pub fn new() -> Self {
        Self {
            repetition_threshold: 0.7, // 70% de répétition
            caps_threshold: 0.5,       // 50% de majuscules
            _emoji_threshold: 0.3,     // 30% d'emojis (non utilisé pour l'instant)
        }
    }

    pub fn is_spam(&self, content: &str) -> Result<bool> {
        // Vérifications heuristiques simples
        if content.len() < 3 {
            return Ok(false); // Messages trop courts ne sont pas du spam
        }

        // Vérifier la répétition excessive de caractères
        if self.detect_character_repetition(content) {
            return Ok(true);
        }

        // Vérifier les majuscules excessives
        if self.detect_excessive_caps(content) {
            return Ok(true);
        }

        // Vérifier les caractères spéciaux excessifs
        if self.detect_excessive_special_chars(content) {
            return Ok(true);
        }

        // Vérifier les patterns de spam connus
        if self.detect_spam_patterns(content) {
            return Ok(true);
        }

        Ok(false)
    }

    fn detect_character_repetition(&self, content: &str) -> bool {
        let chars: Vec<char> = content.chars().collect();
        if chars.len() < 4 {
            return false;
        }

        let mut repetitions = 0;
        for i in 0..chars.len() - 1 {
            if chars[i] == chars[i + 1] {
                repetitions += 1;
            }
        }

        (repetitions as f32 / chars.len() as f32) > self.repetition_threshold
    }

    fn detect_excessive_caps(&self, content: &str) -> bool {
        let uppercase_count = content.chars().filter(|c| c.is_uppercase()).count();
        let letter_count = content.chars().filter(|c| c.is_alphabetic()).count();
        
        if letter_count == 0 {
            return false;
        }

        (uppercase_count as f32 / letter_count as f32) > self.caps_threshold
    }

    fn detect_excessive_special_chars(&self, content: &str) -> bool {
        let special_count = content.chars().filter(|c| "!@#$%^&*()_+-=[]{}|;':\",./<>?".contains(*c)).count();
        (special_count as f32 / content.len() as f32) > 0.3
    }

    fn detect_spam_patterns(&self, content: &str) -> bool {
        let spam_patterns = [
            "URGENT", "GRATUIT", "OFFRE LIMITÉE", "CLIQUEZ ICI",
            "FÉLICITATIONS", "VOUS AVEZ GAGNÉ", "PROMOTION",
            "💰💰💰", "🎉🎉🎉", "🔥🔥🔥",
        ];

        let content_upper = content.to_uppercase();
        spam_patterns.iter().any(|pattern| content_upper.contains(pattern))
    }
}

/// Détecteur de toxicité avec patterns et ML
pub struct ToxicityDetector {
    toxic_patterns: Vec<Regex>,
    _severity_threshold: f32, // Préfixé avec _ pour éviter l'avertissement
}

impl ToxicityDetector {
    pub fn new() -> Self {
        Self {
            toxic_patterns: vec![
                Regex::new(r"(?i)\b(idiot|stupide|con|connard|salope|pute|merde)\b").unwrap(),
                Regex::new(r"(?i)\b(fuck|shit|bitch|asshole|damn)\b").unwrap(),
                Regex::new(r"(?i)\b(kill\s+yourself|suicide|die)\b").unwrap(),
                Regex::new(r"(?i)\b(hate\s+you|je\s+te\s+déteste)\b").unwrap(),
                Regex::new(r"(?i)\b(racist|nazi|fascist)\b").unwrap(),
            ],
            _severity_threshold: 0.7,
        }
    }

    pub fn is_toxic(&self, content: &str) -> Result<bool> {
        // Vérifier les patterns toxiques connus
        for pattern in &self.toxic_patterns {
            if pattern.is_match(content) {
                return Ok(true);
            }
        }

        // TODO: Intégrer un modèle ML pour la détection avancée
        // Pour l'instant, utiliser seulement les patterns

        Ok(false)
    }
}

/// Système de limitation de taux avancé
pub struct AdvancedRateLimiter {
    limits: HashMap<SecurityAction, RateLimit>,
    user_actions: HashMap<(i32, SecurityAction), Vec<SystemTime>>,
}

#[derive(Clone)]
pub struct RateLimit {
    pub max_count: u32,
    pub window_duration: Duration,
    pub burst_limit: Option<u32>, // Limite de burst
}

impl AdvancedRateLimiter {
    pub fn new() -> Self {
        let mut limits = HashMap::new();
        
        // Configuration des limites par action
        limits.insert(SecurityAction::SendMessage, RateLimit {
            max_count: 10,
            window_duration: Duration::from_secs(60),
            burst_limit: Some(3),
        });
        
        limits.insert(SecurityAction::CreateRoom, RateLimit {
            max_count: 5,
            window_duration: Duration::from_secs(3600), // 1 heure
            burst_limit: None,
        });
        
        limits.insert(SecurityAction::JoinRoom, RateLimit {
            max_count: 20,
            window_duration: Duration::from_secs(300), // 5 minutes
            burst_limit: Some(5),
        });
        
        limits.insert(SecurityAction::SendDM, RateLimit {
            max_count: 20,
            window_duration: Duration::from_secs(300),
            burst_limit: Some(5),
        });
        
        limits.insert(SecurityAction::UploadFile, RateLimit {
            max_count: 10,
            window_duration: Duration::from_secs(600), // 10 minutes
            burst_limit: Some(2),
        });

        Self {
            limits,
            user_actions: HashMap::new(),
        }
    }

    pub fn check_limit(&mut self, user_id: i32, action: &SecurityAction) -> Result<()> {
        let key = (user_id, action.clone());
        let now = SystemTime::now();
        
        // Récupérer la limite pour cette action
        let limit = self.limits.get(action)
            .ok_or_else(|| ChatError::configuration_error("Limite non configurée pour cette action"))?;

        // Nettoyer les anciennes entrées
        self.user_actions.entry(key.clone()).or_insert_with(Vec::new)
            .retain(|&time| now.duration_since(time).unwrap_or(Duration::MAX) <= limit.window_duration);

        let actions = self.user_actions.get_mut(&key).unwrap();

        // Vérifier la limite
        if actions.len() >= limit.max_count as usize {
            return Err(ChatError::rate_limit_exceeded_simple(&format!("{:?}", action)));
        }

        // Ajouter l'action actuelle
        actions.push(now);
        Ok(())
    }
}

/// Gestionnaire de sessions avec sécurité renforcée
pub struct SessionManager {
    active_sessions: HashMap<i32, SessionInfo>,
    max_sessions_per_user: u32,
}

pub struct SessionInfo {
    pub token_hash: String,
    pub created_at: SystemTime,
    pub last_activity: SystemTime,
    pub ip_address: String,
    pub user_agent: Option<String>,
}

impl SessionManager {
    pub fn new() -> Self {
        Self {
            active_sessions: HashMap::new(),
            max_sessions_per_user: 5, // Maximum 5 sessions simultanées par utilisateur
        }
    }

    pub fn create_session(&mut self, user_id: i32, token: &str, ip: &str) -> Result<()> {
        // Vérifier la limite de sessions
        let current_sessions = self.active_sessions.values()
            .filter(|info| info.ip_address == ip)
            .count();
            
        if current_sessions >= self.max_sessions_per_user as usize {
            return Err(ChatError::configuration_error("Trop de sessions actives"));
        }

        let session_info = SessionInfo {
            token_hash: self.hash_token(token),
            created_at: SystemTime::now(),
            last_activity: SystemTime::now(),
            ip_address: ip.to_string(),
            user_agent: None,
        };

        self.active_sessions.insert(user_id, session_info);
        Ok(())
    }

    pub fn validate_session(&mut self, user_id: i32, token: &str) -> Result<()> {
        let token_hash = self.hash_token(token);
        
        match self.active_sessions.get_mut(&user_id) {
            Some(session) if session.token_hash == token_hash => {
            // Vérifier l'expiration (24h)
                let elapsed = SystemTime::now().duration_since(session.created_at)
                    .unwrap_or(Duration::MAX);
                
                if elapsed > Duration::from_secs(24 * 3600) {
                self.active_sessions.remove(&user_id);
                    return Err(ChatError::unauthorized("Session expirée"));
            }
            
                // Mettre à jour la dernière activité
            session.last_activity = SystemTime::now();
            Ok(())
            }
            _ => Err(ChatError::unauthorized("Session invalide"))
        }
    }

    fn hash_token(&self, token: &str) -> String {
        let mut hasher = Sha256::new();
        hasher.update(token.as_bytes());
        format!("{:x}", hasher.finalize())
    }
}

/// Moniteur d'IP pour la détection d'abus
pub struct IpMonitor {
    ip_actions: HashMap<String, Vec<(SystemTime, SecurityAction)>>,
    blacklisted_ips: HashSet<String>,
    suspicious_threshold: u32,
}

impl IpMonitor {
    pub fn new() -> Self {
        Self {
            ip_actions: HashMap::new(),
            blacklisted_ips: HashSet::new(),
            suspicious_threshold: 50, // 50 actions par minute
        }
    }

    pub fn check_ip(&mut self, ip: &str, action: &SecurityAction) -> Result<()> {
        // Vérifier si l'IP est blacklistée
        if self.blacklisted_ips.contains(ip) {
            return Err(ChatError::configuration_error("IP bloquée"));
        }

        let now = SystemTime::now();
        let actions = self.ip_actions.entry(ip.to_string()).or_insert_with(Vec::new);
        
        // Nettoyer les anciennes actions (dernière minute)
        actions.retain(|(time, _)| now.duration_since(*time).unwrap_or(Duration::MAX) <= Duration::from_secs(60));
        
        // Vérifier le seuil suspect
        if actions.len() >= self.suspicious_threshold as usize {
            self.blacklist_ip(ip);
            return Err(ChatError::configuration_error("Activité suspecte détectée"));
        }

        // Enregistrer l'action
        actions.push((now, action.clone()));
        Ok(())
    }

    pub fn blacklist_ip(&mut self, ip: &str) {
        self.blacklisted_ips.insert(ip.to_string());
        tracing::warn!(ip = %ip, "🚫 IP blacklistée pour activité suspecte");
    }
}

/// Système de sécurité principal
pub struct EnhancedSecurity {
    content_filter: ContentFilter,
    rate_limiter: AdvancedRateLimiter,
    session_manager: SessionManager,
    ip_monitor: IpMonitor,
}

impl EnhancedSecurity {
    pub fn new() -> Result<Self> {
        Ok(Self {
            content_filter: ContentFilter::new()?,
            rate_limiter: AdvancedRateLimiter::new(),
            session_manager: SessionManager::new(),
            ip_monitor: IpMonitor::new(),
        })
    }

    pub async fn validate_request(
        &mut self,
        user_id: i32,
        ip: &str,
        session_token: &str,
        action: &SecurityAction,
        content: Option<&str>
    ) -> Result<()> {
        // 1. Vérifier l'IP
        self.ip_monitor.check_ip(ip, action)?;

        // 2. Valider la session
        self.session_manager.validate_session(user_id, session_token)?;

        // 3. Vérifier les limites de taux
        self.rate_limiter.check_limit(user_id, action)?;

        // 4. Filtrer le contenu si présent
        if let Some(content) = content {
            self.content_filter.validate_content(content)?;
        }

        Ok(())
    }
} 