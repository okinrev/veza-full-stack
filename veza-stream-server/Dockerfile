# Dockerfile multi-stage pour optimiser la taille finale
FROM rust:1.75-slim AS builder

# Installation des dépendances système nécessaires
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Création du répertoire de travail
WORKDIR /app

# Copie des fichiers de configuration Cargo
COPY Cargo.toml Cargo.lock ./

# Pré-compilation des dépendances (optimisation Docker layer caching)
RUN mkdir src && \
    echo "fn main() {}" > src/main.rs && \
    cargo build --release && \
    rm -rf src

# Copie du code source
COPY src ./src
COPY audio ./audio

# Variables de build
ARG BUILD_TIME
ARG RUST_VERSION
ENV BUILD_TIME=${BUILD_TIME}
ENV RUST_VERSION=${RUST_VERSION}

# Compilation finale avec optimisations
RUN cargo build --release --locked

# ===================================================================
# Stage final - Image minimale pour la production
# ===================================================================
FROM debian:bookworm-slim

# Installation des dépendances runtime minimales
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Création d'un utilisateur non-root pour la sécurité
RUN groupadd -r streamserver && \
    useradd -r -g streamserver -d /app -s /sbin/nologin streamserver

# Création des répertoires nécessaires
WORKDIR /app
RUN mkdir -p audio logs && \
    chown -R streamserver:streamserver /app

# Copie des binaires depuis l'étape de build
COPY --from=builder /app/target/release/stream_server /usr/local/bin/
COPY --from=builder /app/target/release/generate_url /usr/local/bin/

# Copie des fichiers de configuration
COPY --from=builder /app/audio ./audio/
COPY env.example .env.example

# Ajout d'un script de santé pour Docker
RUN echo '#!/bin/bash\ncurl -f http://localhost:$STREAM_SERVER_PORT/health || exit 1' > /usr/local/bin/healthcheck.sh && \
    chmod +x /usr/local/bin/healthcheck.sh

# Configuration des permissions finales
RUN chown -R streamserver:streamserver /app && \
    chmod +x /usr/local/bin/stream_server /usr/local/bin/generate_url

# Basculement vers l'utilisateur non-root
USER streamserver

# Variables d'environnement par défaut
ENV STREAM_SERVER_PORT=8082
ENV AUDIO_DIR=/app/audio
ENV RUST_LOG=stream_server=info
ENV ALLOWED_ORIGINS=*
ENV MAX_FILE_SIZE=104857600
ENV MAX_RANGE_SIZE=10485760
ENV SIGNATURE_TOLERANCE=60

# Exposition du port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

# Labels pour la documentation
LABEL org.opencontainers.image.title="Stream Server" \
      org.opencontainers.image.description="Serveur de streaming audio sécurisé en Rust" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.authors="Votre nom <email@example.com>" \
      org.opencontainers.image.source="https://github.com/votre-user/stream-server" \
      org.opencontainers.image.licenses="MIT"

# Point d'entrée
ENTRYPOINT ["/usr/local/bin/stream_server"] 