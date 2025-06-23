use std::{
    collections::HashMap,
    path::PathBuf,
    sync::Arc,
    time::{Duration, SystemTime},
};
use tokio::sync::RwLock;
use tracing::{info, debug};
use serde::{Serialize, Deserialize};
use crate::config::Config;
use sha2::{Sha256, Digest};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileMetadata {
    pub size: u64,
    pub modified: SystemTime,
    pub mime_type: String,
    pub etag: String,
    pub cached_at: SystemTime,
}

impl FileMetadata {
    pub fn new(size: u64, modified: SystemTime, mime_type: String) -> Self {
        let etag = format!("\"{}\"", 
            sha2::Sha256::digest(format!("{}-{}", size, modified.duration_since(SystemTime::UNIX_EPOCH).unwrap().as_secs()))
                .iter()
                .map(|b| format!("{:02x}", b))
                .collect::<String>()[..16]
                .to_string()
        );
        
        Self {
            size,
            modified,
            mime_type,
            etag,
            cached_at: SystemTime::now(),
        }
    }

    pub fn is_expired(&self, max_age: Duration) -> bool {
        self.cached_at.elapsed().unwrap_or(Duration::MAX) > max_age
    }

    pub fn is_still_valid(&self, current_modified: SystemTime) -> bool {
        self.modified == current_modified
    }
}

#[derive(Clone)]
pub struct FileCache {
    cache: Arc<RwLock<HashMap<PathBuf, FileMetadata>>>,
    max_age: Duration,
    max_entries: usize,
    last_cleanup: Arc<RwLock<SystemTime>>,
}

impl FileCache {
    pub fn new(max_age: Duration, max_entries: usize) -> Self {
        Self {
            cache: Arc::new(RwLock::new(HashMap::new())),
            max_age,
            max_entries,
            last_cleanup: Arc::new(RwLock::new(SystemTime::now())),
        }
    }

    pub async fn get(&self, path: &PathBuf) -> Option<FileMetadata> {
        let cache = self.cache.read().await;
        if let Some(metadata) = cache.get(path) {
            if !metadata.is_expired(self.max_age) {
                // Vérification additionnelle de la validité du fichier
                if let Ok(current_metadata) = tokio::fs::metadata(path).await {
                    if let Ok(current_modified) = current_metadata.modified() {
                        if metadata.is_still_valid(current_modified) {
                            debug!("Cache hit pour: {:?}", path);
                            return Some(metadata.clone());
                        }
                    }
                }
            }
        }
        debug!("Cache miss pour: {:?}", path);
        None
    }

    pub async fn set(&self, path: PathBuf, metadata: FileMetadata) {
        let mut cache = self.cache.write().await;
        
        // Limitation du nombre d'entrées
        if cache.len() >= self.max_entries {
            self.cleanup_cache(&mut cache).await;
        }

        cache.insert(path.clone(), metadata);
        debug!("Mise en cache pour: {:?}", path);
    }

    pub async fn invalidate(&self, path: &PathBuf) {
        let mut cache = self.cache.write().await;
        cache.remove(path);
        debug!("Invalidation du cache pour: {:?}", path);
    }

    pub async fn cleanup_expired(&self) {
        let now = SystemTime::now();
        let mut last_cleanup = self.last_cleanup.write().await;
        
        // Nettoyage toutes les 10 minutes
        if now.duration_since(*last_cleanup).unwrap_or_default() < Duration::from_secs(600) {
            return;
        }

        let mut cache = self.cache.write().await;
        let before_count = cache.len();
        
        self.cleanup_cache(&mut cache).await;
        
        let after_count = cache.len();
        if before_count > after_count {
            info!("Nettoyage du cache: {} -> {} entrées", before_count, after_count);
        }

        *last_cleanup = now;
    }

    async fn cleanup_cache(&self, cache: &mut HashMap<PathBuf, FileMetadata>) {
        // Suppression des entrées expirées
        cache.retain(|_, metadata| !metadata.is_expired(self.max_age));

        // Si encore trop d'entrées, suppression des plus anciennes
        if cache.len() > self.max_entries {
            let mut entries: Vec<_> = cache.iter().collect();
            entries.sort_by_key(|(_, metadata)| metadata.cached_at);
            
            let to_remove = cache.len() - self.max_entries + self.max_entries / 4; // Supprime 25% de plus
            let paths_to_remove: Vec<_> = entries.iter().take(to_remove).map(|(path, _)| (*path).clone()).collect();
            for path in paths_to_remove {
                cache.remove(&path);
            }
        }
    }

    pub async fn get_stats(&self) -> CacheStats {
        let cache = self.cache.read().await;
        let total_entries = cache.len();
        let expired_entries = cache.values()
            .filter(|metadata| metadata.is_expired(self.max_age))
            .count();

        CacheStats {
            total_entries,
            valid_entries: total_entries - expired_entries,
            expired_entries,
            max_entries: self.max_entries,
            max_age_seconds: self.max_age.as_secs(),
        }
    }
}

#[derive(Debug, Serialize)]
pub struct CacheStats {
    pub total_entries: usize,
    pub valid_entries: usize,
    pub expired_entries: usize,
    pub max_entries: usize,
    pub max_age_seconds: u64,
}

impl Default for FileCache {
    fn default() -> Self {
        Self::new(
            Duration::from_secs(3600), // Cache 1 heure
            1000, // Maximum 1000 entrées
        )
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::tempdir;
    use tokio::fs;

    #[tokio::test]
    async fn test_cache_basic_operations() {
        let cache = FileCache::default();
        let temp_dir = tempdir().unwrap();
        let file_path = temp_dir.path().join("test.mp3");
        
        // Création d'un fichier de test
        fs::write(&file_path, b"test content").await.unwrap();
        let file_metadata = fs::metadata(&file_path).await.unwrap();
        
        let metadata = FileMetadata::new(
            file_metadata.len(),
            file_metadata.modified().unwrap(),
            "audio/mpeg".to_string()
        );

        // Test de mise en cache
        cache.set(file_path.clone(), metadata.clone()).await;
        
        // Test de récupération
        let cached = cache.get(&file_path).await;
        assert!(cached.is_some());
        assert_eq!(cached.unwrap().size, metadata.size);
    }

    #[tokio::test]
    async fn test_cache_expiration() {
        let cache = FileCache::new(Duration::from_millis(100), 100);
        let temp_dir = tempdir().unwrap();
        let file_path = temp_dir.path().join("test.mp3");
        
        fs::write(&file_path, b"test content").await.unwrap();
        let file_metadata = fs::metadata(&file_path).await.unwrap();
        
        let metadata = FileMetadata::new(
            file_metadata.len(),
            file_metadata.modified().unwrap(),
            "audio/mpeg".to_string()
        );

        cache.set(file_path.clone(), metadata).await;
        
        // Immédiatement disponible
        assert!(cache.get(&file_path).await.is_some());
        
        // Attendre l'expiration
        tokio::time::sleep(Duration::from_millis(150)).await;
        
        // Devrait être expiré
        assert!(cache.get(&file_path).await.is_none());
    }
} 