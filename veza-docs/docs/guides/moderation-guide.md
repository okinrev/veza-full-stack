---
id: moderation-guide
title: Guide de Modération - Veza Platform
sidebar_label: Guide de Modération
---

# Guide de Modération - Veza Platform

> **Guide complet pour la modération automatique et manuelle de la plateforme Veza**

## Vue d'ensemble

La modération sur Veza combine des systèmes automatiques et manuels pour maintenir un environnement sûr et respectueux.

## Système de Modération Automatique

### Filtres de Contenu

#### Filtres de Texte

```python
# Exemple de configuration des filtres
MODERATION_FILTERS = {
    'spam': {
        'enabled': True,
        'threshold': 0.8,
        'patterns': [
            r'(buy|sell|click|free|money|earn)',
            r'(http|https)://[^\s]+',
            r'[A-Z]{5,}'
        ]
    },
    'toxicity': {
        'enabled': True,
        'threshold': 0.7,
        'ai_model': 'openai/gpt-3.5-turbo'
    },
    'harassment': {
        'enabled': True,
        'threshold': 0.6,
        'keywords': [
            'hate', 'discrimination', 'bullying'
        ]
    }
}
```

#### Détection de Spam

```rust
// src/moderation/spam_detection.rs
pub struct SpamDetector {
    patterns: Vec<Regex>,
    threshold: f32,
}

impl SpamDetector {
    pub fn new() -> Self {
        let patterns = vec![
            Regex::new(r"(buy|sell|click|free|money|earn)").unwrap(),
            Regex::new(r"(http|https)://[^\s]+").unwrap(),
            Regex::new(r"[A-Z]{5,}").unwrap(),
        ];
        
        Self {
            patterns,
            threshold: 0.8,
        }
    }
    
    pub fn detect_spam(&self, text: &str) -> SpamResult {
        let mut score = 0.0;
        let mut matches = Vec::new();
        
        for pattern in &self.patterns {
            if pattern.is_match(text) {
                score += 0.3;
                matches.push(pattern.as_str());
            }
        }
        
        SpamResult {
            is_spam: score >= self.threshold,
            score,
            matches,
        }
    }
}
```

### Modération IA

#### Intégration OpenAI

```go
// internal/moderation/ai_service.go
package moderation

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type AIService struct {
    client  *http.Client
    apiKey  string
    baseURL string
}

type ModerationRequest struct {
    Input string `json:"input"`
}

type ModerationResponse struct {
    Results []ModerationResult `json:"results"`
}

type ModerationResult struct {
    Categories Categories `json:"categories"`
    CategoryScores CategoryScores `json:"category_scores"`
    Flagged bool `json:"flagged"`
}

func (ai *AIService) ModerateText(ctx context.Context, text string) (*ModerationResult, error) {
    req := ModerationRequest{Input: text}
    
    reqBody, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    httpReq, err := http.NewRequestWithContext(
        ctx,
        "POST",
        fmt.Sprintf("%s/v1/moderations", ai.baseURL),
        bytes.NewBuffer(reqBody),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    httpReq.Header.Set("Authorization", "Bearer "+ai.apiKey)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := ai.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    var moderationResp ModerationResponse
    if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    if len(moderationResp.Results) == 0 {
        return nil, fmt.Errorf("no moderation results")
    }
    
    return &moderationResp.Results[0], nil
}
```

#### Configuration des Seuils

```yaml
# config/moderation.yaml
moderation:
  ai:
    enabled: true
    provider: "openai"
    model: "gpt-3.5-turbo"
    api_key: "${OPENAI_API_KEY}"
    
  thresholds:
    toxicity: 0.7
    harassment: 0.6
    spam: 0.8
    violence: 0.5
    sexual_content: 0.6
    
  actions:
    auto_delete: true
    auto_warn: true
    auto_ban: false
    notify_moderators: true
    
  filters:
    spam:
      enabled: true
      patterns:
        - "(buy|sell|click|free|money|earn)"
        - "(http|https)://[^\\s]+"
        - "[A-Z]{5,}"
    
    profanity:
      enabled: true
      word_list: "config/profanity_list.txt"
      
    links:
      enabled: true
      allowed_domains:
        - "veza.com"
        - "youtube.com"
        - "spotify.com"
```

