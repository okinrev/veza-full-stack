---
id: troubleshooting
title: Guide de Troubleshooting - Veza Platform
sidebar_label: Troubleshooting
---

# Guide de Troubleshooting - Veza Platform

> **Guide complet pour diagnostiquer et résoudre les problèmes de la plateforme Veza**

## Vue d'ensemble

Ce guide fournit des procédures systématiques pour diagnostiquer et résoudre les problèmes courants de la plateforme Veza.

## Diagnostic Système

### Vérification de l'État des Services

```bash
#!/bin/bash
# scripts/diagnostic.sh

echo "=== Diagnostic Veza Platform ==="
echo "Date: $(date)"
echo "Hostname: $(hostname)"
echo "Uptime: $(uptime)"
echo ""

# Vérification des services
echo "=== Services Status ==="
services=("veza-backend-api" "veza-chat-server" "veza-stream-server" "postgresql" "redis" "nats-server")

for service in "${services[@]}"; do
    status=$(systemctl is-active $service 2>/dev/null || echo "not-found")
    echo "$service: $status"
done
echo ""

# Vérification des ports
echo "=== Ports Status ==="
ports=("8080" "8081" "8082" "5432" "6379" "4222" "8222")

for port in "${ports[@]}"; do
    if netstat -tuln | grep ":$port " > /dev/null; then
        echo "Port $port: OPEN"
    else
        echo "Port $port: CLOSED"
    fi
done
echo ""

# Vérification de l'espace disque
echo "=== Disk Usage ==="
df -h | grep -E "(Filesystem|/dev/)"
echo ""

# Vérification de la mémoire
echo "=== Memory Usage ==="
free -h
echo ""

# Vérification des logs récents
echo "=== Recent Logs ==="
journalctl -u veza-backend-api --since "1 hour ago" | tail -10
echo ""
```

### Script de Diagnostic Automatique

```bash
#!/bin/bash
# scripts/auto_diagnostic.sh

LOG_FILE="/var/log/veza/diagnostic_$(date +%Y%m%d_%H%M%S).log"

# Fonction de logging
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a $LOG_FILE
}

# Vérification de la connectivité réseau
check_network() {
    log "=== Network Connectivity ==="
    
    # Test DNS
    if nslookup google.com > /dev/null 2>&1; then
        log "DNS: OK"
    else
        log "DNS: FAILED"
    fi
    
    # Test connectivité externe
    if curl -s --connect-timeout 5 https://httpbin.org/ip > /dev/null; then
        log "External connectivity: OK"
    else
        log "External connectivity: FAILED"
    fi
}

# Vérification de la base de données
check_database() {
    log "=== Database Check ==="
    
    if pg_isready -h localhost -p 5432; then
        log "PostgreSQL: OK"
        
        # Test de connexion
        if psql -h localhost -U veza_user -d veza_production -c "SELECT 1;" > /dev/null 2>&1; then
            log "Database connection: OK"
        else
            log "Database connection: FAILED"
        fi
        
        # Vérification des connexions actives
        connections=$(psql -h localhost -U veza_user -d veza_production -t -c "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null)
        log "Active connections: $connections"
    else
        log "PostgreSQL: FAILED"
    fi
}

# Vérification de Redis
check_redis() {
    log "=== Redis Check ==="
    
    if redis-cli ping > /dev/null 2>&1; then
        log "Redis: OK"
        
        # Vérification de la mémoire
        memory=$(redis-cli info memory | grep "used_memory_human" | cut -d: -f2)
        log "Redis memory usage: $memory"
        
        # Vérification des clés
        keys=$(redis-cli dbsize)
        log "Redis keys: $keys"
    else
        log "Redis: FAILED"
    fi
}

# Vérification de NATS
check_nats() {
    log "=== NATS Check ==="
    
    if curl -s http://localhost:8222/healthz > /dev/null; then
        log "NATS: OK"
        
        # Vérification des connexions
        connections=$(curl -s http://localhost:8222/connz | jq '.num_connections' 2>/dev/null || echo "unknown")
        log "NATS connections: $connections"
    else
        log "NATS: FAILED"
    fi
}

# Vérification des applications
check_applications() {
    log "=== Applications Check ==="
    
    # Backend API
    if curl -s http://localhost:8080/health > /dev/null; then
        log "Backend API: OK"
    else
        log "Backend API: FAILED"
    fi
    
    # Chat Server
    if curl -s http://localhost:8081/health > /dev/null; then
        log "Chat Server: OK"
    else
        log "Chat Server: FAILED"
    fi
    
    # Stream Server
    if curl -s http://localhost:8082/health > /dev/null; then
        log "Stream Server: OK"
    else
        log "Stream Server: FAILED"
    fi
}

# Exécution des vérifications
main() {
    log "Starting diagnostic..."
    
    check_network
    check_database
    check_redis
    check_nats
    check_applications
    
    log "Diagnostic completed. Log file: $LOG_FILE"
}

main
```

