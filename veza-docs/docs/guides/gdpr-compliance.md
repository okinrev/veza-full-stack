---
title: Conformité RGPD
sidebar_label: RGPD
---

# 🛡️ Conformité RGPD

Ce guide explique la conformité RGPD sur Veza.

## Vue d'ensemble

Ce guide détaille la conformité RGPD (Règlement Général sur la Protection des Données) de la plateforme Veza, incluant les procédures, les outils et les bonnes pratiques.

## 📋 **Principes RGPD**

### 1. Licéité, Loyauté et Transparence
- **Licéité** : Traitement basé sur une base légale
- **Loyauté** : Traitement équitable et non trompeur
- **Transparence** : Information claire sur le traitement

### 2. Limitation des Finalités
- Traitement pour des finalités déterminées
- Compatibilité avec les finalités initiales
- Documentation des finalités

### 3. Minimisation des Données
- Données adéquates, pertinentes et limitées
- Collecte minimale nécessaire
- Suppression des données non essentielles

### 4. Exactitude
- Données exactes et à jour
- Rectification des inexactitudes
- Procédures de mise à jour

### 5. Limitation de Conservation
- Durée de conservation limitée
- Critères de suppression
- Archivage sécurisé

### 6. Intégrité et Confidentialité
- Sécurité technique et organisationnelle
- Protection contre les accès non autorisés
- Chiffrement des données

### 7. Responsabilité
- Documentation des mesures
- Preuve de conformité
- Évaluation continue

## 🔧 **Implémentation Technique**

### Gestion des Consentements

```go
// internal/infrastructure/gdpr/consent.go
package gdpr

import (
    "time"
    "github.com/google/uuid"
)

type ConsentType string

const (
    ConsentMarketing    ConsentType = "marketing"
    ConsentAnalytics    ConsentType = "analytics"
    ConsentNecessary    ConsentType = "necessary"
    ConsentPreferences  ConsentType = "preferences"
)

type Consent struct {
    ID          string      `json:"id"`
    UserID      string      `json:"user_id"`
    Type        ConsentType `json:"type"`
    Granted     bool        `json:"granted"`
    Timestamp   time.Time   `json:"timestamp"`
    IP          string      `json:"ip"`
    UserAgent   string      `json:"user_agent"`
    Version     string      `json:"version"`
    ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
}

type ConsentService struct {
    repo ConsentRepository
    logger *logrus.Logger
}

func (s *ConsentService) RecordConsent(userID string, consentType ConsentType, granted bool, ip, userAgent string) error {
    consent := Consent{
        ID:        uuid.New().String(),
        UserID:    userID,
        Type:      consentType,
        Granted:   granted,
        Timestamp: time.Now(),
        IP:        ip,
        UserAgent: userAgent,
        Version:   "1.0",
    }
    
    // Log pour audit
    s.logger.WithFields(logrus.Fields{
        "event":       "gdpr_consent_change",
        "user_id":     userID,
        "consent_type": consentType,
        "granted":     granted,
        "ip":          ip,
        "timestamp":   time.Now(),
    }).Info("Consent change recorded")
    
    return s.repo.StoreConsent(consent)
}

func (s *ConsentService) GetUserConsents(userID string) ([]Consent, error) {
    return s.repo.GetConsentsByUser(userID)
}

func (s *ConsentService) RevokeConsent(userID string, consentType ConsentType) error {
    // Marquer le consentement comme révoqué
    consent := Consent{
        UserID:    userID,
        Type:      consentType,
        Granted:   false,
        Timestamp: time.Now(),
    }
    
    s.logger.WithFields(logrus.Fields{
        "event":       "gdpr_consent_revoked",
        "user_id":     userID,
        "consent_type": consentType,
        "timestamp":   time.Now(),
    }).Info("Consent revoked")
    
    return s.repo.UpdateConsent(consent)
}
```

### Droit à l'Oubli

