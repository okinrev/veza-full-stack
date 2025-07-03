/// Utilitaire pour générer des URLs signées
/// 
/// Exemple d'utilisation:
/// ```bash
/// cargo run --bin generate_url -- --filename track1.mp3 --duration 3600
/// ```

use chrono::Utc;
use hmac::{Hmac, Mac};
use sha2::Sha256;
use std::env;
use clap::Parser;


#[derive(Parser)]
#[command(name = "generate_url")]
#[command(about = "Génère une URL signée pour le serveur de streaming")]
struct Args {
    /// Nom du fichier (avec extension)
    #[arg(short, long)]
    filename: String,
    
    /// Durée de validité en secondes
    #[arg(short, long, default_value = "3600")]
    duration: i64,
    
    /// URL de base du serveur
    #[arg(short, long, default_value = "http://localhost:8082")]
    base_url: String,
}



#[allow(dead_code)]
fn generate_signature(filename: &str, expires: i64, secret: &str) -> String {
    let to_sign = format!("{}|{}", filename, expires);
    let mut mac = Hmac::<Sha256>::new_from_slice(secret.as_bytes())
        .expect("HMAC can take key of any size");
    mac.update(to_sign.as_bytes());
    let result = mac.finalize();
    hex::encode(result.into_bytes())
}

#[allow(dead_code)]
fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();

    let secret_key = env::var("SECRET_KEY")
        .map_err(|_| "SECRET_KEY environment variable must be set")?;

    if secret_key.len() < 32 {
        return Err("SECRET_KEY must be at least 32 characters long".into());
    }

    let expires = Utc::now().timestamp() + args.duration;
    let signature = generate_signature(&args.filename, expires, &secret_key);

    let url = format!(
        "{}/stream/{}?expires={}&sig={}",
        args.base_url.trim_end_matches('/'),
        args.filename,
        expires,
        signature
    );

    println!("URL signée générée:");
    println!("{}", url);
    println!();
    println!("Valide jusqu'au: {}", chrono::DateTime::<chrono::Utc>::from_timestamp(expires, 0).unwrap().format("%Y-%m-%d %H:%M:%S UTC"));
    println!("Durée: {} secondes", args.duration);

    // Test avec curl
    println!();
    println!("Test avec curl:");
    println!("curl -v \"{}\"", url);

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_signature_generation() {
        let secret = "test_secret_key_that_is_long_enough_32chars";
        let filename = "test.mp3";
        let expires = 1609459200; // 2021-01-01 00:00:00 UTC
        
        let sig1 = generate_signature(filename, expires, secret);
        let sig2 = generate_signature(filename, expires, secret);
        
        // Les signatures doivent être identiques
        assert_eq!(sig1, sig2);
        
        // La signature doit être déterministe
        assert_eq!(sig1.len(), 64); // 32 bytes en hex = 64 caractères
    }

    #[test]
    fn test_different_inputs_different_signatures() {
        let secret = "test_secret_key_that_is_long_enough_32chars";
        
        let sig1 = generate_signature("file1.mp3", 1609459200, secret);
        let sig2 = generate_signature("file2.mp3", 1609459200, secret);
        let sig3 = generate_signature("file1.mp3", 1609459201, secret);
        
        // Tous doivent être différents
        assert_ne!(sig1, sig2);
        assert_ne!(sig1, sig3);
        assert_ne!(sig2, sig3);
    }
} 