## Problèmes Courants

### Erreurs de Base de Données

#### Problème : Connexions PostgreSQL épuisées

**Symptômes :**
- Erreur `FATAL: remaining connection slots are reserved for non-replication superuser connections`
- Applications ne peuvent pas se connecter à la base de données

**Diagnostic :**
```bash
# Vérifier les connexions actives
psql -h localhost -U veza_user -d veza_production -c "
SELECT 
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active_connections,
    count(*) FILTER (WHERE state = 'idle') as idle_connections
FROM pg_stat_activity;"

# Vérifier les connexions par application
psql -h localhost -U veza_user -d veza_production -c "
SELECT 
    application_name,
    count(*) as connections
FROM pg_stat_activity 
GROUP BY application_name;"
```

**Solutions :**
```bash
# Augmenter le nombre de connexions max
sudo -u postgres psql -c "ALTER SYSTEM SET max_connections = 200;"
sudo -u postgres psql -c "SELECT pg_reload_conf();"

# Tuer les connexions inactives
psql -h localhost -U veza_user -d veza_production -c "
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' 
AND state_change < now() - interval '10 minutes';"
```

#### Problème : Performance lente des requêtes

**Diagnostic :**
```bash
# Vérifier les requêtes lentes
psql -h localhost -U veza_user -d veza_production -c "
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"

# Vérifier les index manquants
psql -h localhost -U veza_user -d veza_production -c "
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public' 
AND n_distinct > 0 
ORDER BY n_distinct DESC;"
```

**Solutions :**
```sql
-- Créer des index pour les requêtes fréquentes
CREATE INDEX CONCURRENTLY idx_messages_room_created 
ON messages(room_id, created_at);

CREATE INDEX CONCURRENTLY idx_users_email 
ON users(email);

-- Analyser les tables
ANALYZE messages;
ANALYZE users;
ANALYZE streams;
```

### Erreurs Redis

#### Problème : Mémoire Redis épuisée

**Symptômes :**
- Erreur `OOM command not allowed when used memory > 'maxmemory'`
- Performance dégradée

**Diagnostic :**
```bash
# Vérifier l'utilisation mémoire
redis-cli info memory

# Vérifier les clés par type
redis-cli --scan --pattern "*" | wc -l
redis-cli --scan --pattern "user:*" | wc -l
redis-cli --scan --pattern "session:*" | wc -l
```

**Solutions :**
```bash
# Augmenter la mémoire max
redis-cli config set maxmemory 2gb

# Nettoyer les clés expirées
redis-cli FLUSHDB

# Configurer l'éviction
redis-cli config set maxmemory-policy allkeys-lru
```

#### Problème : Connexions Redis épuisées

**Diagnostic :**
```bash
# Vérifier les connexions
redis-cli info clients

# Vérifier les connexions par application
redis-cli client list | grep -c "name=veza"
```

**Solutions :**
```bash
# Augmenter le pool de connexions
# Dans la configuration Go
redis.Options{
    PoolSize: 50,
    MinIdleConns: 10,
    MaxRetries: 3,
}
```

### Erreurs NATS

#### Problème : Messages perdus

**Diagnostic :**
```bash
# Vérifier les statistiques NATS
curl -s http://localhost:8222/varz | jq '.'

# Vérifier les connexions
curl -s http://localhost:8222/connz | jq '.'
```

**Solutions :**
```bash
# Configurer la persistance JetStream
nats-server -js -js_domain=veza -js_store_dir=/data/jetstream
```