### Système de Signaux

#### Structure des Signaux

```rust
// src/moderation/signal.rs
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Signal {
    pub id: String,
    pub user_id: String,
    pub content_id: String,
    pub content_type: ContentType,
    pub reason: SignalReason,
    pub description: String,
    pub evidence: Vec<String>,
    pub reporter_id: Option<String>,
    pub created_at: DateTime<Utc>,
    pub status: SignalStatus,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ContentType {
    Message,
    Stream,
    Comment,
    Profile,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SignalReason {
    Spam,
    Harassment,
    HateSpeech,
    Violence,
    SexualContent,
    Copyright,
    Impersonation,
    Other(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SignalStatus {
    Pending,
    UnderReview,
    Resolved,
    Dismissed,
}
```

#### API de Signalement

```http
POST /api/v1/moderation/signals
```

**Headers :**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Body :**
```json
{
  "content_id": "message_123",
  "content_type": "message",
  "reason": "harassment",
  "description": "Contenu harcelant",
  "evidence": ["screenshot_url"]
}
```

**Réponse :**
```json
{
  "signal_id": "signal_456",
  "status": "pending",
  "created_at": "2024-01-15T10:00:00Z",
  "estimated_review_time": "2-4 hours"
}
```

## Interface de Modération

### Dashboard Modérateur

```typescript
// src/components/ModerationDashboard.tsx
interface ModerationDashboardProps {
  signals: Signal[];
  onReview: (signalId: string, action: ModerationAction) => void;
}

const ModerationDashboard: React.FC<ModerationDashboardProps> = ({
  signals,
  onReview
}) => {
  const [filter, setFilter] = useState<SignalFilter>('pending');
  const [selectedSignal, setSelectedSignal] = useState<Signal | null>(null);
  
  return (
    <div className="moderation-dashboard">
      <div className="filters">
        <select value={filter} onChange={(e) => setFilter(e.target.value)}>
          <option value="pending">En attente</option>
          <option value="under_review">En cours</option>
          <option value="resolved">Résolus</option>
        </select>
      </div>
      
      <div className="signals-list">
        {signals.map(signal => (
          <SignalCard
            key={signal.id}
            signal={signal}
            onClick={() => setSelectedSignal(signal)}
          />
        ))}
      </div>
      
      {selectedSignal && (
        <SignalDetail
          signal={selectedSignal}
          onAction={onReview}
          onClose={() => setSelectedSignal(null)}
        />
      )}
    </div>
  );
};
```

### Actions de Modération

```typescript
// src/types/moderation.ts
export enum ModerationAction {
  WARN = 'warn',
  DELETE = 'delete',
  BAN_TEMPORARY = 'ban_temporary',
  BAN_PERMANENT = 'ban_permanent',
  DISMISS = 'dismiss',
}

export interface ModerationDecision {
  signalId: string;
  action: ModerationAction;
  reason: string;
  duration?: number; // Pour les bannissements temporaires
  moderatorId: string;
}
```

## Système de Bannissement

### Types de Bannissement

```rust
// src/moderation/ban.rs
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Ban {
    pub id: String,
    pub user_id: String,
    pub reason: String,
    pub ban_type: BanType,
    pub duration: Option<Duration>,
    pub moderator_id: String,
    pub created_at: DateTime<Utc>,
    pub expires_at: Option<DateTime<Utc>>,
    pub is_active: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum BanType {
    Warning,
    Temporary,
    Permanent,
}

impl Ban {
    pub fn is_expired(&self) -> bool {
        if let Some(expires_at) = self.expires_at {
            Utc::now() > expires_at
        } else {
            false
        }
    }
    
    pub fn deactivate(&mut self) {
        self.is_active = false;
    }
}
```

### API de Bannissement

```http
POST /api/v1/moderation/bans
```

**Body :**
```json
{
  "user_id": "user_123",
  "reason": "Harcèlement répété",
  "ban_type": "temporary",
  "duration_hours": 24,
  "evidence": ["signal_456", "signal_789"]
}
```

**Réponse :**
```json
{
  "ban_id": "ban_123",
  "user_id": "user_123",
  "expires_at": "2024-01-16T10:00:00Z",
  "created_at": "2024-01-15T10:00:00Z"
}
```

