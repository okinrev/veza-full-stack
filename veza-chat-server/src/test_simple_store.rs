//! Tests pour le store simple

use crate::error::{ChatError, Result};
use crate::simple_message_store::SimpleMessageStore;

pub async fn test_simple_store(_database_url: &str) -> Result<()> {
    let store = SimpleMessageStore::new();
    
    // Test d'envoi de message
    let msg_id = store.send_simple_message(
        "Hello World!",
        "test_user",
        Some("general"),
        false
    ).await?;

    println!("✅ Message envoyé avec ID: {}", msg_id);

    // Test de récupération
    let messages = store.get_room_messages("general", 10).await?;
    println!("✅ Messages récupérés: {}", messages.len());

    // Test de message direct
    let dm_id = store.send_simple_message(
        "Private message",
        "user1",
        None,
        true
    ).await?;

    println!("✅ Message direct envoyé avec ID: {}", dm_id);

    // Test de récupération DM
    let dms = store.get_direct_messages("user1", "user2", 10).await?;
    println!("✅ Messages directs récupérés: {}", dms.len());

    // Test d'édition
    store.edit_message(msg_id, "Hello World Updated!").await?;
    println!("✅ Message édité");

    // Test de suppression
    store.delete_message(msg_id).await?;
    println!("✅ Message supprimé");

    // Tests d'autres fonctions
    store.pin_message(1).await?;
    let exists = store.message_exists(999).await?;
    println!("✅ Message 999 existe: {}", exists);

    store.add_reaction(1, 1, "👍").await?;
    store.remove_reaction(1, 1, "👍").await?;
    store.mark_as_read(1, "conversation1").await?;

    let unread = store.count_unread(1).await?;
    let unread_dms = store.count_unread_dms(1).await?;
    let unread_mentions = store.count_unread_mentions(1).await?;
    let reactions = store.count_reactions(1).await?;

    println!("✅ Compteurs - Non lus: {}, DMs: {}, Mentions: {}, Réactions: {}", 
             unread, unread_dms, unread_mentions, reactions);

    println!("🎉 Tous les tests passés!");
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_store_integration() {
        let database_url = std::env::var("DATABASE_URL")
            .unwrap_or_else(|_| "postgresql://veza_user:veza_password@10.5.191.134:5432/veza_db".to_string());
        
        if let Err(e) = test_simple_store(&database_url).await {
            panic!("Test failed: {}", e);
        }
    }
} 