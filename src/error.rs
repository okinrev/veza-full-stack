    // Erreurs de validation et parsing
    ValidationError(String),
    ParseError(String),
    ParameterMismatch { expected: String, got: String },
    InvalidRange,
    
    // Erreurs de synchronisation
    TimeSync,
    NoSyncPoint,
    
    // Erreurs de playback 
    FileError { message: String },
}

// Conversions depuis les erreurs standard
impl From<std::io::Error> for AppError {
    fn from(err: std::io::Error) -> Self {
        AppError::FileError { message: err.to_string() }
    }
} 