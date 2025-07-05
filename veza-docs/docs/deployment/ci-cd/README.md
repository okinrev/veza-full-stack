---
id: ci-cd-readme
title: CI/CD Pipeline
sidebar_label: CI/CD
---

# Pipeline CI/CD - Veza

## Vue d'ensemble

Ce document décrit le pipeline CI/CD de la plateforme Veza.

## Étapes du Pipeline

### 1. Build
```yaml
- name: Build
  run: |
    go build -o server cmd/server/main.go
    cargo build --release
```

### 2. Tests
```yaml
- name: Tests
  run: |
    go test ./...
    cargo test
```

### 3. Déploiement
```yaml
- name: Deploy
  run: |
    docker build -t veza:latest .
    kubectl apply -f k8s/
```

## Environnements

- **Development** : Tests automatiques
- **Staging** : Tests d'intégration
- **Production** : Déploiement manuel

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 