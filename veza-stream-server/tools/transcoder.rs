use clap::Parser;
use std::{
    fs,
    path::{Path, PathBuf},
    process::Command,
    time::Instant,
};
use serde::{Serialize, Deserialize};
use tracing::{info, error, warn, debug};

#[derive(Parser)]
#[command(name = "transcoder")]
#[command(about = "Transcode les fichiers audio en plusieurs qualités")]
struct Args {
    /// Dossier contenant les fichiers audio source
    #[arg(short, long, default_value = "audio")]
    input_dir: String,
    
    /// Dossier de sortie pour les fichiers transcodés
    #[arg(short, long, default_value = "transcoded")]
    output_dir: String,
    
    /// Extensions de fichier à traiter
    #[arg(short, long, default_values_t = vec!["mp3".to_string(), "wav".to_string(), "flac".to_string()])]
    extensions: Vec<String>,
    
    /// Qualités à générer (séparées par des virgules)
    #[arg(short, long, default_value = "high,medium,low,mobile")]
    qualities: String,
    
    /// Forcer la régénération même si le fichier existe
    #[arg(short, long)]
    force: bool,
    
    /// Traitement en parallèle (nombre de workers)
    #[arg(short, long, default_value = "2")]
    workers: usize,
    
    /// Chemin vers ffmpeg (optionnel)
    #[arg(long)]
    ffmpeg_path: Option<String>,
    
    /// Préserver les métadonnées
    #[arg(long, default_value = "true")]
    preserve_metadata: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct QualityProfile {
    name: String,
    bitrate_kbps: u32,
    sample_rate: u32,
    channels: u8,
    codec: String,
    extension: String,
}

impl QualityProfile {
    fn high() -> Self {
        Self {
            name: "high".to_string(),
            bitrate_kbps: 320,
            sample_rate: 44100,
            channels: 2,
            codec: "libmp3lame".to_string(),
            extension: "mp3".to_string(),
        }
    }

    fn medium() -> Self {
        Self {
            name: "medium".to_string(),
            bitrate_kbps: 192,
            sample_rate: 44100,
            channels: 2,
            codec: "libmp3lame".to_string(),
            extension: "mp3".to_string(),
        }
    }

    fn low() -> Self {
        Self {
            name: "low".to_string(),
            bitrate_kbps: 128,
            sample_rate: 22050,
            channels: 2,
            codec: "libmp3lame".to_string(),
            extension: "mp3".to_string(),
        }
    }

    fn mobile() -> Self {
        Self {
            name: "mobile".to_string(),
            bitrate_kbps: 96,
            sample_rate: 22050,
            channels: 1,
            codec: "libmp3lame".to_string(),
            extension: "mp3".to_string(),
        }
    }

    fn get_ffmpeg_args(&self) -> Vec<String> {
        vec![
            "-acodec".to_string(),
            self.codec.clone(),
            "-ab".to_string(),
            format!("{}k", self.bitrate_kbps),
            "-ar".to_string(),
            self.sample_rate.to_string(),
            "-ac".to_string(),
            self.channels.to_string(),
        ]
    }
}

#[derive(Debug)]
enum TranscodeResult {
    Success,
    Skipped,
    Error(String),
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Configuration du logging
    tracing_subscriber::fmt()
        .with_env_filter("transcoder=info")
        .init();

    let args = Args::parse();

    info!("🎵 Transcodeur audio démarré");
    info!("📁 Dossier d'entrée: {}", args.input_dir);
    info!("📁 Dossier de sortie: {}", args.output_dir);
    info!("⚡ Workers: {}", args.workers);

    // Vérifier que ffmpeg est disponible
    let ffmpeg_cmd = args.ffmpeg_path.as_deref().unwrap_or("ffmpeg");
    if !check_ffmpeg_available(ffmpeg_cmd) {
        error!("❌ FFmpeg non trouvé. Veuillez l'installer ou spécifier le chemin avec --ffmpeg-path");
        return Err("FFmpeg requis".into());
    }

    info!("✅ FFmpeg disponible");

    // Parser les qualités demandées
    let quality_profiles = parse_quality_profiles(&args.qualities)?;
    info!("🎯 Qualités à générer: {:?}", quality_profiles.iter().map(|q| &q.name).collect::<Vec<_>>());

    // Créer la structure de dossiers
    create_output_structure(&args.output_dir, &quality_profiles)?;

    // Trouver tous les fichiers audio
    let audio_files = find_audio_files(&args.input_dir, &args.extensions)?;
    info!("🔍 {} fichiers audio trouvés", audio_files.len());

    if audio_files.is_empty() {
        warn!("Aucun fichier audio trouvé dans {}", args.input_dir);
        return Ok(());
    }

    let start_time = Instant::now();
    let mut stats = TranscodeStats::default();

    // Traitement en parallèle avec limite de workers
    let semaphore = tokio::sync::Semaphore::new(args.workers);
    let mut handles = Vec::new();

    for audio_file in audio_files {
        for quality in &quality_profiles {
            let permit = Arc::clone(&semaphore).acquire_owned().await?;
            let audio_file = audio_file.clone();
            let quality = quality.clone();
            let output_dir = args.output_dir.clone();
            let force = args.force;
            let preserve_metadata = args.preserve_metadata;
            let ffmpeg_cmd = ffmpeg_cmd.to_string();

            let handle = tokio::spawn(async move {
                let _permit = permit;
                transcode_file(&audio_file, &quality, &output_dir, &ffmpeg_cmd, force, preserve_metadata).await
            });

            handles.push(handle);
        }
    }

