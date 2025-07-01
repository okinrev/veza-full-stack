/// Module des features sociales SoundCloud-like
/// 
/// Features :
/// - Follow/Followers système
/// - Likes avec notifications
/// - Reposts avec messages
/// - Partage avec analytics
/// - Système de commentaires
/// - Feed social personnalisé

use std::sync::Arc;
use std::collections::{HashMap, HashSet};
use std::time::{Duration, SystemTime, Instant};

use serde::{Serialize, Deserialize};
use uuid::Uuid;
use tokio::sync::{RwLock, broadcast, mpsc};
use parking_lot::Mutex;
use tracing::{info, debug, warn};

use crate::error::AppError;

/// Gestionnaire principal du système social
#[derive(Debug)]
pub struct SocialManager {
    /// Relations follows/followers
    follow_graph: Arc<RwLock<FollowGraph>>,
    /// Likes par track
    track_likes: Arc<RwLock<HashMap<Uuid, LikeData>>>,
    /// Reposts par track
    track_reposts: Arc<RwLock<HashMap<Uuid, RepostData>>>,
    /// Commentaires par track
    track_comments: Arc<RwLock<HashMap<Uuid, CommentData>>>,
    /// Configuration
    config: SocialConfig,
    /// Événements sociaux
    event_sender: broadcast::Sender<SocialEvent>,
    /// Cache des feeds
    feed_cache: Arc<RwLock<HashMap<i64, UserFeed>>>,
}

/// Graphe des relations sociales
#[derive(Debug, Clone, Default)]
pub struct FollowGraph {
    /// user_id -> Set des utilisateurs suivis
    following: HashMap<i64, HashSet<i64>>,
    /// user_id -> Set des followers
    followers: HashMap<i64, HashSet<i64>>,
    /// Statistiques par utilisateur
    user_stats: HashMap<i64, UserSocialStats>,
}

/// Données de likes pour une track
#[derive(Debug, Clone, Default)]
pub struct LikeData {
    /// Total de likes
    pub total_count: u64,
    /// Utilisateurs qui ont liké (pour éviter doublons)
    pub liked_by: HashSet<i64>,
    /// Timeline des likes pour analytics
    pub like_timeline: Vec<LikeEntry>,
}

/// Entry de like individuel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LikeEntry {
    pub user_id: i64,
    pub timestamp: SystemTime,
    pub source: LikeSource,
}

/// Source du like
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum LikeSource {
    Player,
    TrackPage,
    Playlist,
    Feed,
    Search,
    Embed,
}

/// Données de reposts pour une track
#[derive(Debug, Clone, Default)]
pub struct RepostData {
    /// Total de reposts
    pub total_count: u64,
    /// Reposts individuels
    pub reposts: Vec<RepostEntry>,
}

/// Entry de repost individuel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepostEntry {
    pub id: Uuid,
    pub user_id: i64,
    pub track_id: Uuid,
    pub message: Option<String>,
    pub timestamp: SystemTime,
    pub visibility: RepostVisibility,
    pub likes_count: u32,
    pub comments_count: u32,
}

/// Visibilité du repost
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RepostVisibility {
    Public,
    Followers,
    Private,
}

/// Données de commentaires pour une track
#[derive(Debug, Clone, Default)]
pub struct CommentData {
    /// Total de commentaires
    pub total_count: u64,
    /// Commentaires par ordre chronologique
    pub comments: Vec<CommentEntry>,
    /// Index par timestamp pour commentaires temporels
    pub timed_comments: HashMap<u64, Vec<Uuid>>, // timestamp_ms -> comment_ids
}

/// Entry de commentaire individuel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommentEntry {
    pub id: Uuid,
    pub user_id: i64,
    pub track_id: Uuid,
    pub parent_id: Option<Uuid>, // Pour les réponses
    pub content: String,
    pub timestamp_ms: Option<u64>, // Pour commentaires temporels sur waveform
    pub created_at: SystemTime,
    pub likes_count: u32,
    pub replies_count: u32,
    pub edited: bool,
    pub edited_at: Option<SystemTime>,
}