### Erreurs d'Application

#### Problème : Backend API ne démarre pas

**Diagnostic :**
```bash
# Vérifier les logs
journalctl -u veza-backend-api -f

# Vérifier la configuration
veza-backend-api --config=/etc/veza/config.yaml --validate

# Vérifier les permissions
ls -la /etc/veza/
ls -la /var/log/veza/
```

**Solutions :**
```bash
# Redémarrer le service
sudo systemctl restart veza-backend-api

# Vérifier les dépendances
sudo systemctl status postgresql redis nats-server

# Corriger les permissions
sudo chown -R veza:veza /etc/veza/
sudo chown -R veza:veza /var/log/veza/
```

#### Problème : Chat Server déconnecté

**Diagnostic :**
```bash
# Vérifier les logs
journalctl -u veza-chat-server -f

# Vérifier la connectivité WebSocket
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" \
     -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
     http://localhost:8081/ws
```

**Solutions :**
```bash
# Redémarrer le service
sudo systemctl restart veza-chat-server

# Vérifier la configuration NATS
curl -s http://localhost:8222/varz | jq '.cluster'
```

## Monitoring et Alertes

### Script de Monitoring Automatique

```bash
#!/bin/bash
# scripts/monitor.sh

# Configuration
ALERT_EMAIL="admin@veza.com"
LOG_FILE="/var/log/veza/monitoring.log"

# Fonction d'envoi d'alerte
send_alert() {
    local message="$1"
    echo "$(date): ALERT - $message" >> $LOG_FILE
    echo "$message" | mail -s "Veza Platform Alert" $ALERT_EMAIL
}

# Vérification de la mémoire
check_memory() {
    local memory_usage=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
    if (( $(echo "$memory_usage > 80" | bc -l) )); then
        send_alert "High memory usage: ${memory_usage}%"
    fi
}

# Vérification de l'espace disque
check_disk() {
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ $disk_usage -gt 80 ]; then
        send_alert "High disk usage: ${disk_usage}%"
    fi
}

# Vérification des services
check_services() {
    local services=("veza-backend-api" "veza-chat-server" "veza-stream-server")
    
    for service in "${services[@]}"; do
        if ! systemctl is-active --quiet $service; then
            send_alert "Service $service is down"
        fi
    done
}

# Vérification de la base de données
check_database() {
    if ! pg_isready -h localhost -p 5432 > /dev/null 2>&1; then
        send_alert "PostgreSQL is not responding"
    fi
}

# Vérification de Redis
check_redis() {
    if ! redis-cli ping > /dev/null 2>&1; then
        send_alert "Redis is not responding"
    fi
}

# Exécution des vérifications
main() {
    check_memory
    check_disk
    check_services
    check_database
    check_redis
}

main
```

### Configuration des Alertes Prometheus

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@veza.com'
  smtp_auth_username: 'alertmanager'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team-veza'

receivers:
  - name: 'team-veza'
    email_configs:
      - to: 'admin@veza.com'
        send_resolved: true
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'
        channel: '#veza-alerts'
        send_resolved: true
```

## Récupération de Données

### Script de Récupération

```bash
#!/bin/bash
# scripts/recovery.sh

BACKUP_DIR="/backups/veza"
RECOVERY_DIR="/tmp/veza_recovery_$(date +%Y%m%d_%H%M%S)"

# Créer le répertoire de récupération
mkdir -p $RECOVERY_DIR

# Fonction de récupération PostgreSQL
recover_postgresql() {
    local backup_file="$1"
    
    echo "Recovering PostgreSQL from $backup_file..."
    
    # Arrêter les applications
    systemctl stop veza-backend-api
    systemctl stop veza-chat-server
    systemctl stop veza-stream-server
    
    # Restaurer la base de données
    gunzip -c $backup_file | psql -h localhost -U veza_user -d veza_production
    
    # Redémarrer les applications
    systemctl start veza-backend-api
    systemctl start veza-chat-server
    systemctl start veza-stream-server
    
    echo "PostgreSQL recovery completed"
}

