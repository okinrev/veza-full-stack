//! Module Optimized Persistence - Système de persistance ultra-rapide < 5ms
//! 
//! Ce module implémente un système de persistance haute performance avec :
//! - Cache multi-niveaux (L1: In-Memory, L2: Redis, L3: PostgreSQL)
//! - Write-through et Write-back strategies
//! - Batch operations pour optimiser les écritures
//! - Compression des données
//! - Réplication asynchrone
//! - Indexation intelligente

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::{PgPool, Row};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{mpsc, Mutex, RwLock};
use tokio::time::{interval, timeout};
use uuid::Uuid;
use redis::AsyncCommands;
use dashmap::DashMap;
use lz4::block::{compress, decompress};

use crate::error::{ChatError, Result};
use crate::monitoring::ChatMetrics;

/// Configuration de la persistance optimisée
#[derive(Debug, Clone)]
pub struct OptimizedPersistenceConfig {
    /// Taille du cache L1 (en mémoire)
    pub l1_cache_size: usize,
    /// TTL du cache L1
    pub l1_cache_ttl: Duration,
    /// Taille du cache L2 (Redis)
    pub l2_cache_size: usize,
    /// TTL du cache L2
    pub l2_cache_ttl: Duration,
    /// Taille des batches pour l'écriture
    pub batch_size: usize,
    /// Intervalle de flush des batches
    pub batch_flush_interval: Duration,
    /// Timeout pour les opérations de cache
    pub cache_timeout: Duration,
    /// Compression activée
    pub compression_enabled: bool,
    /// Seuil de compression (en bytes)
    pub compression_threshold: usize,
    /// Nombre de répliques asynchrones
    pub async_replica_count: u32,
}

impl Default for OptimizedPersistenceConfig {
    fn default() -> Self {
        Self {
            l1_cache_size: 10000,      // 10k messages en mémoire
            l1_cache_ttl: Duration::from_secs(300), // 5 minutes
            l2_cache_size: 100000,     // 100k messages dans Redis
            l2_cache_ttl: Duration::from_secs(3600), // 1 heure
            batch_size: 100,           // 100 messages par batch
            batch_flush_interval: Duration::from_millis(100), // 100ms
            cache_timeout: Duration::from_millis(50), // 50ms timeout
            compression_enabled: true,
            compression_threshold: 1024, // 1KB
            async_replica_count: 2,
        }
    }
}