/// Statistiques sociales d'un utilisateur
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct UserSocialStats {
    pub followers_count: u64,
    pub following_count: u64,
    pub tracks_count: u64,
    pub likes_count: u64,
    pub reposts_count: u64,
    pub comments_count: u64,
    pub total_plays: u64,
    pub total_likes_received: u64,
    pub total_reposts_received: u64,
    pub total_comments_received: u64,
}

/// Feed social personnalisé d'un utilisateur
#[derive(Debug, Clone)]
pub struct UserFeed {
    pub user_id: i64,
    pub items: Vec<FeedItem>,
    pub last_updated: SystemTime,
    pub has_more: bool,
    pub next_cursor: Option<String>,
}

/// Item dans le feed social
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FeedItem {
    pub id: Uuid,
    pub item_type: FeedItemType,
    pub created_at: SystemTime,
    pub relevance_score: f32, // 0.0 - 1.0 pour algorithme
}

/// Types d'items dans le feed
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum FeedItemType {
    /// Track uploadée par un utilisateur suivi
    TrackUploaded {
        user_id: i64,
        track_id: Uuid,
        track_title: String,
        track_artist: String,
    },
    /// Track repostée par un utilisateur suivi
    TrackReposted {
        user_id: i64,
        track_id: Uuid,
        repost: RepostEntry,
    },
    /// Playlist créée ou mise à jour
    PlaylistUpdated {
        user_id: i64,
        playlist_id: Uuid,
        playlist_name: String,
        tracks_added: u32,
    },
    /// Utilisateur a aimé une track
    TrackLiked {
        user_id: i64,
        track_id: Uuid,
        track_title: String,
    },
    /// Nouvel utilisateur suivi a rejoint
    UserJoined {
        user_id: i64,
        username: String,
        followed_by: Vec<i64>, // Utilisateurs en commun
    },
    /// Recommandation algorithmique
    RecommendedTrack {
        track_id: Uuid,
        reason: RecommendationReason,
        confidence: f32,
    },
}

/// Raisons de recommandation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum RecommendationReason {
    SimilarToLiked,
    PopularInGenre,
    FriendsAlsoLike,
    TrendingNow,
    BasedOnHistory,
}

/// Configuration du système social
#[derive(Debug, Clone)]
pub struct SocialConfig {
    pub max_following_per_user: usize,
    pub max_feed_items: usize,
    pub feed_cache_duration: Duration,
    pub enable_repost_notifications: bool,
    pub enable_like_notifications: bool,
    pub enable_comment_notifications: bool,
    pub max_comment_length: usize,
    pub enable_timed_comments: bool,
    pub rate_limit_follows_per_hour: u32,
    pub rate_limit_likes_per_minute: u32,
    pub rate_limit_comments_per_minute: u32,
}

/// Événements sociaux
#[derive(Debug, Clone)]
pub enum SocialEvent {
    /// Nouvel abonnement
    UserFollowed {
        follower_id: i64,
        followed_id: i64,
        timestamp: SystemTime,
    },
    /// Désabonnement
    UserUnfollowed {
        follower_id: i64,
        unfollowed_id: i64,
        timestamp: SystemTime,
    },
    /// Track likée
    TrackLiked {
        user_id: i64,
        track_id: Uuid,
        source: LikeSource,
        timestamp: SystemTime,
    },
    /// Track dislikée
    TrackUnliked {
        user_id: i64,
        track_id: Uuid,
        timestamp: SystemTime,
    },
    /// Track repostée
    TrackReposted {
        user_id: i64,
        repost: RepostEntry,
        timestamp: SystemTime,
    },
    /// Commentaire ajouté
    CommentAdded {
        comment: CommentEntry,
        timestamp: SystemTime,
    },
    /// Commentaire modifié
    CommentEdited {
        comment_id: Uuid,
        new_content: String,
        timestamp: SystemTime,
    },
    /// Commentaire supprimé
    CommentDeleted {
        comment_id: Uuid,
        user_id: i64,
        track_id: Uuid,
        timestamp: SystemTime,
    },
}

