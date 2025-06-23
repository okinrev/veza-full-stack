# Audio Module Documentation

Le module audio fournit des fonctionnalités avancées de traitement et de compression audio pour le serveur de streaming.

## Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Modules](#modules)
  - [Audio Processing](#audio-processing)
  - [Audio Compression](#audio-compression)
- [Types et Structures](#types-et-structures)
- [API Reference](#api-reference)
- [Exemples d'utilisation](#exemples-dutilisation)
- [Intégration](#intégration)

## Vue d'ensemble

Le module audio se compose de deux sous-modules principaux :
- **processing** : Extraction de métadonnées, génération de waveforms, analyse spectrale
- **compression** : Compression audio adaptative avec profils de qualité multiples

## Modules

### Audio Processing

#### AudioProcessor

```rust
pub struct AudioProcessor {
    config: Arc<Config>,
    metadata_cache: Arc<RwLock<HashMap<PathBuf, AudioMetadata>>>,
    waveform_cache: Arc<RwLock<HashMap<PathBuf, WaveformData>>>,
}
```

**Fonctionnalités :**
- Extraction automatique de métadonnées audio
- Génération de waveforms pour visualisation
- Analyse spectrale FFT
- Cache intelligent des résultats
- Support multi-formats (MP3, FLAC, WAV, AAC, OGG)

#### AudioMetadata

```rust
pub struct AudioMetadata {
    pub duration_seconds: f64,
    pub sample_rate: u32,
    pub channels: u32,
    pub bitrate_kbps: Option<u32>,
    pub codec: String,
    pub title: Option<String>,
    pub artist: Option<String>,
    pub album: Option<String>,
    pub year: Option<u32>,
    pub genre: Option<String>,
    pub artwork_available: bool,
    pub file_size: u64,
    pub last_modified: SystemTime,
}
```

#### WaveformData

```rust
pub struct WaveformData {
    pub peaks: Vec<f32>,        // Amplitude peaks pour visualisation
    pub rms: Vec<f32>,          // RMS values pour rendu smooth
    pub sample_rate: u32,       // Taux d'échantillonnage
    pub duration_ms: u32,       // Durée en millisecondes
    pub generated_at: SystemTime,
}
```

#### AudioQuality

```rust
pub struct AudioQuality {
    pub bitrate_kbps: u32,
    pub sample_rate: u32,
    pub channels: u32,
    pub codec: String,
}
```

**Profils prédéfinis :**
- `high_quality()` : 320kbps, 44.1kHz, Stéréo
- `medium_quality()` : 192kbps, 44.1kHz, Stéréo  
- `low_quality()` : 128kbps, 22.05kHz, Stéréo

### Audio Compression

#### CompressionEngine

```rust
pub struct CompressionEngine {
    config: Arc<Config>,
    profiles: HashMap<String, CompressionProfile>,
    active_jobs: Arc<RwLock<HashMap<String, CompressionJob>>>,
    job_queue: Arc<RwLock<Vec<String>>>,
    stats: Arc<RwLock<CompressionStats>>,
    worker_count: usize,
}
```

**Caractéristiques :**
- Compression asynchrone avec workers parallèles
- Support de multiples codecs (MP3, AAC, OGG, OPUS, FLAC, WAV)
- Profils de compression personnalisables
- Système de queue avec priorités
- Suivi de progression en temps réel
- Statistiques détaillées

#### CompressionProfile

```rust
pub struct CompressionProfile {
    pub name: String,
    pub codec: AudioCodec,
    pub bitrate_kbps: u32,
    pub sample_rate: u32,
    pub channels: u8,
    pub quality_factor: f32,
    pub compression_level: u8,
    pub target_size_reduction: f32,
}
```

**Profils intégrés :**
- `ultra_high` : FLAC, 1411kbps, 30% réduction
- `high` : AAC, 320kbps, 50% réduction
- `medium` : MP3, 192kbps, 70% réduction
- `low` : MP3, 128kbps, 80% réduction
- `mobile` : OPUS, 96kbps, 85% réduction
- `podcast` : OPUS mono, 64kbps, 90% réduction

#### CompressionJob

```rust
pub struct CompressionJob {
    pub id: String,
    pub input_path: PathBuf,
    pub output_path: PathBuf,
    pub profile: CompressionProfile,
    pub status: JobStatus,
    pub progress: f32,
    pub created_at: SystemTime,
    pub started_at: Option<SystemTime>,
    pub completed_at: Option<SystemTime>,
    pub error_message: Option<String>,
    pub original_size_bytes: u64,
    pub compressed_size_bytes: Option<u64>,
    pub compression_ratio: Option<f32>,
}
```

#### JobStatus

```rust
pub enum JobStatus {
    Pending,
    InProgress,
    Completed,
    Failed,
    Cancelled,
}
```

## Types et Structures

### AudioCodec

```rust
pub enum AudioCodec {
    MP3,    // MPEG-1/2 Audio Layer III
    AAC,    // Advanced Audio Coding
    OGG,    // Ogg Vorbis
    OPUS,   // Opus (optimisé pour streaming)
    FLAC,   // Free Lossless Audio Codec
    WAV,    // Waveform Audio File Format
}
```

### CompressionRequest

```rust
pub struct CompressionRequest {
    pub input_file: String,
    pub target_quality: String,
    pub preserve_metadata: bool,
    pub async_processing: bool,
}
```

### CompressionResponse

```rust
pub struct CompressionResponse {
    pub job_id: String,
    pub status: JobStatus,
    pub estimated_completion_time: Option<u64>,
    pub download_url: Option<String>,
}
```

### CompressionStats

```rust
pub struct CompressionStats {
    pub total_jobs: u64,
    pub completed_jobs: u64,
    pub failed_jobs: u64,
    pub average_compression_ratio: f32,
    pub total_space_saved_mb: u64,
    pub processing_queue_size: usize,
    pub average_processing_time_ms: u64,
}
```

## API Reference

### AudioProcessor Methods

#### `new(config: Arc<Config>) -> Self`
Crée une nouvelle instance d'AudioProcessor.

#### `extract_metadata(&self, file_path: &Path) -> Result<AudioMetadata, Error>`
Extrait les métadonnées d'un fichier audio avec mise en cache.

**Exemple :**
```rust
let metadata = processor.extract_metadata(Path::new("audio/track.mp3")).await?;
println!("Durée: {:.2}s, Artiste: {:?}", metadata.duration_seconds, metadata.artist);
```

#### `generate_waveform(&self, file_path: &Path, resolution: usize) -> Result<WaveformData, Error>`
Génère les données de waveform pour visualisation.

**Paramètres :**
- `file_path` : Chemin vers le fichier audio
- `resolution` : Nombre de points de données (recommandé: 1000-4000)

#### `analyze_spectrum(&self, file_path: &Path, fft_size: usize) -> Result<Vec<f32>, Error>`
Analyse le spectre fréquentiel du fichier audio.

#### `transcode_quality(&self, input_path: &Path, output_path: &Path, quality: AudioQuality) -> Result<(), Error>`
Transcode un fichier vers une qualité différente.

#### `cleanup_caches(&self, max_age_hours: u64)`
Nettoie les caches expirés.

#### `get_cache_stats(&self) -> serde_json::Value`
Retourne les statistiques des caches.

### CompressionEngine Methods

#### `new(config: Arc<Config>) -> Self`
Crée une nouvelle instance du moteur de compression.

#### `start_workers(&self)`
Démarre les workers de compression en arrière-plan.

#### `compress_audio(&self, request: CompressionRequest) -> Result<CompressionResponse, CompressionError>`
Lance une tâche de compression audio.

**Exemple synchrone :**
```rust
let request = CompressionRequest {
    input_file: "track.mp3".to_string(),
    target_quality: "medium".to_string(),
    preserve_metadata: true,
    async_processing: false,
};

let response = engine.compress_audio(request).await?;
```

**Exemple asynchrone :**
```rust
let request = CompressionRequest {
    input_file: "track.mp3".to_string(),
    target_quality: "mobile".to_string(),
    preserve_metadata: true,
    async_processing: true,
};

let response = engine.compress_audio(request).await?;
// Vérifier le statut plus tard avec get_job_status()
```

#### `get_job_status(&self, job_id: &str) -> Option<CompressionJob>`
Récupère le statut d'une tâche de compression.

#### `cancel_job(&self, job_id: &str) -> Result<(), CompressionError>`
Annule une tâche en cours ou en attente.

#### `get_compression_stats(&self) -> CompressionStats`
Retourne les statistiques de compression.

#### `list_profiles(&self) -> Vec<CompressionProfile>`
Liste tous les profils de compression disponibles.

#### `add_custom_profile(&mut self, name: String, profile: CompressionProfile)`
Ajoute un profil de compression personnalisé.

#### `cleanup_completed_jobs(&self, max_age: Duration)`
Nettoie les tâches terminées anciennes.

#### `get_queue_info(&self) -> serde_json::Value`
Retourne des informations sur la queue de compression.

## Exemples d'utilisation

### Extraction de métadonnées

```rust
use stream_server::audio::processing::{AudioProcessor, AudioQuality};

async fn example_metadata_extraction() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let processor = AudioProcessor::new(config);
    
    // Extraire métadonnées
    let metadata = processor.extract_metadata(Path::new("audio/song.mp3")).await?;
    
    println!("📀 Titre: {:?}", metadata.title);
    println!("🎤 Artiste: {:?}", metadata.artist);
    println!("⏱️  Durée: {:.2}s", metadata.duration_seconds);
    println!("🔊 Bitrate: {:?} kbps", metadata.bitrate_kbps);
    
    // Générer waveform
    let waveform = processor.generate_waveform(Path::new("audio/song.mp3"), 2000).await?;
    println!("📊 Points de waveform: {}", waveform.peaks.len());
    
    Ok(())
}
```

### Compression audio

```rust
use stream_server::audio::compression::{CompressionEngine, CompressionRequest};

async fn example_audio_compression() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let engine = CompressionEngine::new(config);
    
    // Démarrer les workers
    engine.start_workers().await;
    
    // Compression asynchrone
    let request = CompressionRequest {
        input_file: "original.wav".to_string(),
        target_quality: "high".to_string(),
        preserve_metadata: true,
        async_processing: true,
    };
    
    let response = engine.compress_audio(request).await?;
    println!("🆔 Job ID: {}", response.job_id);
    
    // Suivre la progression
    loop {
        if let Some(job) = engine.get_job_status(&response.job_id).await {
            println!("📊 Progression: {:.1}%", job.progress);
            
            match job.status {
                JobStatus::Completed => {
                    println!("✅ Compression terminée!");
                    println!("📉 Ratio: {:.2}", job.compression_ratio.unwrap_or(0.0));
                    break;
                }
                JobStatus::Failed => {
                    println!("❌ Échec: {:?}", job.error_message);
                    break;
                }
                _ => tokio::time::sleep(Duration::from_secs(1)).await,
            }
        }
    }
    
    Ok(())
}
```

### Profil de compression personnalisé

```rust
use stream_server::audio::compression::{CompressionProfile, AudioCodec};

async fn example_custom_profile() -> Result<(), Box<dyn std::error::Error>> {
    let config = Arc::new(Config::from_env()?);
    let mut engine = CompressionEngine::new(config);
    
    // Créer un profil personnalisé pour podcast
    let podcast_profile = CompressionProfile {
        name: "Podcast Ultra".to_string(),
        codec: AudioCodec::OPUS,
        bitrate_kbps: 48,
        sample_rate: 16000,
        channels: 1, // Mono
        quality_factor: 0.9,
        compression_level: 7,
        target_size_reduction: 0.95, // 95% de réduction
    };
    
    engine.add_custom_profile("podcast_ultra".to_string(), podcast_profile).await;
    
    // Utiliser le profil personnalisé
    let request = CompressionRequest {
        input_file: "interview.wav".to_string(),
        target_quality: "podcast_ultra".to_string(),
        preserve_metadata: true,
        async_processing: false,
    };
    
    let response = engine.compress_audio(request).await?;
    println!("🎙️ Podcast compressé: {:?}", response.download_url);
    
    Ok(())
}
```

## Intégration

### Avec le serveur principal

```rust
// Dans main.rs
let audio_processor = Arc::new(AudioProcessor::new(config.clone()));
let compression_engine = Arc::new(CompressionEngine::new(config.clone()));

// Démarrer les workers de compression
compression_engine.start_workers().await;

// Inclure dans l'état de l'application
let app_state = AppState {
    audio_processor,
    compression_engine,
    // ... autres composants
};
```

### Avec l'API REST

```rust
// Endpoint pour métadonnées
async fn get_audio_metadata(
    Path(filename): Path<String>,
    State(state): State<AppState>,
) -> Result<Json<AudioMetadata>, StatusCode> {
    let path = PathBuf::from(&state.config.audio_dir).join(&filename);
    
    match state.audio_processor.extract_metadata(&path).await {
        Ok(metadata) => Ok(Json(metadata)),
        Err(_) => Err(StatusCode::NOT_FOUND),
    }
}

// Endpoint pour compression
async fn compress_audio(
    State(state): State<AppState>,
    Json(request): Json<CompressionRequest>,
) -> Result<Json<CompressionResponse>, StatusCode> {
    match state.compression_engine.compress_audio(request).await {
        Ok(response) => Ok(Json(response)),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}
```

### Avec le frontend React

```typescript
// Service de métadonnées
export class AudioMetadataService {
  async getMetadata(filename: string): Promise<AudioMetadata> {
    const response = await fetch(`/api/audio/${filename}/metadata`);
    return response.json();
  }
  
  async getWaveform(filename: string, resolution = 2000): Promise<WaveformData> {
    const response = await fetch(`/api/audio/${filename}/waveform?resolution=${resolution}`);
    return response.json();
  }
}

// Service de compression
export class CompressionService {
  async compressAudio(request: CompressionRequest): Promise<CompressionResponse> {
    const response = await fetch('/api/compression/compress', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    return response.json();
  }
  
  async getJobStatus(jobId: string): Promise<CompressionJob> {
    const response = await fetch(`/api/compression/jobs/${jobId}`);
    return response.json();
  }
}
```

### Avec l'API Go

```go
// Structure correspondante en Go
type AudioMetadata struct {
    DurationSeconds    float64   `json:"duration_seconds"`
    SampleRate        uint32    `json:"sample_rate"`
    Channels          uint32    `json:"channels"`
    BitrateKbps       *uint32   `json:"bitrate_kbps,omitempty"`
    Codec             string    `json:"codec"`
    Title             *string   `json:"title,omitempty"`
    Artist            *string   `json:"artist,omitempty"`
    Album             *string   `json:"album,omitempty"`
    Year              *uint32   `json:"year,omitempty"`
    Genre             *string   `json:"genre,omitempty"`
    ArtworkAvailable  bool      `json:"artwork_available"`
    FileSize          uint64    `json:"file_size"`
    LastModified      time.Time `json:"last_modified"`
}

// Client pour l'API Rust
type AudioServiceClient struct {
    baseURL string
    client  *http.Client
}

func (c *AudioServiceClient) GetMetadata(filename string) (*AudioMetadata, error) {
    resp, err := c.client.Get(fmt.Sprintf("%s/api/audio/%s/metadata", c.baseURL, filename))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var metadata AudioMetadata
    if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
        return nil, err
    }
    
    return &metadata, nil
}
```

## Configuration

### Variables d'environnement

```bash
# Configuration audio
AUDIO_DIR=./audio
MAX_FILE_SIZE=104857600  # 100MB

# Configuration de compression
COMPRESSION_ENABLED=true
COMPRESSION_OUTPUT_DIR=./compressed
COMPRESSION_TEMP_DIR=./temp
COMPRESSION_MAX_CONCURRENT_JOBS=4
COMPRESSION_CLEANUP_AFTER_DAYS=30
FFMPEG_PATH=/usr/bin/ffmpeg  # Optionnel
```

### Dans config.toml

```toml
[compression]
enabled = true
output_dir = "./compressed"
temp_dir = "./temp"
max_concurrent_jobs = 4
cleanup_after_days = 30
ffmpeg_path = "/usr/bin/ffmpeg"
quality_profiles = ["ultra_high", "high", "medium", "low", "mobile", "podcast"]
```

Cette documentation complète du module audio vous permet d'intégrer facilement les fonctionnalités de traitement et compression audio dans votre architecture Rust/Go/React. 