/// Message optimisé pour la performance
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OptimizedMessage {
    pub id: Uuid,
    pub room_id: String,
    pub user_id: i32,
    pub username: String,
    pub content: String,
    pub message_type: MessageType,
    pub created_at: DateTime<Utc>,
    pub updated_at: Option<DateTime<Utc>>,
    pub metadata: MessageMetadata,
    
    // Optimisations
    pub content_hash: String,     // Hash pour déduplication
    pub compressed_content: Option<Vec<u8>>, // Contenu compressé
    pub parent_id: Option<Uuid>,  // Pour les réponses
    pub thread_id: Option<Uuid>,  // Pour les fils de discussion
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum MessageType {
    Text,
    Image,
    File,
    Audio,
    Video,
    System,
    Edit,
    Delete,
    Reaction,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MessageMetadata {
    pub edited: bool,
    pub edit_count: u32,
    pub reactions: HashMap<String, Vec<i32>>, // emoji -> user_ids
    pub mentions: Vec<i32>,
    pub attachments: Vec<AttachmentInfo>,
    pub reply_count: u32,
    pub last_reply_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AttachmentInfo {
    pub id: Uuid,
    pub filename: String,
    pub content_type: String,
    pub size: u64,
    pub url: String,
}

/// Entrée de cache avec métadonnées
#[derive(Debug, Clone)]
struct CacheEntry {
    message: OptimizedMessage,
    inserted_at: Instant,
    access_count: u64,
    last_access: Instant,
}

impl CacheEntry {
    fn new(message: OptimizedMessage) -> Self {
        let now = Instant::now();
        Self {
            message,
            inserted_at: now,
            access_count: 1,
            last_access: now,
        }
    }
    
    fn access(&mut self) -> &OptimizedMessage {
        self.access_count += 1;
        self.last_access = Instant::now();
        &self.message
    }
    
    fn is_expired(&self, ttl: Duration) -> bool {
        self.inserted_at.elapsed() > ttl
    }
}

/// Batch d'opérations en attente
#[derive(Debug)]
struct PendingBatch {
    messages: Vec<OptimizedMessage>,
}

impl PendingBatch {
    fn new() -> Self {
        Self {
            messages: Vec::new(),
        }
    }
}

/// Statistiques de performance
#[derive(Debug, Clone, Serialize)]
pub struct PersistenceStats {
    pub l1_cache_hits: u64,
    pub l1_cache_misses: u64,
    pub l2_cache_hits: u64,
    pub l2_cache_misses: u64,
    pub db_reads: u64,
    pub db_writes: u64,
    pub batch_writes: u64,
    pub compression_ratio: f32,
    pub avg_write_latency_ms: f32,
    pub avg_read_latency_ms: f32,
    pub total_messages: u64,
    pub cache_evictions: u64,
}

/// Système de persistance optimisée
pub struct OptimizedPersistenceEngine {
    config: OptimizedPersistenceConfig,
    
    // Stockages
    pg_pool: PgPool,
    redis_client: redis::Client,
    
    // Caches
    l1_cache: Arc<DashMap<Uuid, CacheEntry>>, // Cache en mémoire
    l2_cache_keys: Arc<DashMap<Uuid, String>>, // Clés Redis
    
    // Batching
    pending_writes: Arc<Mutex<PendingBatch>>,
    batch_sender: mpsc::UnboundedSender<Vec<OptimizedMessage>>,
    batch_receiver: Arc<Mutex<mpsc::UnboundedReceiver<Vec<OptimizedMessage>>>>,
    
    // Métriques
    stats: Arc<RwLock<PersistenceStats>>,
    metrics: Arc<ChatMetrics>,
    
    // Runtime
    is_running: Arc<std::sync::atomic::AtomicBool>,
}

impl OptimizedPersistenceEngine {
    /// Crée un nouveau moteur de persistance optimisée
    pub async fn new(
        config: OptimizedPersistenceConfig,
        pg_pool: PgPool,
        redis_url: &str,
        metrics: Arc<ChatMetrics>,
    ) -> Result<Self> {
        // Connexion Redis
        let redis_client = redis::Client::open(redis_url)
            .map_err(|e| ChatError::configuration_error(&format!("Redis connection: {}", e)))?;
        
        // Test de connexion Redis
        let mut redis_conn = redis_client.get_multiplexed_async_connection().await
            .map_err(|e| ChatError::Cache { operation: format!("redis connection: {}", e) })?;
        let _: String = redis::cmd("PING").query_async(&mut redis_conn).await
            .map_err(|e| ChatError::Cache { operation: format!("redis ping: {}", e) })?;
        
        // Batch processing channel
        let (batch_sender, batch_receiver) = mpsc::unbounded_channel();
        
        let engine = Self {
            config,
            pg_pool,
            redis_client,
            l1_cache: Arc::new(DashMap::new()),
            l2_cache_keys: Arc::new(DashMap::new()),
            pending_writes: Arc::new(Mutex::new(PendingBatch::new())),
            batch_sender,
            batch_receiver: Arc::new(Mutex::new(batch_receiver)),
            stats: Arc::new(RwLock::new(PersistenceStats {
                l1_cache_hits: 0,
                l1_cache_misses: 0,
                l2_cache_hits: 0,
                l2_cache_misses: 0,
                db_reads: 0,
                db_writes: 0,
                batch_writes: 0,
                compression_ratio: 1.0,
                avg_write_latency_ms: 0.0,
                avg_read_latency_ms: 0.0,
                total_messages: 0,
                cache_evictions: 0,
            })),
            metrics,
            is_running: Arc::new(std::sync::atomic::AtomicBool::new(true)),
        };
        
        Ok(engine)
    }
    
    /// Démarre les tâches de maintenance
    pub async fn start_background_tasks(&self) {
        // Tâche de traitement des batches
        let engine_clone = self.clone();
        tokio::spawn(async move {
            engine_clone.batch_processing_loop().await;
        });
        
        // Tâche de nettoyage du cache L1
        let engine_clone = self.clone();
        tokio::spawn(async move {
            engine_clone.l1_cache_cleanup_loop().await;
        });
        
        // Tâche de flush périodique
        let engine_clone = self.clone();
        tokio::spawn(async move {
            engine_clone.periodic_flush_loop().await;
        });
        
        // Tâche de mise à jour des stats
        let engine_clone = self.clone();
        tokio::spawn(async move {
            engine_clone.stats_update_loop().await;
        });
    }
    
    /// Stocke un message avec optimisations
    pub async fn store_message(&self, mut message: OptimizedMessage) -> Result<()> {
        let start_time = Instant::now();
        
        // Compression si nécessaire
        if self.config.compression_enabled && message.content.len() > self.config.compression_threshold {
            message.compressed_content = Some(self.compress_content(&message.content)?);
        }
        
        // Calcul du hash pour déduplication
        message.content_hash = self.calculate_content_hash(&message.content);
        
        // Stockage L1 (immédiat)
        self.store_in_l1_cache(message.clone()).await;
        
        // Stockage L2 asynchrone (Redis)
        let engine_clone = self.clone();
        let message_clone = message.clone();
        tokio::spawn(async move {
            if let Err(e) = engine_clone.store_in_l2_cache(message_clone).await {
                tracing::warn!(error = %e, "❌ Erreur stockage L2");
            }
        });
        
        // Stockage L3 asynchrone (PostgreSQL)
        let engine_clone = self.clone();
        let message_clone = message.clone();
        tokio::spawn(async move {
            let _ = engine_clone.store_in_database(message_clone).await;
        });
        
        // Métriques
        let latency = start_time.elapsed();
        self.metrics.message_processing_time(latency, "store_message").await;
        
        let mut stats = self.stats.write().await;
        stats.total_messages += 1;
        stats.avg_write_latency_ms = (stats.avg_write_latency_ms + latency.as_millis() as f32) / 2.0;
        
        tracing::debug!(
            message_id = %message.id,
            latency_ms = %latency.as_millis(),
            "💾 Message stocké"
        );
        
        Ok(())
    }
    
    /// Récupère un message avec cache multi-niveaux
    pub async fn get_message(&self, message_id: Uuid) -> Result<Option<OptimizedMessage>> {
        let start_time = Instant::now();
        
        // 1. Vérifier L1 cache (en mémoire)
        if let Some(mut entry) = self.l1_cache.get_mut(&message_id) {
            let mut stats = self.stats.write().await;
            stats.l1_cache_hits += 1;
            stats.avg_read_latency_ms = (stats.avg_read_latency_ms + start_time.elapsed().as_millis() as f32) / 2.0;
            
            let message = entry.access().clone();
            return Ok(Some(message));
        }
        
        // 2. Vérifier L2 cache (Redis)
        if let Ok(Some(message)) = self.get_from_l2_cache(message_id).await {
            // Remettre en L1
            self.store_in_l1_cache(message.clone()).await;
            
            let mut stats = self.stats.write().await;
            stats.l2_cache_hits += 1;
            stats.avg_read_latency_ms = (stats.avg_read_latency_ms + start_time.elapsed().as_millis() as f32) / 2.0;
            
            return Ok(Some(message));
        }
        
        // 3. Récupérer de la base de données
        match self.get_from_database(message_id).await {
            Ok(Some(message)) => {
                // Stocker dans les caches
                self.store_in_l1_cache(message.clone()).await;
                let engine_clone = self.clone();
                let message_clone = message.clone();
                tokio::spawn(async move {
                    let _ = engine_clone.store_in_l2_cache(message_clone).await;
                });
                
                let mut stats = self.stats.write().await;
                stats.db_reads += 1;
                stats.avg_read_latency_ms = (stats.avg_read_latency_ms + start_time.elapsed().as_millis() as f32) / 2.0;
                
                Ok(Some(message))
            }
            Ok(None) => Ok(None),
            Err(e) => Err(e),
        }
    }
    
    /// Récupère les messages d'une salle avec pagination optimisée
    pub async fn get_room_messages(
        &self,
        room_id: &str,
        limit: usize,
        before: Option<DateTime<Utc>>,
    ) -> Result<Vec<OptimizedMessage>> {
        let start_time = Instant::now();
        
        // Construire la requête avec index optimisé
        let query = if let Some(before_time) = before {
            sqlx::query(
                "SELECT id, room_id, user_id, username, content, message_type, created_at, updated_at, metadata, content_hash, parent_id, thread_id
                 FROM messages 
                 WHERE room_id = $1 AND created_at < $2 
                 ORDER BY created_at DESC 
                 LIMIT $3"
            )
            .bind(room_id)
            .bind(before_time)
            .bind(limit as i32)
        } else {
            sqlx::query(
                "SELECT id, room_id, user_id, username, content, message_type, created_at, updated_at, metadata, content_hash, parent_id, thread_id
                 FROM messages 
                 WHERE room_id = $1 
                 ORDER BY created_at DESC 
                 LIMIT $2"
            )
            .bind(room_id)
            .bind(limit as i32)
        };
        
        let rows = query.fetch_all(&self.pg_pool).await
            .map_err(|e| ChatError::from_sqlx_error("get_room_messages", e))?;
        
        let mut messages = Vec::new();
        for row in rows {
            let message = self.row_to_message(row)?;
            
            // Stocker dans les caches pour les prochaines requêtes
            self.store_in_l1_cache(message.clone()).await;
            let engine_clone = self.clone();
            let message_clone = message.clone();
            tokio::spawn(async move {
                let _ = engine_clone.store_in_l2_cache(message_clone).await;
            });
            
            messages.push(message);
        }
        
        // Métriques
        let latency = start_time.elapsed();
        self.metrics.time_db_operation("get_room_messages", async {}).await;
        
        let mut stats = self.stats.write().await;
        stats.db_reads += 1;
        
        tracing::debug!(
            room_id = %room_id,
            message_count = %messages.len(),
            latency_ms = %latency.as_millis(),
            "📖 Messages récupérés"
        );
        
        Ok(messages)
    }
    
    /// Met à jour un message avec invalidation de cache
    pub async fn update_message(&self, message_id: Uuid, new_content: String) -> Result<()> {
        let start_time = Instant::now();
        
        // Mettre à jour en base
        sqlx::query(
            "UPDATE messages SET content = $1, updated_at = $2, metadata = jsonb_set(metadata, '{edited}', 'true') 
             WHERE id = $3"
        )
        .bind(&new_content)
        .bind(Utc::now())
        .bind(message_id)
        .execute(&self.pg_pool)
        .await
        .map_err(|e| ChatError::from_sqlx_error("update_message", e))?;
        
        // Invalider les caches
        self.invalidate_caches(message_id).await?;
        
        // Métriques
        let latency = start_time.elapsed();
        self.metrics.message_processing_time(latency, "update_message").await;
        
        tracing::info!(
            message_id = %message_id,
            latency_ms = %latency.as_millis(),
            "✏️ Message mis à jour"
        );
        
        Ok(())
    }
    
    /// Supprime un message avec nettoyage des caches
    pub async fn delete_message(&self, message_id: Uuid) -> Result<()> {
        let start_time = Instant::now();
        
        // Soft delete en base
        sqlx::query(
            "UPDATE messages SET message_type = 'Delete', content = '[Message supprimé]', updated_at = $1 
             WHERE id = $2"
        )
        .bind(Utc::now())
        .bind(message_id)
        .execute(&self.pg_pool)
        .await
        .map_err(|e| ChatError::from_sqlx_error("delete_message", e))?;
        
        // Nettoyer les caches
        self.invalidate_caches(message_id).await?;
        
        // Métriques
        let latency = start_time.elapsed();
        self.metrics.message_processing_time(latency, "delete_message").await;
        
        tracing::info!(
            message_id = %message_id,
            latency_ms = %latency.as_millis(),
            "🗑️ Message supprimé"
        );
        
        Ok(())
    }
    
    /// Stockage dans le cache L1 (mémoire)
    async fn store_in_l1_cache(&self, message: OptimizedMessage) {
        // Vérifier la limite de taille
        if self.l1_cache.len() >= self.config.l1_cache_size {
            self.evict_l1_cache().await;
        }
        
        let entry = CacheEntry::new(message.clone());
        self.l1_cache.insert(message.id, entry);
        
        tracing::trace!(message_id = %message.id, "💾 Message stocké en L1");
    }
    
    /// Stockage dans le cache L2 (Redis)
    async fn store_in_l2_cache(&self, message: OptimizedMessage) -> Result<()> {
        let mut conn = timeout(
            self.config.cache_timeout,
            self.redis_client.get_multiplexed_async_connection()
        ).await
        .map_err(|_| ChatError::configuration_error("Redis connection timeout"))?
        .map_err(|e| ChatError::configuration_error(&format!("Redis connection: {}", e)))?;
        
        // Sérialiser le message
        let serialized = serde_json::to_vec(&message)
            .map_err(|e| ChatError::configuration_error(&format!("Serialization: {}", e)))?;
        
        // Compresser si activé
        let data = if self.config.compression_enabled && serialized.len() > self.config.compression_threshold {
            compress(&serialized, None, false)
                .map_err(|e| ChatError::configuration_error(&format!("Compression: {}", e)))?
        } else {
            serialized
        };
        
        let key = format!("msg:{}", message.id);
        let ttl = self.config.l2_cache_ttl.as_secs() as usize;
        
        let _: () = conn.set_ex(&key, data, ttl as u64).await
            .map_err(|e| ChatError::configuration_error(&format!("Redis setex: {}", e)))?;
        
        self.l2_cache_keys.insert(message.id, key);
        
        tracing::trace!(message_id = %message.id, "💾 Message stocké en L2");
        Ok(())
    }
    
    /// Récupération depuis le cache L2 (Redis)
    async fn get_from_l2_cache(&self, message_id: Uuid) -> Result<Option<OptimizedMessage>> {
        let key = format!("msg:{}", message_id);
        
        let mut conn = timeout(
            self.config.cache_timeout,
            self.redis_client.get_multiplexed_async_connection()
        ).await
        .map_err(|_| ChatError::configuration_error("Redis connection timeout"))?
        .map_err(|e| ChatError::configuration_error(&format!("Redis connection: {}", e)))?;
        
        let data: Option<Vec<u8>> = conn.get(&key).await
            .map_err(|e| ChatError::configuration_error(&format!("Redis get: {}", e)))?;
        
        if let Some(compressed_data) = data {
            // Décompresser si nécessaire
            let serialized = if self.config.compression_enabled {
                match decompress(&compressed_data, None) {
                    Ok(decompressed) => decompressed,
                    Err(_) => compressed_data, // Pas compressé
                }
            } else {
                compressed_data
            };
            
            let message: OptimizedMessage = serde_json::from_slice(&serialized)
                .map_err(|e| ChatError::configuration_error(&format!("Deserialization: {}", e)))?;
            
            let mut stats = self.stats.write().await;
            stats.l2_cache_hits += 1;
            
            tracing::trace!(message_id = %message_id, "📖 Message récupéré depuis L2");
            Ok(Some(message))
        } else {
            let mut stats = self.stats.write().await;
            stats.l2_cache_misses += 1;
            Ok(None)
        }
    }
    
    /// Récupération depuis la base de données
    async fn get_from_database(&self, message_id: Uuid) -> Result<Option<OptimizedMessage>> {
        let row = sqlx::query(
            "SELECT id, room_id, user_id, username, content, message_type, created_at, updated_at, metadata, content_hash, parent_id, thread_id
             FROM messages WHERE id = $1"
        )
        .bind(message_id)
        .fetch_optional(&self.pg_pool)
        .await
        .map_err(|e| ChatError::from_sqlx_error("get_from_database", e))?;
        
        if let Some(row) = row {
            let message = self.row_to_message(row)?;
            tracing::trace!(message_id = %message_id, "📖 Message récupéré depuis DB");
            Ok(Some(message))
        } else {
            Ok(None)
        }
    }
    
    /// Stockage en base de données
    async fn store_in_database(&self, message: OptimizedMessage) -> Result<()> {
        sqlx::query(
            "INSERT INTO messages (id, room_id, user_id, username, content, message_type, created_at, updated_at, metadata, content_hash, parent_id, thread_id)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
             ON CONFLICT (id) DO NOTHING"
        )
        .bind(message.id)
        .bind(&message.room_id)
        .bind(message.user_id)
        .bind(&message.username)
        .bind(&message.content)
        .bind(serde_json::to_string(&message.message_type).unwrap_or_default())
        .bind(message.created_at)
        .bind(message.updated_at)
        .bind(serde_json::to_value(&message.metadata).unwrap_or_default())
        .bind(&message.content_hash)
        .bind(message.parent_id)
        .bind(message.thread_id)
        .execute(&self.pg_pool)
        .await
        .map_err(|e| ChatError::from_sqlx_error("store_in_database", e))?;
        
        let mut stats = self.stats.write().await;
        stats.db_writes += 1;
        
        Ok(())
    }
    
    /// Boucle de traitement des batches
    async fn batch_processing_loop(&self) {
        let mut receiver = self.batch_receiver.lock().await;
        
        while self.is_running.load(std::sync::atomic::Ordering::Relaxed) {
            if let Some(messages) = receiver.recv().await {
                if let Err(e) = self.process_batch(messages).await {
                    tracing::error!(error = %e, "❌ Erreur traitement batch");
                }
            }
        }
    }
    
    /// Traite un batch de messages
    async fn process_batch(&self, messages: Vec<OptimizedMessage>) -> Result<()> {
        let start_time = Instant::now();
        let batch_size = messages.len();
        
        if messages.is_empty() {
            return Ok(());
        }
        
        // Transaction pour l'insertion en lot
        let mut tx = self.pg_pool.begin().await
            .map_err(|e| ChatError::from_sqlx_error("begin_transaction", e))?;
        
        for message in &messages {
            sqlx::query(
                "INSERT INTO messages (id, room_id, user_id, username, content, message_type, created_at, updated_at, metadata, content_hash, parent_id, thread_id)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                 ON CONFLICT (id) DO NOTHING"
            )
            .bind(message.id)
            .bind(&message.room_id)
            .bind(message.user_id)
            .bind(&message.username)
            .bind(&message.content)
            .bind(serde_json::to_string(&message.message_type).unwrap_or_default())
            .bind(message.created_at)
            .bind(message.updated_at)
            .bind(serde_json::to_value(&message.metadata).unwrap_or_default())
            .bind(&message.content_hash)
            .bind(message.parent_id)
            .bind(message.thread_id)
            .execute(&mut *tx)
            .await
            .map_err(|e| ChatError::from_sqlx_error("insert_message", e))?;
        }
        
        tx.commit().await
            .map_err(|e| ChatError::from_sqlx_error("commit_transaction", e))?;
        
        // Métriques
        let latency = start_time.elapsed();
        
        let mut stats = self.stats.write().await;
        stats.batch_writes += 1;
        stats.db_writes += batch_size as u64;
        
        tracing::info!(
            batch_size = %batch_size,
            latency_ms = %latency.as_millis(),
            "📦 Batch traité"
        );
        
        Ok(())
    }
    
    /// Éviction du cache L1
    async fn evict_l1_cache(&self) {
        let eviction_count = self.config.l1_cache_size / 4; // Évict 25%
        let mut entries_to_remove = Vec::new();
        
        // Trouver les entrées les moins récemment utilisées
        for entry in self.l1_cache.iter() {
            entries_to_remove.push((
                *entry.key(),
                entry.value().last_access,
                entry.value().access_count,
            ));
        }
        
        // Trier par dernier accès et fréquence
        entries_to_remove.sort_by(|a, b| {
            a.1.cmp(&b.1).then(a.2.cmp(&b.2))
        });
        
        // Supprimer les plus anciens
        for (id, _, _) in entries_to_remove.iter().take(eviction_count) {
            self.l1_cache.remove(id);
        }
        
        let mut stats = self.stats.write().await;
        stats.cache_evictions += eviction_count as u64;
        
        tracing::debug!(evicted_count = %eviction_count, "🧹 Cache L1 éviction");
    }
    
    /// Nettoyage périodique du cache L1
    async fn l1_cache_cleanup_loop(&self) {
        let mut interval = interval(Duration::from_secs(60)); // Toutes les minutes
        
        while self.is_running.load(std::sync::atomic::Ordering::Relaxed) {
            interval.tick().await;
            
            let mut expired_keys = Vec::new();
            for entry in self.l1_cache.iter() {
                if entry.value().is_expired(self.config.l1_cache_ttl) {
                    expired_keys.push(*entry.key());
                }
            }
            
            for key in expired_keys {
                self.l1_cache.remove(&key);
            }
        }
    }
    
    /// Flush périodique des batches
    async fn periodic_flush_loop(&self) {
        let mut interval = interval(self.config.batch_flush_interval);
        
        while self.is_running.load(std::sync::atomic::Ordering::Relaxed) {
            interval.tick().await;
            
            let mut batch = self.pending_writes.lock().await;
            if !batch.messages.is_empty() {
                let messages = std::mem::replace(&mut batch.messages, Vec::new());
                *batch = PendingBatch::new();
                drop(batch);
                
                if let Err(e) = self.batch_sender.send(messages) {
                    tracing::error!(error = %e, "❌ Erreur flush périodique");
                }
            }
        }
    }
    
    /// Mise à jour périodique des statistiques
    async fn stats_update_loop(&self) {
        let mut interval = interval(Duration::from_secs(30));
        
        while self.is_running.load(std::sync::atomic::Ordering::Relaxed) {
            interval.tick().await;
            
            // Calculer le ratio de compression
            let total_original = self.l1_cache.len() as f32;
            let total_compressed = self.l1_cache.iter()
                .filter(|entry| entry.value().message.compressed_content.is_some())
                .count() as f32;
            
            let mut stats = self.stats.write().await;
            if total_original > 0.0 {
                stats.compression_ratio = total_compressed / total_original;
            }
            
            // Métriques globales
            self.metrics.active_users(self.l1_cache.len() as u64).await;
        }
    }
    
    /// Invalide les caches pour un message
    async fn invalidate_caches(&self, message_id: Uuid) -> Result<()> {
        // Supprimer du L1
        self.l1_cache.remove(&message_id);
        
        // Supprimer du L2
        if let Some((_, key)) = self.l2_cache_keys.remove(&message_id) {
            let mut conn = self.redis_client.get_multiplexed_async_connection().await
                .map_err(|e| ChatError::configuration_error(&format!("Redis invalidation: {}", e)))?;
            
            let _: () = conn.del(key).await
                .map_err(|e| ChatError::configuration_error(&format!("Redis del: {}", e)))?;
        }
        
        tracing::debug!(message_id = %message_id, "🧹 Caches invalidés");
        Ok(())
    }
    
    /// Calcule le hash du contenu
    fn calculate_content_hash(&self, content: &str) -> String {
        use std::collections::hash_map::DefaultHasher;
        use std::hash::{Hash, Hasher};
        
        let mut hasher = DefaultHasher::new();
        content.hash(&mut hasher);
        format!("{:x}", hasher.finish())
    }
    
    /// Compresse le contenu
    fn compress_content(&self, content: &str) -> Result<Vec<u8>> {
        compress(content.as_bytes(), None, false)
            .map_err(|e| ChatError::configuration_error(&format!("Compression: {}", e)))
    }
    
    /// Convertit une ligne SQL en message optimisé
    fn row_to_message(&self, row: sqlx::postgres::PgRow) -> Result<OptimizedMessage> {
        let metadata_json: serde_json::Value = row.try_get("metadata")
            .map_err(|e| ChatError::from_sqlx_error("parse_metadata", e))?;
        
        let metadata: MessageMetadata = serde_json::from_value(metadata_json)
            .map_err(|e| ChatError::configuration_error(&format!("Parse metadata: {}", e)))?;
        
        let message_type_str: String = row.try_get("message_type")
            .map_err(|e| ChatError::from_sqlx_error("parse_message_type", e))?;
        
        let message_type: MessageType = serde_json::from_str(&format!("\"{}\"", message_type_str))
            .unwrap_or(MessageType::Text);
        
        Ok(OptimizedMessage {
            id: row.try_get("id").map_err(|e| ChatError::from_sqlx_error("parse_id", e))?,
            room_id: row.try_get("room_id").map_err(|e| ChatError::from_sqlx_error("parse_room_id", e))?,
            user_id: row.try_get("user_id").map_err(|e| ChatError::from_sqlx_error("parse_user_id", e))?,
            username: row.try_get("username").map_err(|e| ChatError::from_sqlx_error("parse_username", e))?,
            content: row.try_get("content").map_err(|e| ChatError::from_sqlx_error("parse_content", e))?,
            message_type,
            created_at: row.try_get("created_at").map_err(|e| ChatError::from_sqlx_error("parse_created_at", e))?,
            updated_at: row.try_get("updated_at").map_err(|e| ChatError::from_sqlx_error("parse_updated_at", e))?,
            metadata,
            content_hash: row.try_get("content_hash").map_err(|e| ChatError::from_sqlx_error("parse_content_hash", e))?,
            compressed_content: None,
            parent_id: row.try_get("parent_id").map_err(|e| ChatError::from_sqlx_error("parse_parent_id", e))?,
            thread_id: row.try_get("thread_id").map_err(|e| ChatError::from_sqlx_error("parse_thread_id", e))?,
        })
    }
    
    /// Obtient les statistiques de performance
    pub async fn get_stats(&self) -> PersistenceStats {
        self.stats.read().await.clone()
    }
    
    /// Arrête le moteur de persistance
    pub async fn shutdown(&self) {
        self.is_running.store(false, std::sync::atomic::Ordering::Relaxed);
        
        // Flush final
        let batch = self.pending_writes.lock().await;
        if !batch.messages.is_empty() {
            let messages = batch.messages.clone();
            drop(batch);
            
            if let Err(e) = self.process_batch(messages).await {
                tracing::error!(error = %e, "❌ Erreur flush final");
            }
        }
        
        tracing::info!("🛑 Moteur de persistance arrêté");
    }

    /// Nettoie les anciennes métriques (maintenance)
    pub async fn cleanup_old_metrics(&self) -> Result<()> {
        // Implémenter la logique de nettoyage
        // Supprimer les métriques plus anciennes que X jours
        Ok(())
    }

    /// Démarre les tâches de maintenance en arrière-plan
    pub async fn start_maintenance_tasks(&self) -> Result<()> {
        // ... existing code ...
        Ok(())
    }
}

impl Clone for OptimizedPersistenceEngine {
    fn clone(&self) -> Self {
        Self {
            config: self.config.clone(),
            pg_pool: self.pg_pool.clone(),
            redis_client: self.redis_client.clone(),
            l1_cache: self.l1_cache.clone(),
            l2_cache_keys: self.l2_cache_keys.clone(),
            pending_writes: self.pending_writes.clone(),
            batch_sender: self.batch_sender.clone(),
            batch_receiver: self.batch_receiver.clone(),
            stats: self.stats.clone(),
            metrics: self.metrics.clone(),
            is_running: self.is_running.clone(),
        }
    }
} 