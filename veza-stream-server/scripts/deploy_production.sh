#!/bin/bash

# === SCRIPT DE DÉPLOIEMENT PRODUCTION ===
# Déploie les modules Rust Veza en production
# Usage: ./scripts/deploy_production.sh [version]

set -euo pipefail

# Configuration
PROJECT_NAME="veza-stream-server"
NAMESPACE="veza-production"
VERSION="${1:-latest}"
REGISTRY="ghcr.io/veza"

# Couleurs pour les logs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonctions utilitaires
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Vérifications pré-déploiement
check_prerequisites() {
    log_info "Vérification des prérequis..."
    
    # Vérifier kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl non trouvé. Veuillez l'installer."
        exit 1
    fi
    
    # Vérifier la connexion au cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Impossible de se connecter au cluster Kubernetes"
        exit 1
    fi
    
    # Vérifier le namespace
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Namespace $NAMESPACE non trouvé, création..."
        kubectl create namespace "$NAMESPACE"
    fi
    
    log_success "Prérequis vérifiés"
}

# Déploiement des secrets
deploy_secrets() {
    log_info "Déploiement des secrets..."
    
    if kubectl get secret postgres-secret -n "$NAMESPACE" &> /dev/null; then
        log_warning "Secret postgres-secret existe déjà"
    else
        kubectl apply -f k8s/production/secrets.yaml
    fi
    
    log_success "Secrets déployés"
}

# Déploiement des ConfigMaps
deploy_config() {
    log_info "Déploiement des ConfigMaps..."
    kubectl apply -f k8s/production/configmap.yaml
    log_success "ConfigMaps déployées"
}

# Déploiement de l'application
deploy_application() {
    log_info "Déploiement de l'application version $VERSION..."
    
    # Mettre à jour l'image dans le manifest
    sed "s|image: veza/stream-server:.*|image: $REGISTRY/stream-server:$VERSION|g" \
        k8s/production/stream-server-deployment.yaml | kubectl apply -f -
    
    # Attendre le déploiement
    log_info "Attente du déploiement..."
    kubectl rollout status deployment/"$PROJECT_NAME" -n "$NAMESPACE" --timeout=600s
    
    log_success "Application déployée"
}

# Vérifications post-déploiement
verify_deployment() {
    log_info "Vérification du déploiement..."
    
    # Vérifier les pods
    local pods_ready
    pods_ready=$(kubectl get pods -n "$NAMESPACE" -l app="$PROJECT_NAME" --no-headers | awk '{print $2}' | grep -c "1/1" || true)
    local total_pods
    total_pods=$(kubectl get pods -n "$NAMESPACE" -l app="$PROJECT_NAME" --no-headers | wc -l)
    
    log_info "Pods prêts: $pods_ready/$total_pods"
    
    if [ "$pods_ready" -eq "$total_pods" ] && [ "$total_pods" -gt 0 ]; then
        log_success "Tous les pods sont prêts"
    else
        log_error "Certains pods ne sont pas prêts"
        kubectl get pods -n "$NAMESPACE" -l app="$PROJECT_NAME"
        exit 1
    fi
    
    # Vérifier les services
    if kubectl get service "$PROJECT_NAME" -n "$NAMESPACE" &> /dev/null; then
        log_success "Service accessible"
    else
        log_error "Service non accessible"
        exit 1
    fi
    
    # Test de santé
    log_info "Test de santé de l'application..."
    local service_ip
    service_ip=$(kubectl get service "$PROJECT_NAME" -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}')
    
    if kubectl run test-pod --rm -i --restart=Never --image=curlimages/curl -- \
        curl -f "http://$service_ip:8080/health" &> /dev/null; then
        log_success "Application répond correctement"
    else
        log_error "Application ne répond pas"
        exit 1
    fi
}

# Fonction de rollback
rollback() {
    log_warning "Rollback en cours..."
    kubectl rollout undo deployment/"$PROJECT_NAME" -n "$NAMESPACE"
    kubectl rollout status deployment/"$PROJECT_NAME" -n "$NAMESPACE"
    log_success "Rollback terminé"
}

# Fonction de nettoyage
cleanup() {
    if [ $? -ne 0 ]; then
        log_error "Erreur détectée, rollback automatique..."
        rollback
    fi
}

# Main
main() {
    echo "🚀 DÉPLOIEMENT PRODUCTION VEZA STREAM SERVER"
    echo "============================================="
    echo "Version: $VERSION"
    echo "Namespace: $NAMESPACE"
    echo "Registry: $REGISTRY"
    echo ""
    
    # Trap pour le nettoyage en cas d'erreur
    trap cleanup ERR
    
    check_prerequisites
    deploy_secrets
    deploy_config
    deploy_application
    verify_deployment
    
    echo ""
    log_success "🎉 DÉPLOIEMENT RÉUSSI !"
    echo ""
    echo "Informations de déploiement:"
    kubectl get pods -n "$NAMESPACE" -l app="$PROJECT_NAME"
    echo ""
    kubectl get services -n "$NAMESPACE" -l app="$PROJECT_NAME"
}

# Exécution
main "$@"