## Métriques et Analytics

### Dashboard de Modération

```sql
-- Requêtes pour les métriques de modération
-- Signaux par jour
SELECT 
    DATE(created_at) as date,
    COUNT(*) as signals_count,
    COUNT(*) FILTER (WHERE status = 'resolved') as resolved_count
FROM signals 
WHERE created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date;

-- Temps de résolution moyen
SELECT 
    AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600) as avg_resolution_hours
FROM signals 
WHERE status = 'resolved' 
AND resolved_at IS NOT NULL;

-- Signaux par raison
SELECT 
    reason,
    COUNT(*) as count,
    COUNT(*) * 100.0 / SUM(COUNT(*)) OVER () as percentage
FROM signals 
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY reason
ORDER BY count DESC;
```

### Alertes Automatiques

```yaml
# config/alerts/moderation.yaml
alerts:
  - name: "High Signal Volume"
    condition: "signals_per_hour > 50"
    severity: "warning"
    notification:
      - email: "moderators@veza.com"
      - slack: "#moderation-alerts"
  
  - name: "Slow Resolution Time"
    condition: "avg_resolution_hours > 4"
    severity: "warning"
    notification:
      - email: "moderators@veza.com"
  
  - name: "Spam Attack"
    condition: "spam_signals_per_minute > 10"
    severity: "critical"
    notification:
      - email: "admin@veza.com"
      - slack: "#urgent-alerts"
    actions:
      - "enable_auto_moderation"
      - "increase_spam_threshold"
```

## Formation des Modérateurs

### Guide de Décision

| Type de Contenu | Action Recommandée | Durée |
|-----------------|-------------------|-------|
| Spam léger | Avertissement | - |
| Harcèlement | Bannissement temporaire | 24-72h |
| Discours de haine | Bannissement permanent | - |
| Violence | Bannissement permanent | - |
| Contenu sexuel | Suppression + avertissement | - |
| Copyright | Suppression + avertissement | - |

### Checklist de Modération

```markdown
## Checklist de Modération

### Avant de Prendre une Décision
- [ ] Lire le contenu complet
- [ ] Vérifier le contexte
- [ ] Consulter l'historique de l'utilisateur
- [ ] Vérifier les preuves fournies

### Actions à Prendre
- [ ] Choisir l'action appropriée
- [ ] Rédiger une raison claire
- [ ] Notifier l'utilisateur
- [ ] Documenter la décision

### Après l'Action
- [ ] Vérifier que l'action a été appliquée
- [ ] Surveiller les réactions
- [ ] Ajuster si nécessaire
```

## Intégration avec l'Écosystème

### Webhooks de Modération

```http
POST /webhooks/moderation
```

**Headers :**
```
X-Veza-Signature: <signature>
Content-Type: application/json
```

**Body :**
```json
{
  "event": "signal_created",
  "data": {
    "signal_id": "signal_123",
    "content_type": "message",
    "reason": "harassment",
    "user_id": "user_456",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### Intégration Chat

```rust
// src/chat/moderation_integration.rs
pub struct ChatModeration {
    detector: SpamDetector,
    ai_service: AIService,
}

impl ChatModeration {
    pub async fn moderate_message(&self, message: &Message) -> ModerationResult {
        // Vérification spam
        let spam_result = self.detector.detect_spam(&message.content);
        if spam_result.is_spam {
            return ModerationResult::Blocked(ModerationReason::Spam);
        }
        
        // Vérification IA
        let ai_result = self.ai_service.moderate_text(&message.content).await?;
        if ai_result.flagged {
            return ModerationResult::Flagged(ModerationReason::Toxicity);
        }
        
        ModerationResult::Approved
    }
}
```

---

## 🔗 Liens croisés

- [Sécurité](../security/README.md)
- [Chat API](../api/websocket/README.md)
- [Monitoring](../monitoring/alerts/alerting-guide.md)
- [Troubleshooting](../troubleshooting/README.md)

---

## Pour aller plus loin

- [Configuration Avancée](../guides/advanced-configuration.md)
- [API Reference](../api/README.md)
- [Base de Données](../database/schema.md)
- [Déploiement](../deployment/README.md) 