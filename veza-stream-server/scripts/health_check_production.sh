#!/bin/bash

# === SCRIPT DE SANTÉ PRODUCTION ===
# Vérifie l'état de santé complet des modules Rust en production
# Usage: ./scripts/health_check_production.sh

set -euo pipefail

# Configuration
NAMESPACE="veza-production"
SERVICE_NAME="veza-stream-server"
CHAT_SERVICE="veza-chat-server"

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Fonctions
log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# Variables de tracking
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

run_check() {
    local check_name="$1"
    local check_command="$2"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    log_info "Test: $check_name"
    
    if eval "$check_command" &> /dev/null; then
        log_success "$check_name - OK"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        log_error "$check_name - ÉCHEC"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

# Vérifications
check_cluster_connection() {
    run_check "Connexion cluster Kubernetes" "kubectl cluster-info"
}

check_namespace() {
    run_check "Namespace $NAMESPACE" "kubectl get namespace $NAMESPACE"
}

check_deployments() {
    log_info "=== Vérification des déploiements ==="
    
    # Stream Server
    local stream_ready
    stream_ready=$(kubectl get deployment $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    local stream_desired
    stream_desired=$(kubectl get deployment $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    
    if [ "$stream_ready" -eq "$stream_desired" ] && [ "$stream_desired" -gt 0 ]; then
        log_success "Stream Server: $stream_ready/$stream_desired pods prêts"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        log_error "Stream Server: $stream_ready/$stream_desired pods prêts"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    # Chat Server
    local chat_ready
    chat_ready=$(kubectl get deployment $CHAT_SERVICE -n $NAMESPACE -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    local chat_desired
    chat_desired=$(kubectl get deployment $CHAT_SERVICE -n $NAMESPACE -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    
    if [ "$chat_ready" -eq "$chat_desired" ] && [ "$chat_desired" -gt 0 ]; then
        log_success "Chat Server: $chat_ready/$chat_desired pods prêts"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        log_error "Chat Server: $chat_ready/$chat_desired pods prêts"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

check_services() {
    log_info "=== Vérification des services ==="
    
    run_check "Service Stream Server" "kubectl get service $SERVICE_NAME -n $NAMESPACE"
    run_check "Service Chat Server" "kubectl get service $CHAT_SERVICE -n $NAMESPACE"
}

check_health_endpoints() {
    log_info "=== Tests des endpoints de santé ==="
    
    # Stream Server health
    local stream_ip
    stream_ip=$(kubectl get service $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
    
    if [ -n "$stream_ip" ]; then
        if kubectl run health-test-stream --rm -i --restart=Never --image=curlimages/curl -- \
            curl -f -m 10 "http://$stream_ip:8080/health" &> /dev/null; then
            log_success "Stream Server health endpoint"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_error "Stream Server health endpoint"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        log_error "Stream Server IP non trouvée"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    # Chat Server health
    local chat_ip
    chat_ip=$(kubectl get service $CHAT_SERVICE -n $NAMESPACE -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
    
    if [ -n "$chat_ip" ]; then
        if kubectl run health-test-chat --rm -i --restart=Never --image=curlimages/curl -- \
            curl -f -m 10 "http://$chat_ip:8080/health" &> /dev/null; then
            log_success "Chat Server health endpoint"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_error "Chat Server health endpoint"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        log_error "Chat Server IP non trouvée"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

check_database_connectivity() {
    log_info "=== Vérification connectivité base de données ==="
    
    run_check "Service PostgreSQL" "kubectl get pods -n $NAMESPACE -l app=postgres"
    run_check "Service Redis" "kubectl get pods -n $NAMESPACE -l app=redis"
}

check_metrics() {
    log_info "=== Vérification des métriques ==="
    
    # Métriques Prometheus
    local stream_ip
    stream_ip=$(kubectl get service $SERVICE_NAME -n $NAMESPACE -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
    
    if [ -n "$stream_ip" ]; then
        if kubectl run metrics-test --rm -i --restart=Never --image=curlimages/curl -- \
            curl -f -m 10 "http://$stream_ip:9090/metrics" &> /dev/null; then
            log_success "Métriques Prometheus disponibles"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_error "Métriques Prometheus indisponibles"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        log_error "IP service non trouvée pour métriques"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

check_resource_usage() {
    log_info "=== Utilisation des ressources ==="
    
    echo "Pods Stream Server:"
    kubectl top pods -n $NAMESPACE -l app=$SERVICE_NAME 2>/dev/null || log_warning "Metrics server non disponible"
    
    echo "Pods Chat Server:"
    kubectl top pods -n $NAMESPACE -l app=$CHAT_SERVICE 2>/dev/null || log_warning "Metrics server non disponible"
    
    echo "Utilisation des nœuds:"
    kubectl top nodes 2>/dev/null || log_warning "Metrics server non disponible"
}

check_recent_logs() {
    log_info "=== Vérification des logs récents ==="
    
    echo "Logs Stream Server (dernières 10 lignes):"
    kubectl logs -n $NAMESPACE -l app=$SERVICE_NAME --tail=10 2>/dev/null || log_warning "Logs non disponibles"
    
    echo "Logs Chat Server (dernières 10 lignes):"
    kubectl logs -n $NAMESPACE -l app=$CHAT_SERVICE --tail=10 2>/dev/null || log_warning "Logs non disponibles"
}

# Rapport final
generate_report() {
    echo ""
    echo "========================================"
    echo "           RAPPORT DE SANTÉ"
    echo "========================================"
    echo "Date: $(date)"
    echo "Namespace: $NAMESPACE"
    echo ""
    echo "Résultats:"
    echo "  Total des tests: $TOTAL_CHECKS"
    echo "  Tests réussis: $PASSED_CHECKS"
    echo "  Tests échoués: $FAILED_CHECKS"
    echo ""
    
    local success_rate
    success_rate=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    
    if [ "$FAILED_CHECKS" -eq 0 ]; then
        log_success "🎉 SYSTÈME EN PARFAITE SANTÉ ($success_rate%)"
        echo "Status: HEALTHY"
        exit 0
    elif [ "$success_rate" -ge 80 ]; then
        log_warning "⚠️  SYSTÈME OPÉRATIONNEL AVEC ALERTES ($success_rate%)"
        echo "Status: WARNING"
        exit 1
    else
        log_error "🚨 SYSTÈME EN ÉTAT CRITIQUE ($success_rate%)"
        echo "Status: CRITICAL"
        exit 2
    fi
}

# Main
main() {
    echo "🩺 VÉRIFICATION SANTÉ PRODUCTION VEZA"
    echo "====================================="
    echo ""
    
    check_cluster_connection
    check_namespace
    check_deployments
    check_services
    check_health_endpoints
    check_database_connectivity
    check_metrics
    check_resource_usage
    check_recent_logs
    
    generate_report
}

# Exécution
main "$@"