impl Default for SocialConfig {
    fn default() -> Self {
        Self {
            max_following_per_user: 5000,
            max_feed_items: 100,
            feed_cache_duration: Duration::from_secs(300), // 5 minutes
            enable_repost_notifications: true,
            enable_like_notifications: true,
            enable_comment_notifications: true,
            max_comment_length: 1000,
            enable_timed_comments: true,
            rate_limit_follows_per_hour: 200,
            rate_limit_likes_per_minute: 60,
            rate_limit_comments_per_minute: 10,
        }
    }
}

impl SocialManager {
    /// Crée un nouveau gestionnaire social
    pub fn new(config: SocialConfig) -> Self {
        let (event_sender, _) = broadcast::channel(10_000);
        
        Self {
            follow_graph: Arc::new(RwLock::new(FollowGraph::default())),
            track_likes: Arc::new(RwLock::new(HashMap::new())),
            track_reposts: Arc::new(RwLock::new(HashMap::new())),
            track_comments: Arc::new(RwLock::new(HashMap::new())),
            feed_cache: Arc::new(RwLock::new(HashMap::new())),
            config,
            event_sender,
        }
    }
    
    /// Suivre un utilisateur
    pub async fn follow_user(&self, follower_id: i64, followed_id: i64) -> Result<(), AppError> {
        if follower_id == followed_id {
            return Err(AppError::ValidationError("Cannot follow yourself".to_string()));
        }
        
        let mut graph = self.follow_graph.write().await;
        
        // Vérifier la limite de following
        let following_count = graph.following.get(&follower_id)
            .map(|s| s.len())
            .unwrap_or(0);
        
        if following_count >= self.config.max_following_per_user {
            return Err(AppError::ValidationError(format!(
                "Max following limit reached: {}", 
                self.config.max_following_per_user
            )));
        }
        
        // Ajouter la relation
        let following_set = graph.following.entry(follower_id).or_insert_with(HashSet::new);
        let was_new = following_set.insert(followed_id);
        
        if was_new {
            // Ajouter aux followers
            let followers_set = graph.followers.entry(followed_id).or_insert_with(HashSet::new);
            followers_set.insert(follower_id);
            
            // Mettre à jour les stats
            self.update_user_stats(&mut graph, follower_id, |stats| {
                stats.following_count += 1;
            });
            self.update_user_stats(&mut graph, followed_id, |stats| {
                stats.followers_count += 1;
            });
            
            // Invalider le cache du feed
            self.invalidate_feed_cache(follower_id).await;
            
            // Émettre l'événement
            let _ = self.event_sender.send(SocialEvent::UserFollowed {
                follower_id,
                followed_id,
                timestamp: SystemTime::now(),
            });
            
            info!("User {} now follows user {}", follower_id, followed_id);
        }
        
        Ok(())
    }
    
    /// Ne plus suivre un utilisateur
    pub async fn unfollow_user(&self, follower_id: i64, unfollowed_id: i64) -> Result<(), AppError> {
        let mut graph = self.follow_graph.write().await;
        
        // Retirer la relation
        let was_following = if let Some(following_set) = graph.following.get_mut(&follower_id) {
            following_set.remove(&unfollowed_id)
        } else {
            false
        };
        
        if was_following {
            // Retirer des followers
            if let Some(followers_set) = graph.followers.get_mut(&unfollowed_id) {
                followers_set.remove(&follower_id);
            }
            
            // Mettre à jour les stats
            self.update_user_stats(&mut graph, follower_id, |stats| {
                if stats.following_count > 0 {
                    stats.following_count -= 1;
                }
            });
            self.update_user_stats(&mut graph, unfollowed_id, |stats| {
                if stats.followers_count > 0 {
                    stats.followers_count -= 1;
                }
            });
            
            // Invalider le cache du feed
            self.invalidate_feed_cache(follower_id).await;
            
            // Émettre l'événement
            let _ = self.event_sender.send(SocialEvent::UserUnfollowed {
                follower_id,
                unfollowed_id: unfollowed_id,
                timestamp: SystemTime::now(),
            });
            
            info!("User {} unfollowed user {}", follower_id, unfollowed_id);
        }
        
        Ok(())
    }
    
