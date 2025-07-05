---
title: Reprise après Sinistre
sidebar_label: Disaster Recovery
---

# ⚡ Reprise après Sinistre

Ce guide explique la stratégie de disaster recovery sur Veza.

## Vue d'ensemble

Ce guide détaille les procédures de reprise après sinistre (DR) pour la plateforme Veza, incluant les stratégies de sauvegarde, les procédures de restauration et les plans de continuité d'activité.

## 🚨 **Scénarios de Sinistre**

### 1. Panne de Base de Données
- **Impact** : Perte d'accès aux données utilisateur
- **RTO** : 4 heures
- **RPO** : 1 heure

### 2. Panne de Services Backend
- **Impact** : Indisponibilité des API
- **RTO** : 2 heures
- **RPO** : 0 (pas de perte de données)

### 3. Panne d'Infrastructure
- **Impact** : Indisponibilité complète
- **RTO** : 6 heures
- **RPO** : 4 heures

### 4. Attaque de Sécurité
- **Impact** : Compromission des données
- **RTO** : 8 heures
- **RPO** : 24 heures

## 📋 **Plan de Reprise Après Sinistre**

### Phase 1 : Évaluation et Notification

```bash
#!/bin/bash
# disaster_assessment.sh

echo "=== ÉVALUATION DU SINISTRE ==="

# Vérifier l'état des services
services=("veza-backend-api" "veza-chat-server" "veza-stream-server" "postgres" "redis")
for service in "${services[@]}"; do
    if docker ps | grep -q $service; then
        echo "✅ $service: ACTIF"
    else
        echo "❌ $service: INACTIF"
    fi
done

# Vérifier la connectivité réseau
echo "Test de connectivité:"
ping -c 3 google.com

# Vérifier l'espace disque
echo "Espace disque:"
df -h
```

### Phase 2 : Activation du Site de Secours

#### Procédure de Basculement

```bash
# 1. Arrêter les services primaires
docker-compose -f production.yml down

# 2. Activer le site de secours
ssh backup-server "cd /opt/veza && docker-compose -f disaster-recovery.yml up -d"

# 3. Mettre à jour le DNS
aws route53 change-resource-record-sets \
  --hosted-zone-id Z123456789 \
  --change-batch file://dns-failover.json

# 4. Vérifier la santé du site de secours
curl -f https://backup.veza.com/health
```

#### Configuration du Site de Secours

```yaml
# disaster-recovery.yml
version: '3.8'
services:
  veza-backend-api:
    image: veza/backend-api:latest
    environment:
      - DB_HOST=backup-db.veza.com
      - REDIS_HOST=backup-redis.veza.com
      - ENVIRONMENT=disaster-recovery
    ports:
      - "8080:8080"
    restart: unless-stopped

  veza-chat-server:
    image: veza/chat-server:latest
    environment:
      - DB_HOST=backup-db.veza.com
      - REDIS_HOST=backup-redis.veza.com
    ports:
      - "8081:8081"
    restart: unless-stopped

  veza-stream-server:
    image: veza/stream-server:latest
    environment:
      - STORAGE_BACKEND=s3-backup
    ports:
      - "8082:8082"
    restart: unless-stopped
```

### Phase 3 : Restauration des Données

#### Procédure de Restauration PostgreSQL

