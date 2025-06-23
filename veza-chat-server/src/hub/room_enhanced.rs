//! Module de compatibilité pour room_enhanced
//! 
//! Ce module fait le pont avec le nouveau module channels pour maintenir
//! la compatibilité avec l'ancien code qui référençait room_enhanced

use crate::error::Result;
use crate::hub::{ChatHub, channels};
use serde_json::Value;

/// Fonction de compatibilité pour envoyer un message dans un salon
pub async fn send_room_message(
    hub: &ChatHub,
    room_id: i64,
    user_id: i64,
    username: &str,
    content: &str,
    parent_id: Option<i64>,
) -> Result<String> {
    // Délégation vers le nouveau module channels
    let result = channels::send_room_message(hub, room_id, user_id, username, content, parent_id, None).await?;
    Ok(result.to_string())
}

/// Fonction de compatibilité pour récupérer l'historique d'un salon
pub async fn fetch_room_history(
    hub: &ChatHub,
    room_id: i64,
    user_id: i64,
    limit: i32,
    before_id: Option<i64>,
) -> Result<Vec<channels::RoomMessage>> {
    channels::fetch_room_history(hub, room_id, user_id, limit.into(), before_id).await
}

/// Fonction de compatibilité pour récupérer les messages épinglés
pub async fn fetch_pinned_messages(
    hub: &ChatHub,
    room_id: i64,
    user_id: i64,
) -> Result<Vec<channels::RoomMessage>> {
    channels::fetch_pinned_messages(hub, room_id, user_id).await
}

/// Fonction de compatibilité pour créer un salon
pub async fn create_room(
    hub: &ChatHub,
    room_name: &str,
    creator_id: i64,
    description: Option<&str>,
) -> Result<i64> {
    let result = channels::create_room(hub, creator_id, room_name, description, true, None).await?;
    Ok(result.id)
}

/// Fonction de compatibilité pour rejoindre un salon
pub async fn join_room(
    hub: &ChatHub,
    room_id: i64,
    user_id: i64,
) -> Result<()> {
    channels::join_room(hub, room_id, user_id).await
}

/// Fonction de compatibilité pour quitter un salon
pub async fn leave_room(
    hub: &ChatHub,
    room_id: i64,
    user_id: i64,
) -> Result<()> {
    channels::leave_room(hub, room_id, user_id).await
}

/// Fonction de compatibilité pour obtenir les statistiques d'un salon
pub async fn get_room_stats(
    hub: &ChatHub,
    room_id: i64,
) -> Result<channels::RoomStats> {
    channels::get_room_stats(hub, room_id).await
}

/// Fonction de compatibilité pour lister les membres d'un salon
pub async fn list_room_members(
    hub: &ChatHub,
    room_id: i64,
) -> Result<Vec<channels::RoomMember>> {
    channels::list_room_members(hub, room_id, 1).await
} 