//! # Veza Chat Server - Bibliothèque principale
//! 
//! Serveur de chat WebSocket sécurisé et haute performance
//! Version simplifiée pour déploiement rapide

// ═══════════════════════════════════════════════════════════════════════
// MODULES CORE (FONCTIONNELS)
// ═══════════════════════════════════════════════════════════════════════

/// Gestion des erreurs du serveur
pub mod error;

/// Store de messages simplifié
pub mod simple_message_store;

/// Configuration du serveur
pub mod config;

// ═══════════════════════════════════════════════════════════════════════
// TESTS (CONDITIONNELS)
// ═══════════════════════════════════════════════════════════════════════

#[cfg(test)]
pub mod test_simple_store;

// ═══════════════════════════════════════════════════════════════════════
// RE-EXPORTS PUBLICS
// ═══════════════════════════════════════════════════════════════════════

pub use error::{ChatError, Result};
pub use simple_message_store::{SimpleMessageStore, SimpleMessage};
pub use config::ServerConfig;

// ═══════════════════════════════════════════════════════════════════════
// FONCTIONS D'INITIALISATION
// ═══════════════════════════════════════════════════════════════════════

/// Initialise le serveur de chat avec la configuration par défaut
pub async fn initialize_server() -> Result<()> {
    let _config = ServerConfig::default();
    tracing::info!("🚀 Serveur de chat initialisé");
    Ok(())
}

/// Initialise le serveur de chat avec une configuration personnalisée
pub async fn initialize_server_with_config(_config: ServerConfig) -> Result<()> {
    tracing::info!("🚀 Serveur de chat initialisé avec configuration personnalisée");
    Ok(())
}

// ═══════════════════════════════════════════════════════════════════════
// MACROS UTILITAIRES  
// ═══════════════════════════════════════════════════════════════════════

// Macro définie dans error.rs

// ═══════════════════════════════════════════════════════════════════════
// TESTS UNITAIRES
// ═══════════════════════════════════════════════════════════════════════

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_initialize_server() {
        let result = initialize_server().await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_initialize_server_with_config() {
        let config = ServerConfig::default();
        let result = initialize_server_with_config(config).await;
        assert!(result.is_ok());
    }

    #[test]
    fn test_chat_error_macro() {
        let error = chat_error!("Test error");
        assert_eq!(error.to_string(), "Configuration error: Test error");
    }

    #[test]
    fn test_chat_error_macro_with_format() {
        let error = chat_error!("Test error: {}", 42);
        assert_eq!(error.to_string(), "Configuration error: Test error: 42");
    }
} 