    // Attendre tous les traitements et collecter les résultats
    for handle in handles {
        match handle.await? {
            TranscodeResult::Success => stats.success += 1,
            TranscodeResult::Skipped => stats.skipped += 1,
            TranscodeResult::Error(_) => stats.errors += 1,
        }
    }

    let elapsed = start_time.elapsed();
    
    info!("✅ Transcodage terminé en {:.2}s", elapsed.as_secs_f64());
    info!("📊 Résultats:");
    info!("  - Succès: {}", stats.success);
    info!("  - Ignorés: {}", stats.skipped);
    info!("  - Erreurs: {}", stats.errors);
    info!("  - Total: {}", stats.total());

    if stats.errors > 0 {
        warn!("⚠️  {} transcodages ont échoué", stats.errors);
    }

    Ok(())
}

#[derive(Default)]
struct TranscodeStats {
    success: u32,
    skipped: u32,
    errors: u32,
}

impl TranscodeStats {
    fn total(&self) -> u32 {
        self.success + self.skipped + self.errors
    }
}

async fn transcode_file(
    input_path: &Path,
    quality: &QualityProfile,
    output_dir: &str,
    ffmpeg_cmd: &str,
    force: bool,
    preserve_metadata: bool,
) -> TranscodeResult {
    let file_stem = input_path.file_stem().unwrap().to_str().unwrap();
    let output_path = PathBuf::from(output_dir)
        .join(&quality.name)
        .join(format!("{}.{}", file_stem, quality.extension));

    // Vérifier si le fichier existe déjà
    if !force && output_path.exists() {
        if let (Ok(input_meta), Ok(output_meta)) = (input_path.metadata(), output_path.metadata()) {
            if let (Ok(input_modified), Ok(output_modified)) = (input_meta.modified(), output_meta.modified()) {
                if output_modified >= input_modified {
                    debug!("⏭️  Ignoré (déjà à jour): {} -> {}", input_path.display(), quality.name);
                    return TranscodeResult::Skipped;
                }
            }
        }
    }

    info!("🔄 Transcodage: {} -> {}", input_path.display(), quality.name);

    // Construire la commande ffmpeg
    let mut cmd = Command::new(ffmpeg_cmd);
    cmd.arg("-i").arg(input_path);

    // Arguments de qualité
    for arg in quality.get_ffmpeg_args() {
        cmd.arg(arg);
    }

    // Préserver les métadonnées si demandé
    if preserve_metadata {
        cmd.arg("-map_metadata").arg("0");
    }

    // Options de performance
    cmd.arg("-threads").arg("0"); // Utiliser tous les cores disponibles
    cmd.arg("-y"); // Overwrite output files

    // Fichier de sortie
    cmd.arg(&output_path);

    // Exécuter le transcodage
    match cmd.output() {
        Ok(output) => {
            if output.status.success() {
                info!("✅ Transcodé: {} ({} kbps)", file_stem, quality.bitrate_kbps);
                TranscodeResult::Success
            } else {
                let error_msg = String::from_utf8_lossy(&output.stderr);
                error!("❌ Erreur transcodage {}: {}", file_stem, error_msg);
                TranscodeResult::Error(error_msg.to_string())
            }
        }
        Err(e) => {
            error!("❌ Erreur exécution ffmpeg pour {}: {}", file_stem, e);
            TranscodeResult::Error(e.to_string())
        }
    }
}

fn check_ffmpeg_available(ffmpeg_cmd: &str) -> bool {
    Command::new(ffmpeg_cmd)
        .arg("-version")
        .output()
        .map(|output| output.status.success())
        .unwrap_or(false)
}

fn parse_quality_profiles(qualities_str: &str) -> Result<Vec<QualityProfile>, Box<dyn std::error::Error>> {
    let mut profiles = Vec::new();
    
    for quality_name in qualities_str.split(',') {
        let quality_name = quality_name.trim();
        let profile = match quality_name {
            "high" => QualityProfile::high(),
            "medium" => QualityProfile::medium(),
            "low" => QualityProfile::low(),
            "mobile" => QualityProfile::mobile(),
            _ => return Err(format!("Qualité inconnue: {}", quality_name).into()),
        };
        profiles.push(profile);
    }

    if profiles.is_empty() {
        return Err("Aucune qualité spécifiée".into());
    }

    Ok(profiles)
}

fn create_output_structure(
    output_dir: &str,
    quality_profiles: &[QualityProfile],
) -> Result<(), Box<dyn std::error::Error>> {
    fs::create_dir_all(output_dir)?;
    
    for profile in quality_profiles {
        let quality_dir = PathBuf::from(output_dir).join(&profile.name);
        fs::create_dir_all(quality_dir)?;
    }
    
    Ok(())
}

fn find_audio_files(
    dir: &str,
    extensions: &[String],
) -> Result<Vec<PathBuf>, Box<dyn std::error::Error>> {
    let mut files = Vec::new();
    let dir_path = Path::new(dir);

    if !dir_path.exists() {
        return Err(format!("Le dossier {} n'existe pas", dir).into());
    }

    for entry in fs::read_dir(dir_path)? {
        let entry = entry?;
        let path = entry.path();

        if path.is_file() {
            if let Some(ext) = path.extension() {
                if let Some(ext_str) = ext.to_str() {
                    if extensions.iter().any(|e| e.eq_ignore_ascii_case(ext_str)) {
                        files.push(path);
                    }
                }
            }
        }
    }

    files.sort();
    Ok(files)
} 