```go
// internal/infrastructure/gdpr/right_to_erasure.go
package gdpr

import (
    "time"
    "github.com/sirupsen/logrus"
)

type ErasureRequest struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    RequestedAt time.Time `json:"requested_at"`
    Reason      string    `json:"reason"`
    Status      string    `json:"status"` // pending, processing, completed, failed
    CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ErasureService struct {
    repo   ErasureRepository
    logger *logrus.Logger
}

func (s *ErasureService) RequestErasure(userID, reason string) error {
    request := ErasureRequest{
        ID:          uuid.New().String(),
        UserID:      userID,
        RequestedAt: time.Now(),
        Reason:      reason,
        Status:      "pending",
    }
    
    s.logger.WithFields(logrus.Fields{
        "event":        "gdpr_erasure_requested",
        "user_id":      userID,
        "reason":       reason,
        "request_id":   request.ID,
        "timestamp":    time.Now(),
    }).Info("Erasure requested")
    
    return s.repo.StoreErasureRequest(request)
}

func (s *ErasureService) ProcessErasure(userID string) error {
    // 1. Anonymiser les données utilisateur
    if err := s.anonymizeUserData(userID); err != nil {
        return err
    }
    
    // 2. Supprimer les données personnelles
    if err := s.deletePersonalData(userID); err != nil {
        return err
    }
    
    // 3. Supprimer les consentements
    if err := s.deleteConsents(userID); err != nil {
        return err
    }
    
    // 4. Supprimer les logs d'audit
    if err := s.deleteAuditLogs(userID); err != nil {
        return err
    }
    
    // 5. Marquer comme terminé
    completedAt := time.Now()
    s.repo.UpdateErasureStatus(userID, "completed", &completedAt)
    
    s.logger.WithFields(logrus.Fields{
        "event":     "gdpr_erasure_completed",
        "user_id":   userID,
        "timestamp": time.Now(),
    }).Info("Erasure completed")
    
    return nil
}

func (s *ErasureService) anonymizeUserData(userID string) error {
    // Anonymiser les données dans la base de données
    queries := []string{
        "UPDATE users SET email = CONCAT('anonymized_', id, '@deleted.com') WHERE id = ?",
        "UPDATE users SET username = CONCAT('user_', id) WHERE id = ?",
        "UPDATE users SET first_name = 'Anonymized' WHERE id = ?",
        "UPDATE users SET last_name = 'User' WHERE id = ?",
        "UPDATE users SET phone = NULL WHERE id = ?",
        "UPDATE users SET avatar_url = NULL WHERE id = ?",
    }
    
    for _, query := range queries {
        if err := s.repo.ExecuteQuery(query, userID); err != nil {
            return err
        }
    }
    
    return nil
}

func (s *ErasureService) deletePersonalData(userID string) error {
    // Supprimer les données personnelles sensibles
    tables := []string{
        "user_preferences",
        "user_sessions",
        "user_tokens",
        "user_activity_logs",
    }
    
    for _, table := range tables {
        query := "DELETE FROM " + table + " WHERE user_id = ?"
        if err := s.repo.ExecuteQuery(query, userID); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Droit de Portabilité

```go
// internal/infrastructure/gdpr/data_portability.go
package gdpr

import (
    "encoding/json"
    "time"
)

type DataExport struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    RequestedAt time.Time `json:"requested_at"`
    Status      string    `json:"status"`
    FileURL     string    `json:"file_url,omitempty"`
    ExpiresAt   time.Time `json:"expires_at"`
}

type DataPortabilityService struct {
    repo   DataPortabilityRepository
    logger *logrus.Logger
}

func (s *DataPortabilityService) RequestDataExport(userID string) error {
    export := DataExport{
        ID:          uuid.New().String(),
        UserID:      userID,
        RequestedAt: time.Now(),
        Status:      "pending",
        ExpiresAt:   time.Now().Add(24 * time.Hour), // Expire après 24h
    }
    
    s.logger.WithFields(logrus.Fields{
        "event":      "gdpr_data_export_requested",
        "user_id":    userID,
        "export_id":  export.ID,
        "timestamp":  time.Now(),
    }).Info("Data export requested")
    
    return s.repo.StoreDataExport(export)
}