# Fonction de récupération Redis
recover_redis() {
    local backup_file="$1"
    
    echo "Recovering Redis from $backup_file..."
    
    # Arrêter Redis
    systemctl stop redis
    
    # Copier le fichier de sauvegarde
    cp $backup_file /var/lib/redis/dump.rdb
    
    # Redémarrer Redis
    systemctl start redis
    
    echo "Redis recovery completed"
}

# Fonction de récupération des fichiers
recover_files() {
    local backup_file="$1"
    
    echo "Recovering files from $backup_file..."
    
    # Extraire les fichiers
    tar -xzf $backup_file -C $RECOVERY_DIR
    
    # Copier les fichiers de configuration
    cp -r $RECOVERY_DIR/etc/veza/* /etc/veza/
    
    # Copier les fichiers de logs
    cp -r $RECOVERY_DIR/var/log/veza/* /var/log/veza/
    
    echo "Files recovery completed"
}

# Menu de récupération
main() {
    echo "=== Veza Platform Recovery ==="
    echo "1. Recover PostgreSQL"
    echo "2. Recover Redis"
    echo "3. Recover Files"
    echo "4. Full Recovery"
    echo "5. Exit"
    
    read -p "Choose an option: " choice
    
    case $choice in
        1)
            read -p "Enter PostgreSQL backup file: " backup_file
            recover_postgresql "$backup_file"
            ;;
        2)
            read -p "Enter Redis backup file: " backup_file
            recover_redis "$backup_file"
            ;;
        3)
            read -p "Enter files backup: " backup_file
            recover_files "$backup_file"
            ;;
        4)
            echo "Performing full recovery..."
            # Implémenter la récupération complète
            ;;
        5)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo "Invalid option"
            ;;
    esac
}

main
```

## Logs et Debugging

### Analyse des Logs

```bash
#!/bin/bash
# scripts/log_analyzer.sh

LOG_DIR="/var/log/veza"
ANALYSIS_FILE="/tmp/log_analysis_$(date +%Y%m%d_%H%M%S).txt"

# Fonction d'analyse des erreurs
analyze_errors() {
    echo "=== Error Analysis ===" >> $ANALYSIS_FILE
    
    # Erreurs les plus fréquentes
    echo "Most frequent errors:" >> $ANALYSIS_FILE
    grep -i "error" $LOG_DIR/*.log | cut -d' ' -f5- | sort | uniq -c | sort -nr | head -10 >> $ANALYSIS_FILE
    
    # Erreurs par heure
    echo "Errors by hour:" >> $ANALYSIS_FILE
    grep -i "error" $LOG_DIR/*.log | awk '{print $2}' | cut -d: -f1 | sort | uniq -c >> $ANALYSIS_FILE
}

# Fonction d'analyse des performances
analyze_performance() {
    echo "=== Performance Analysis ===" >> $ANALYSIS_FILE
    
    # Temps de réponse
    echo "Response times:" >> $ANALYSIS_FILE
    grep "response_time" $LOG_DIR/*.log | awk '{print $NF}' | sort -n | tail -10 >> $ANALYSIS_FILE
    
    # Requêtes lentes
    echo "Slow queries:" >> $ANALYSIS_FILE
    grep "slow_query" $LOG_DIR/*.log >> $ANALYSIS_FILE
}

# Fonction d'analyse des connexions
analyze_connections() {
    echo "=== Connection Analysis ===" >> $ANALYSIS_FILE
    
    # Connexions par IP
    echo "Connections by IP:" >> $ANALYSIS_FILE
    grep "connection" $LOG_DIR/*.log | awk '{print $NF}' | sort | uniq -c | sort -nr | head -10 >> $ANALYSIS_FILE
}

# Exécution de l'analyse
main() {
    echo "Starting log analysis..."
    
    analyze_errors
    analyze_performance
    analyze_connections
    
    echo "Analysis completed: $ANALYSIS_FILE"
    cat $ANALYSIS_FILE
}

main
```

---

## 🔗 Liens croisés

- [Configuration Avancée](../guides/advanced-configuration.md)
- [Monitoring](../monitoring/README.md)
- [Sécurité](../security/README.md)
- [Déploiement](../deployment/README.md)

---

## Pour aller plus loin

- [Architecture](../architecture/README.md)
- [API Reference](../api/README.md)
- [Database Schema](../database/README.md)
- [Guides](../guides/README.md) 