```bash
#!/bin/bash
# restore_database.sh

BACKUP_FILE="veza_backup_$(date +%Y%m%d_%H%M%S).sql"
BACKUP_S3="s3://veza-backups/database/"

echo "=== RESTAURATION DE LA BASE DE DONNÉES ==="

# 1. Télécharger la dernière sauvegarde
aws s3 cp $BACKUP_S3$BACKUP_FILE ./$BACKUP_FILE

# 2. Arrêter les applications
docker-compose stop veza-backend-api veza-chat-server

# 3. Restaurer la base de données
psql -h localhost -U veza_user -d veza_db -c "DROP SCHEMA public CASCADE;"
psql -h localhost -U veza_user -d veza_db -c "CREATE SCHEMA public;"
psql -h localhost -U veza_user -d veza_db < $BACKUP_FILE

# 4. Vérifier l'intégrité
psql -h localhost -U veza_user -d veza_db -c "SELECT COUNT(*) FROM users;"
psql -h localhost -U veza_user -d veza_db -c "SELECT COUNT(*) FROM messages;"

# 5. Redémarrer les applications
docker-compose start veza-backend-api veza-chat-server

echo "✅ Restauration terminée"
```

#### Procédure de Restauration Redis

```bash
#!/bin/bash
# restore_redis.sh

echo "=== RESTAURATION DU CACHE REDIS ==="

# 1. Sauvegarder l'état actuel
redis-cli BGSAVE

# 2. Restaurer depuis la sauvegarde
cp /var/lib/redis/dump.rdb.backup /var/lib/redis/dump.rdb

# 3. Redémarrer Redis
docker-compose restart redis

# 4. Vérifier la restauration
redis-cli DBSIZE
redis-cli KEYS "*session*" | wc -l
```

### Phase 4 : Validation et Tests

#### Script de Validation

```bash
#!/bin/bash
# validate_recovery.sh

echo "=== VALIDATION DE LA REPRISE ==="

# Tests de connectivité
echo "1. Test de connectivité API:"
curl -f http://localhost:8080/health || echo "❌ API inaccessible"

echo "2. Test de connectivité WebSocket:"
websocat ws://localhost:8081/chat -H "Authorization: Bearer test" || echo "❌ WebSocket inaccessible"

echo "3. Test de connectivité Stream:"
curl -f http://localhost:8082/health || echo "❌ Stream inaccessible"

# Tests de base de données
echo "4. Test de base de données:"
psql -h localhost -U veza_user -d veza_db -c "SELECT COUNT(*) FROM users;" || echo "❌ DB inaccessible"

# Tests de performance
echo "5. Test de performance:"
ab -n 100 -c 10 http://localhost:8080/api/v1/users || echo "❌ Performance dégradée"

echo "✅ Validation terminée"
```

## 💾 **Stratégies de Sauvegarde**

### Sauvegarde Automatique

```bash
#!/bin/bash
# backup_automated.sh

# Configuration
BACKUP_DIR="/opt/backups"
S3_BUCKET="veza-backups"
RETENTION_DAYS=30

# Sauvegarde PostgreSQL
echo "Sauvegarde PostgreSQL..."
pg_dump -h localhost -U veza_user veza_db | gzip > $BACKUP_DIR/veza_db_$(date +%Y%m%d_%H%M%S).sql.gz

# Sauvegarde Redis
echo "Sauvegarde Redis..."
redis-cli BGSAVE
cp /var/lib/redis/dump.rdb $BACKUP_DIR/redis_$(date +%Y%m%d_%H%M%S).rdb

# Sauvegarde des fichiers de configuration
echo "Sauvegarde configuration..."
tar -czf $BACKUP_DIR/config_$(date +%Y%m%d_%H%M%S).tar.gz /etc/veza/

# Upload vers S3
echo "Upload vers S3..."
aws s3 sync $BACKUP_DIR s3://$S3_BUCKET/ --delete

# Nettoyage des anciennes sauvegardes
find $BACKUP_DIR -name "*.gz" -mtime +$RETENTION_DAYS -delete
find $BACKUP_DIR -name "*.rdb" -mtime +$RETENTION_DAYS -delete
```

### Sauvegarde Continue

```yaml
# backup_config.yml
backup:
  database:
    schedule: "0 */6 * * *"  # Toutes les 6 heures
    retention: 7  # 7 jours
    compression: true
    
  redis:
    schedule: "0 */2 * * *"  # Toutes les 2 heures
    retention: 3  # 3 jours
    
  files:
    schedule: "0 0 * * *"    # Quotidien
    retention: 30  # 30 jours
    include:
      - /etc/veza/
      - /var/log/veza/
      - /opt/veza/config/
```