func (s *DataPortabilityService) GenerateDataExport(userID string) (*DataExport, error) {
    // Collecter toutes les données utilisateur
    userData := s.collectUserData(userID)
    
    // Générer le fichier JSON
    jsonData, err := json.MarshalIndent(userData, "", "  ")
    if err != nil {
        return nil, err
    }
    
    // Sauvegarder le fichier
    filename := "data_export_" + userID + "_" + time.Now().Format("20060102_150405") + ".json"
    fileURL, err := s.saveExportFile(filename, jsonData)
    if err != nil {
        return nil, err
    }
    
    // Mettre à jour le statut
    export := &DataExport{
        UserID:    userID,
        Status:    "completed",
        FileURL:   fileURL,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    s.repo.UpdateDataExport(userID, export)
    
    s.logger.WithFields(logrus.Fields{
        "event":     "gdpr_data_export_completed",
        "user_id":   userID,
        "file_url":  fileURL,
        "timestamp": time.Now(),
    }).Info("Data export completed")
    
    return export, nil
}

func (s *DataPortabilityService) collectUserData(userID string) map[string]interface{} {
    data := make(map[string]interface{})
    
    // Données de base utilisateur
    user, _ := s.repo.GetUser(userID)
    data["user"] = user
    
    // Messages
    messages, _ := s.repo.GetUserMessages(userID)
    data["messages"] = messages
    
    // Préférences
    preferences, _ := s.repo.GetUserPreferences(userID)
    data["preferences"] = preferences
    
    // Consentements
    consents, _ := s.repo.GetUserConsents(userID)
    data["consents"] = consents
    
    // Activité
    activity, _ := s.repo.GetUserActivity(userID)
    data["activity"] = activity
    
    return data
}
```

## 📊 **Monitoring et Conformité**

### Dashboard de Conformité

```json
{
  "dashboard": {
    "title": "RGPD Compliance Dashboard",
    "panels": [
      {
        "title": "Consentements Actifs",
        "type": "visualization",
        "query": {
          "index": "veza-gdpr-*",
          "filter": {
            "event": "consent_granted",
            "timestamp": {
              "gte": "now-30d"
            }
          },
          "aggregation": "count",
          "group_by": "consent_type"
        }
      },
      {
        "title": "Demandes de Suppression",
        "type": "visualization",
        "query": {
          "index": "veza-gdpr-*",
          "filter": {
            "event": "erasure_requested"
          },
          "aggregation": "count",
          "group_by": "status"
        }
      },
      {
        "title": "Exports de Données",
        "type": "visualization",
        "query": {
          "index": "veza-gdpr-*",
          "filter": {
            "event": "data_export_requested"
          },
          "aggregation": "count",
          "group_by": "status"
        }
      }
    ]
  }
}
```

### Alertes de Conformité

```yaml
# prometheus/gdpr_alerts.yml
groups:
- name: gdpr_compliance_alerts
  rules:
  - alert: ConsentExpiring
    expr: gdpr_consent_expiring_total > 0
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Consentements expirant"
      description: "Des consentements utilisateur vont expirer dans les 30 jours"
      
  - alert: ErasureRequestOverdue
    expr: gdpr_erasure_requests_pending_duration_hours > 72
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Demande de suppression en retard"
      description: "Une demande de suppression est en attente depuis plus de 72h"
      
  - alert: DataExportOverdue
    expr: gdpr_data_export_requests_pending_duration_hours > 24
    for: 1m
    labels:
      severity: warning
    annotations:
      summary: "Export de données en retard"
      description: "Une demande d'export de données est en attente depuis plus de 24h"
```

## 📋 **Procédures RGPD**

### Procédure de Gestion des Incidents

```markdown
# Procédure de Gestion des Incidents RGPD

## 1. Détection
- Monitoring automatique des violations
- Signalement par les utilisateurs
- Audit de sécurité

## 2. Évaluation
- Gravité de l'incident
- Nombre d'utilisateurs affectés
- Type de données concernées
- Risque pour les droits et libertés

## 3. Notification
- Notification à l'autorité de contrôle (72h)
- Communication aux personnes concernées
- Documentation de l'incident

## 4. Correction
- Correction immédiate
- Mesures de prévention
- Mise à jour des procédures

## 5. Suivi
- Monitoring post-incident
- Audit de conformité
- Formation des équipes
```

### Procédure d'Impact Assessment

```markdown
# Procédure d'Évaluation d'Impact (DPIA)

## 1. Nécessité de l'évaluation
- Traitement à grande échelle
- Données sensibles
- Surveillance systématique
- Nouvelle technologie

## 2. Contenu de l'évaluation
- Description du traitement
- Finalités et base légale
- Mesures de sécurité
- Évaluation des risques

## 3. Consultation
- DPO (Data Protection Officer)
- Équipe technique
- Juridique
- Utilisateurs (si applicable)

## 4. Documentation
- Rapport d'évaluation
- Mesures de réduction des risques
- Plan de suivi
- Révision périodique
```

## 🛠️ **Outils et Automatisation**

### Script de Nettoyage Automatique

```bash
#!/bin/bash
# gdpr_cleanup.sh

echo "=== Nettoyage RGPD Automatique ==="

# Supprimer les exports expirés
echo "Suppression des exports expirés..."
find /var/exports/ -name "data_export_*.json" -mtime +1 -delete

# Supprimer les logs anonymisés
echo "Suppression des logs anonymisés..."
find /var/log/veza/ -name "*.log" -mtime +90 -exec sed -i 's/[a-zA-Z0-9._%+-]\+@[a-zA-Z0-9.-]\+\.[a-zA-Z]{2,}/[EMAIL_ANONYMIZED]/g' {} \;

# Nettoyer les sessions expirées
echo "Nettoyage des sessions expirées..."
redis-cli --eval /opt/veza/scripts/cleanup_sessions.lua 0

# Vérifier les consentements expirés
echo "Vérification des consentements expirés..."
psql -h localhost -U veza_user -d veza_db -c "
UPDATE consents 
SET granted = false 
WHERE expires_at < NOW() 
AND granted = true;
"

echo "✅ Nettoyage terminé"
```

### Script d'Audit RGPD

```bash
#!/bin/bash
# gdpr_audit.sh

echo "=== Audit RGPD ==="

# Vérifier les consentements
echo "1. Audit des consentements:"
psql -h localhost -U veza_user -d veza_db -c "
SELECT 
    consent_type,
    COUNT(*) as total,
    COUNT(CASE WHEN granted = true THEN 1 END) as granted,
    COUNT(CASE WHEN granted = false THEN 1 END) as denied
FROM consents 
GROUP BY consent_type;
"

# Vérifier les demandes de suppression
echo "2. Audit des demandes de suppression:"
psql -h localhost -U veza_user -d veza_db -c "
SELECT 
    status,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (completed_at - requested_at))/3600) as avg_hours