    /// Aimer une track
    pub async fn like_track(
        &self, 
        user_id: i64, 
        track_id: Uuid, 
        source: LikeSource
    ) -> Result<(), AppError> {
        let mut likes = self.track_likes.write().await;
        let like_data = likes.entry(track_id).or_insert_with(LikeData::default);
        
        // Vérifier si déjà liké
        if like_data.liked_by.contains(&user_id) {
            return Ok(()); // Déjà liké
        }
        
        // Ajouter le like
        like_data.liked_by.insert(user_id);
        like_data.total_count += 1;
        like_data.like_timeline.push(LikeEntry {
            user_id,
            timestamp: SystemTime::now(),
            source: source.clone(),
        });
        
        // Mettre à jour les stats utilisateur
        {
            let mut graph = self.follow_graph.write().await;
            self.update_user_stats(&mut graph, user_id, |stats| {
                stats.likes_count += 1;
            });
        }
        
        // Émettre l'événement
        let _ = self.event_sender.send(SocialEvent::TrackLiked {
            user_id,
            track_id,
            source,
            timestamp: SystemTime::now(),
        });
        
        debug!("User {} liked track {}", user_id, track_id);
        Ok(())
    }
    
    /// Ne plus aimer une track
    pub async fn unlike_track(&self, user_id: i64, track_id: Uuid) -> Result<(), AppError> {
        let mut likes = self.track_likes.write().await;
        
        if let Some(like_data) = likes.get_mut(&track_id) {
            let was_liked = like_data.liked_by.remove(&user_id);
            
            if was_liked && like_data.total_count > 0 {
                like_data.total_count -= 1;
                
                // Mettre à jour les stats utilisateur
                {
                    let mut graph = self.follow_graph.write().await;
                    self.update_user_stats(&mut graph, user_id, |stats| {
                        if stats.likes_count > 0 {
                            stats.likes_count -= 1;
                        }
                    });
                }
                
                // Émettre l'événement
                let _ = self.event_sender.send(SocialEvent::TrackUnliked {
                    user_id,
                    track_id,
                    timestamp: SystemTime::now(),
                });
                
                debug!("User {} unliked track {}", user_id, track_id);
            }
        }
        
        Ok(())
    }
    
    /// Reposter une track
    pub async fn repost_track(
        &self,
        user_id: i64,
        track_id: Uuid,
        message: Option<String>,
        visibility: RepostVisibility,
    ) -> Result<Uuid, AppError> {
        let repost_id = Uuid::new_v4();
        let repost = RepostEntry {
            id: repost_id,
            user_id,
            track_id,
            message,
            timestamp: SystemTime::now(),
            visibility,
            likes_count: 0,
            comments_count: 0,
        };
        
        // Ajouter le repost
        {
            let mut reposts = self.track_reposts.write().await;
            let repost_data = reposts.entry(track_id).or_insert_with(RepostData::default);
            repost_data.reposts.push(repost.clone());
            repost_data.total_count += 1;
        }
        
        // Mettre à jour les stats utilisateur
        {
            let mut graph = self.follow_graph.write().await;
            self.update_user_stats(&mut graph, user_id, |stats| {
                stats.reposts_count += 1;
            });
        }
        
        // Invalider le cache des feeds des followers
        self.invalidate_followers_feed_cache(user_id).await;
        
        // Émettre l'événement
        let _ = self.event_sender.send(SocialEvent::TrackReposted {
            user_id,
            repost: repost.clone(),
            timestamp: SystemTime::now(),
        });
        
        info!("User {} reposted track {}", user_id, track_id);
        Ok(repost_id)
    }
    
