fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Générer les bindings Rust à partir des fichiers .proto
    let proto_dir = "proto";
    let proto_files = vec![
        "proto/stream/stream.proto",
        "proto/common/auth.proto",
    ];

    // Configuration tonic-build
    tonic_build::configure()
        .build_server(true)
        .build_client(false) // Stream server est serveur, pas client
        .out_dir("src/generated")
        .compile(&proto_files, &[proto_dir])?;

    // Recompiler si les fichiers .proto changent
    for proto_file in &proto_files {
        println!("cargo:rerun-if-changed={}", proto_file);
    }
    
    println!("cargo:rerun-if-changed=build.rs");
    
    Ok(())
} 