FROM erasure_requests 
GROUP BY status;
"

# Vérifier les exports de données
echo "3. Audit des exports de données:"
psql -h localhost -U veza_user -d veza_db -c "
SELECT 
    status,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (completed_at - requested_at))/3600) as avg_hours
FROM data_exports 
GROUP BY status;
"

# Vérifier la rétention des données
echo "4. Audit de la rétention:"
psql -h localhost -U veza_user -d veza_db -c "
SELECT 
    'users' as table_name,
    COUNT(*) as record_count,
    MIN(created_at) as oldest_record,
    MAX(created_at) as newest_record
FROM users
UNION ALL
SELECT 
    'messages' as table_name,
    COUNT(*) as record_count,
    MIN(created_at) as oldest_record,
    MAX(created_at) as newest_record
FROM messages;
"

echo "✅ Audit terminé"
```

## 📚 **Formation et Documentation**

### Programme de Formation RGPD

```markdown
# Programme de Formation RGPD

## Modules de Formation

### 1. Principes RGPD (2h)
- Les 7 principes fondamentaux
- Base légale du traitement
- Droits des personnes concernées
- Responsabilité du responsable de traitement

### 2. Implémentation Technique (4h)
- Gestion des consentements
- Droit à l'oubli
- Portabilité des données
- Sécurisation des données

### 3. Procédures Opérationnelles (3h)
- Gestion des incidents
- Évaluation d'impact
- Audit de conformité
- Documentation

### 4. Outils et Monitoring (2h)
- Dashboard de conformité
- Alertes automatiques
- Scripts d'automatisation
- Reporting

## Évaluation
- Quiz de validation des connaissances
- Cas pratiques
- Simulation d'incident
- Audit de conformité
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
**DPO** : dpo@veza.com 