    /// Ajouter un commentaire
    pub async fn add_comment(
        &self,
        user_id: i64,
        track_id: Uuid,
        content: String,
        parent_id: Option<Uuid>,
        timestamp_ms: Option<u64>,
    ) -> Result<Uuid, AppError> {
        // Valider le contenu
        if content.len() > self.config.max_comment_length {
            return Err(AppError::ValidationError(format!(
                "Comment too long: {} chars, max: {}", 
                content.len(), 
                self.config.max_comment_length
            )));
        }
        
        let comment_id = Uuid::new_v4();
        let comment = CommentEntry {
            id: comment_id,
            user_id,
            track_id,
            parent_id,
            content,
            timestamp_ms,
            created_at: SystemTime::now(),
            likes_count: 0,
            replies_count: 0,
            edited: false,
            edited_at: None,
        };
        
        // Ajouter le commentaire
        {
            let mut comments = self.track_comments.write().await;
            let comment_data = comments.entry(track_id).or_insert_with(CommentData::default);
            comment_data.comments.push(comment.clone());
            comment_data.total_count += 1;
            
            // Indexer les commentaires temporels
            if let Some(timestamp) = timestamp_ms {
                comment_data.timed_comments
                    .entry(timestamp)
                    .or_insert_with(Vec::new)
                    .push(comment_id);
            }
            
            // Mettre à jour le count des réponses si c'est une réponse
            if let Some(parent_id) = parent_id {
                for comment in &mut comment_data.comments {
                    if comment.id == parent_id {
                        comment.replies_count += 1;
                        break;
                    }
                }
            }
        }
        
        // Mettre à jour les stats utilisateur
        {
            let mut graph = self.follow_graph.write().await;
            self.update_user_stats(&mut graph, user_id, |stats| {
                stats.comments_count += 1;
            });
        }
        
        // Émettre l'événement
        let _ = self.event_sender.send(SocialEvent::CommentAdded {
            comment: comment.clone(),
            timestamp: SystemTime::now(),
        });
        
        debug!("User {} commented on track {}", user_id, track_id);
        Ok(comment_id)
    }
    
    /// Obtenir les statistiques sociales d'un utilisateur
    pub async fn get_user_stats(&self, user_id: i64) -> UserSocialStats {
        let graph = self.follow_graph.read().await;
        graph.user_stats.get(&user_id).cloned().unwrap_or_default()
    }
    
    /// Obtenir la liste des utilisateurs suivis
    pub async fn get_following(&self, user_id: i64) -> Vec<i64> {
        let graph = self.follow_graph.read().await;
        graph.following.get(&user_id)
            .map(|set| set.iter().copied().collect())
            .unwrap_or_default()
    }
    
    /// Obtenir la liste des followers
    pub async fn get_followers(&self, user_id: i64) -> Vec<i64> {
        let graph = self.follow_graph.read().await;
        graph.followers.get(&user_id)
            .map(|set| set.iter().copied().collect())
            .unwrap_or_default()
    }
    
    /// Obtenir les likes d'une track
    pub async fn get_track_likes(&self, track_id: Uuid) -> LikeData {
        let likes = self.track_likes.read().await;
        likes.get(&track_id).cloned().unwrap_or_default()
    }
    
    /// Obtenir les commentaires d'une track
    pub async fn get_track_comments(&self, track_id: Uuid, limit: Option<usize>) -> Vec<CommentEntry> {
        let comments = self.track_comments.read().await;
        
        if let Some(comment_data) = comments.get(&track_id) {
            let mut result = comment_data.comments.clone();
            // Trier par date de création (plus récents en premier)
            result.sort_by(|a, b| b.created_at.cmp(&a.created_at));
            
            if let Some(limit) = limit {
                result.truncate(limit);
            }
            
            result
        } else {
            Vec::new()
        }
    }
    
    /// Abonnement aux événements sociaux
    pub fn subscribe_events(&self) -> broadcast::Receiver<SocialEvent> {
        self.event_sender.subscribe()
    }
    
    /// Met à jour les stats d'un utilisateur
    fn update_user_stats<F>(&self, graph: &mut FollowGraph, user_id: i64, updater: F)
    where
        F: FnOnce(&mut UserSocialStats),
    {
        let stats = graph.user_stats.entry(user_id).or_insert_with(UserSocialStats::default);
        updater(stats);
    }
    
    /// Invalide le cache du feed d'un utilisateur
    async fn invalidate_feed_cache(&self, user_id: i64) {
        let mut cache = self.feed_cache.write().await;
        cache.remove(&user_id);
    }
    
    /// Invalide le cache des feeds des followers d'un utilisateur
    async fn invalidate_followers_feed_cache(&self, user_id: i64) {
        let followers = self.get_followers(user_id).await;
        let mut cache = self.feed_cache.write().await;
        
        for follower_id in followers {
            cache.remove(&follower_id);
        }
    }
} 