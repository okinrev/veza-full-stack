use clap::Parser;
use std::{
    fs,
    path::{Path, PathBuf},
    time::Instant,
};
use tokio::time::{sleep, Duration};
use tracing::{info, error, warn};

#[derive(Parser)]
#[command(name = "waveform_generator")]
#[command(about = "Génère les waveforms pour tous les fichiers audio")]
struct Args {
    /// Dossier contenant les fichiers audio
    #[arg(short, long, default_value = "audio")]
    input_dir: String,
    
    /// Dossier de sortie pour les waveforms
    #[arg(short, long, default_value = "waveforms")]
    output_dir: String,
    
    /// Résolution de la waveform (nombre de points)
    #[arg(short, long, default_value = "1000")]
    resolution: usize,
    
    /// Extensions de fichier à traiter
    #[arg(short, long, default_values_t = vec!["mp3".to_string(), "wav".to_string(), "flac".to_string()])]
    extensions: Vec<String>,
    
    /// Forcer la régénération même si le fichier existe
    #[arg(short, long)]
    force: bool,
    
    /// Traitement en parallèle (nombre de workers)
    #[arg(short, long, default_value = "4")]
    workers: usize,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Configuration du logging
    tracing_subscriber::fmt()
        .with_env_filter("waveform_generator=info")
        .init();

    let args = Args::parse();

    info!("🎵 Générateur de waveforms démarré");
    info!("📁 Dossier d'entrée: {}", args.input_dir);
    info!("📁 Dossier de sortie: {}", args.output_dir);
    info!("📊 Résolution: {} points", args.resolution);
    info!("⚡ Workers: {}", args.workers);

    // Créer le dossier de sortie
    fs::create_dir_all(&args.output_dir)?;

    // Trouver tous les fichiers audio
    let audio_files = find_audio_files(&args.input_dir, &args.extensions)?;
    info!("🔍 {} fichiers audio trouvés", audio_files.len());

    if audio_files.is_empty() {
        warn!("Aucun fichier audio trouvé dans {}", args.input_dir);
        return Ok(());
    }

    let start_time = Instant::now();
    let mut processed = 0;
    let mut errors = 0;
    let mut skipped = 0;

    // Traitement en parallèle avec limite de workers
    let semaphore = tokio::sync::Semaphore::new(args.workers);
    let mut handles = Vec::new();

    for audio_file in audio_files {
        let permit = semaphore.clone().acquire_owned().await?;
        let output_dir = args.output_dir.clone();
        let resolution = args.resolution;
        let force = args.force;

        let handle = tokio::spawn(async move {
            let _permit = permit;
            process_audio_file(&audio_file, &output_dir, resolution, force).await
        });

        handles.push(handle);
    }

    // Attendre tous les traitements
    for handle in handles {
        match handle.await? {
            ProcessResult::Processed => processed += 1,
            ProcessResult::Skipped => skipped += 1,
            ProcessResult::Error => errors += 1,
        }
    }

    let elapsed = start_time.elapsed();
    
    info!("✅ Traitement terminé en {:.2}s", elapsed.as_secs_f64());
    info!("📊 Résultats:");
    info!("  - Traités: {}", processed);
    info!("  - Ignorés: {}", skipped);
    info!("  - Erreurs: {}", errors);

    if errors > 0 {
        warn!("⚠️  {} fichiers n'ont pas pu être traités", errors);
    }

    Ok(())
}

#[derive(Debug)]
enum ProcessResult {
    Processed,
    Skipped,
    Error,
}

async fn process_audio_file(
    audio_path: &Path,
    output_dir: &str,
    resolution: usize,
    force: bool,
) -> ProcessResult {
    let file_stem = audio_path.file_stem().unwrap().to_str().unwrap();
    let waveform_path = PathBuf::from(output_dir).join(format!("{}.json", file_stem));

    // Vérifier si le fichier existe déjà
    if !force && waveform_path.exists() {
        // Vérifier si le fichier waveform est plus récent que l'audio
        if let (Ok(audio_meta), Ok(waveform_meta)) = (audio_path.metadata(), waveform_path.metadata()) {
            if let (Ok(audio_modified), Ok(waveform_modified)) = (audio_meta.modified(), waveform_meta.modified()) {
                if waveform_modified >= audio_modified {
                    return ProcessResult::Skipped;
                }
            }
        }
    }

    info!("🔄 Traitement: {}", audio_path.display());

    // Simuler le traitement avec l'audio processor
    // En production, on utiliserait le module audio_processing
    match generate_waveform_data(audio_path, resolution).await {
        Ok(waveform_data) => {
            // Sauvegarder la waveform en JSON
            if let Err(e) = save_waveform_json(&waveform_path, &waveform_data).await {
                error!("Erreur sauvegarde waveform pour {}: {}", audio_path.display(), e);
                return ProcessResult::Error;
            }

            info!("✅ Waveform générée: {}", waveform_path.display());
            ProcessResult::Processed
        }
        Err(e) => {
            error!("❌ Erreur génération waveform pour {}: {}", audio_path.display(), e);
            ProcessResult::Error
        }
    }
}

async fn generate_waveform_data(
    _audio_path: &Path,
    resolution: usize,
) -> Result<WaveformData, Box<dyn std::error::Error + Send + Sync>> {
    // Simulation du traitement audio
    // En production, on utiliserait symphonia pour décoder l'audio
    sleep(Duration::from_millis(100)).await; // Simule le temps de traitement

    // Générer des données de test
    let peaks: Vec<f32> = (0..resolution)
        .map(|i| {
            let t = i as f32 / resolution as f32;
            (t * std::f32::consts::PI * 4.0).sin().abs() * (1.0 - t * 0.5)
        })
        .collect();

    let rms: Vec<f32> = peaks.iter().map(|&p| p * 0.7).collect();

    Ok(WaveformData {
        peaks,
        rms,
        sample_rate: 100, // 100 points par seconde
        duration_ms: (resolution * 10) as u32, // 10ms par point
        generated_at: std::time::SystemTime::now(),
    })
}

async fn save_waveform_json(
    path: &Path,
    data: &WaveformData,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let json = serde_json::to_string_pretty(data)?;
    tokio::fs::write(path, json).await?;
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

// Structures de données pour la waveform (copiées du module audio_processing)
#[derive(serde::Serialize, serde::Deserialize)]
struct WaveformData {
    peaks: Vec<f32>,
    rms: Vec<f32>,
    sample_rate: u32,
    duration_ms: u32,
    generated_at: std::time::SystemTime,
} 