## 🔄 **Procédures de Récupération Spécifiques**

### Récupération Partielle (Services Individuels)

```bash
#!/bin/bash
# partial_recovery.sh

SERVICE=$1

case $SERVICE in
  "api")
    echo "Récupération API..."
    docker-compose restart veza-backend-api
    ;;
  "chat")
    echo "Récupération Chat..."
    docker-compose restart veza-chat-server
    ;;
  "stream")
    echo "Récupération Stream..."
    docker-compose restart veza-stream-server
    ;;
  "db")
    echo "Récupération Base de données..."
    systemctl restart postgresql
    ;;
  *)
    echo "Usage: $0 {api|chat|stream|db}"
    exit 1
    ;;
esac
```

### Récupération de Données Spécifiques

```sql
-- Restauration d'un utilisateur spécifique
INSERT INTO users (id, email, username, created_at)
SELECT id, email, username, created_at
FROM users_backup 
WHERE id = 'user-uuid';

-- Restauration des messages d'une conversation
INSERT INTO messages (id, room_id, user_id, content, created_at)
SELECT id, room_id, user_id, content, created_at
FROM messages_backup 
WHERE room_id = 'room-uuid'
AND created_at >= '2024-01-01';
```

## 📊 **Monitoring de la Reprise**

### Métriques de Récupération

```yaml
# recovery_metrics.yml
metrics:
  - name: recovery_time
    description: "Temps de récupération total"
    unit: "seconds"
    
  - name: data_loss
    description: "Quantité de données perdues"
    unit: "records"
    
  - name: service_availability
    description: "Disponibilité des services après récupération"
    unit: "percentage"
```

### Alertes de Récupération

```yaml
# recovery_alerts.yml
alerts:
  - name: RecoveryTimeExceeded
    condition: recovery_time > 3600  # 1 heure
    severity: critical
    
  - name: DataLossDetected
    condition: data_loss > 0
    severity: high
    
  - name: ServiceUnavailable
    condition: service_availability < 99.9
    severity: critical
```

## 🧪 **Tests de Reprise**

### Plan de Test Annuel

```bash
#!/bin/bash
# annual_dr_test.sh

echo "=== TEST ANNUEL DE REPRISE APRÈS SINISTRE ==="

# 1. Créer un environnement de test
docker-compose -f test-environment.yml up -d

# 2. Simuler un sinistre
docker-compose -f production.yml down

# 3. Activer la procédure de récupération
./disaster_recovery.sh

# 4. Valider la récupération
./validate_recovery.sh

# 5. Documenter les résultats
echo "Résultats du test:" > dr_test_results.txt
date >> dr_test_results.txt
./validate_recovery.sh >> dr_test_results.txt

# 6. Nettoyer l'environnement de test
docker-compose -f test-environment.yml down
```

## 📞 **Contacts d'Urgence**

### Équipe de Reprise

- **Chef de Projet DR** : dr-lead@veza.com
- **Architecte Système** : architect@veza.com
- **DevOps Lead** : devops-lead@veza.com
- **DBA Senior** : dba@veza.com

### Procédure d'Escalade

1. **Niveau 1** (0-2h) : Équipe DevOps
2. **Niveau 2** (2-4h) : Architecte + DBA
3. **Niveau 3** (4-8h) : CTO + Équipe Senior
4. **Niveau 4** (8h+) : Direction + Support externe

## 📚 **Documentation Supplémentaire**

- [Guide de Monitoring](../monitoring/README.md)
- [Guide de Sécurité](../security/security-guide.md)
- [Architecture de la Plateforme](../architecture/backend-architecture.md)
- [Guide de Déploiement](../deployment/guide.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 