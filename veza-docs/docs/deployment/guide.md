# Guide de Déploiement - Veza Platform

## Vue d'ensemble

Ce guide détaille le processus de déploiement complet de la plateforme Veza, incluant l'environnement de développement, staging et production.

## Architecture de Déploiement

```mermaid
graph TB
    subgraph "Production Environment"
        subgraph "Load Balancer"
            ALB[Application Load Balancer]
            WAF[AWS WAF]
        end
        
        subgraph "Compute Layer"
            EKS[EKS Cluster]
            subgraph "Pods"
                API[Backend API Pods]
                Chat[Chat Server Pods]
                Stream[Stream Server Pods]
            end
        end
        
        subgraph "Data Layer"
            RDS[(RDS PostgreSQL)]
            ElastiCache[(ElastiCache Redis)]
            S3[(S3 Storage)]
        end
        
        subgraph "Message Queue"
            MSK[Amazon MSK NATS]
        end
        
        subgraph "Monitoring"
            CloudWatch[CloudWatch]
            Prometheus[Prometheus]
            Grafana[Grafana]
        end
    end
    
    ALB --> EKS
    WAF --> ALB
    EKS --> API
    EKS --> Chat
    EKS --> Stream
    API --> RDS
    API --> ElastiCache
    Chat --> RDS
    Chat --> ElastiCache
    Stream --> S3
    Stream --> ElastiCache
    API --> MSK
    Chat --> MSK
    Stream --> MSK
    API --> CloudWatch
    Chat --> CloudWatch
    Stream --> CloudWatch
    Prometheus --> Grafana
```

## Prérequis

### Outils Requis

```bash
# Outils de base
- Docker 20.10+
- Docker Compose 2.0+
- kubectl 1.25+
- helm 3.10+
- terraform 1.5+
- aws-cli 2.10+

# Outils de développement
- Go 1.21+
- Rust 1.70+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
```

### Comptes et Services

- **AWS Account** : Pour l'infrastructure cloud
- **Docker Hub** : Pour les images Docker
- **GitHub** : Pour le code source
- **CloudFlare** : Pour le CDN et DNS

## Environnements

### 1. Développement Local

#### Configuration Docker Compose

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  # Base de données
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: veza_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Redis
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # NATS
  nats:
    image: nats:2.9-alpine
    ports:
      - "4222:4222"
    command: -js

  # Backend API
  backend-api:
    build:
      context: ./veza-backend-api
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/veza_dev
      - REDIS_URL=redis://redis:6379
      - NATS_URL=nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats

  # Chat Server
  chat-server:
    build:
      context: ./veza-chat-server
      dockerfile: Dockerfile.dev
    ports:
      - "3001:3001"
    environment:
      - DATABASE_URL=postgres://postgres:password@postgres:5432/veza_dev
      - REDIS_URL=redis://redis:6379
      - NATS_URL=nats://nats:4222
    depends_on:
      - postgres
      - redis
      - nats

  # Stream Server
  stream-server:
    build:
      context: ./veza-stream-server
      dockerfile: Dockerfile.dev
    ports:
      - "3002:3002"
    environment:
      - REDIS_URL=redis://redis:6379
      - NATS_URL=nats://nats:4222
      - S3_ENDPOINT=http://minio:9000
    depends_on:
      - redis
      - nats
      - minio

  # MinIO (S3 compatible)
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ROOT_USER=minioadmin
      - MINIO_ROOT_PASSWORD=minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data

volumes:
  postgres_data:
  redis_data:
  minio_data:
```

#### Démarrage Local

```bash
# Cloner le projet
git clone https://github.com/veza/veza-full-stack.git
cd veza-full-stack

# Copier les variables d'environnement
cp .env.example .env

# Démarrer les services
docker-compose -f docker-compose.dev.yml up -d

# Vérifier les services
docker-compose -f docker-compose.dev.yml ps

# Logs en temps réel
docker-compose -f docker-compose.dev.yml logs -f
```

### 2. Environnement Staging

#### Infrastructure Staging

```hcl
# terraform/staging/main.tf
terraform {
  required_version = ">= 1.5"
  backend "s3" {
    bucket = "veza-terraform-state"
    key    = "staging/terraform.tfstate"
    region = "us-west-2"
  }
}

provider "aws" {
  region = "us-west-2"
}

# VPC
module "vpc" {
  source = "../modules/vpc"
  
  environment = "staging"
  vpc_cidr    = "10.0.0.0/16"
  
  public_subnets  = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets = ["10.0.10.0/24", "10.0.11.0/24"]
}

# EKS Cluster
module "eks" {
  source = "../modules/eks"
  
  cluster_name    = "veza-staging"
  cluster_version = "1.27"
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets
  
  node_groups = {
    general = {
      desired_capacity = 2
      max_capacity     = 4
      min_capacity     = 1
      instance_types   = ["t3.medium"]
    }
  }
}

# RDS
module "rds" {
  source = "../modules/rds"
  
  identifier     = "veza-staging"
  engine_version = "15.4"
  instance_class = "db.t3.micro"
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
}

# ElastiCache
module "elasticache" {
  source = "../modules/elasticache"
  
  cluster_id = "veza-staging"
  node_type  = "cache.t3.micro"
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
}

# S3
module "s3" {
  source = "../modules/s3"
  
  bucket_name = "veza-staging-storage"
  environment = "staging"
}
```

#### Déploiement Staging

```bash
# Configuration Terraform
cd terraform/staging
terraform init
terraform plan
terraform apply

# Configuration kubectl
aws eks update-kubeconfig --region us-west-2 --name veza-staging

# Déploiement avec Helm
helm repo add veza https://charts.veza.com
helm repo update

# Déploiement des services
helm upgrade --install veza-backend veza/backend \
  --namespace veza \
  --create-namespace \
  --values values-staging.yaml

helm upgrade --install veza-chat veza/chat \
  --namespace veza \
  --values values-staging.yaml

helm upgrade --install veza-stream veza/stream \
  --namespace veza \
  --values values-staging.yaml
```

### 3. Environnement Production

#### Infrastructure Production

```hcl
# terraform/production/main.tf
terraform {
  required_version = ">= 1.5"
  backend "s3" {
    bucket = "veza-terraform-state"
    key    = "production/terraform.tfstate"
    region = "us-west-2"
  }
}

provider "aws" {
  region = "us-west-2"
}

# VPC avec haute disponibilité
module "vpc" {
  source = "../modules/vpc"
  
  environment = "production"
  vpc_cidr    = "10.0.0.0/16"
  
  public_subnets  = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  private_subnets = ["10.0.10.0/24", "10.0.11.0/24", "10.0.12.0/24"]
  
  enable_nat_gateways = true
  single_nat_gateway  = false
}

# EKS Cluster haute disponibilité
module "eks" {
  source = "../modules/eks"
  
  cluster_name    = "veza-production"
  cluster_version = "1.27"
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets
  
  node_groups = {
    general = {
      desired_capacity = 3
      max_capacity     = 10
      min_capacity     = 3
      instance_types   = ["m5.large", "m5.xlarge"]
    }
    
    spot = {
      desired_capacity = 2
      max_capacity     = 5
      min_capacity     = 2
      instance_types   = ["m5.large", "m5.xlarge"]
      spot = true
    }
  }
  
  enable_cluster_autoscaler = true
}

# RDS Multi-AZ
module "rds" {
  source = "../modules/rds"
  
  identifier     = "veza-production"
  engine_version = "15.4"
  instance_class = "db.r5.large"
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  multi_az = true
  backup_retention_period = 7
  deletion_protection = true
}

# ElastiCache Cluster
module "elasticache" {
  source = "../modules/elasticache"
  
  cluster_id = "veza-production"
  node_type  = "cache.r5.large"
  num_cache_nodes = 2
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
}

# S3 avec versioning
module "s3" {
  source = "../modules/s3"
  
  bucket_name = "veza-production-storage"
  environment = "production"
  
  versioning = true
  lifecycle_rules = [
    {
      id      = "archive"
      enabled = true
      transition = [
        {
          days          = 90
          storage_class = "STANDARD_IA"
        },
        {
          days          = 365
          storage_class = "GLACIER"
        }
      ]
    }
  ]
}

# CloudFront Distribution
module "cloudfront" {
  source = "../modules/cloudfront"
  
  domain_name = "cdn.veza.com"
  origin_domain_name = module.s3.bucket_regional_domain_name
}
```

#### Déploiement Production

```bash
# Infrastructure
cd terraform/production
terraform init
terraform plan
terraform apply

# Configuration kubectl
aws eks update-kubeconfig --region us-west-2 --name veza-production

# Déploiement avec Helm
helm upgrade --install veza-backend veza/backend \
  --namespace veza \
  --create-namespace \
  --values values-production.yaml \
  --set replicaCount=3

helm upgrade --install veza-chat veza/chat \
  --namespace veza \
  --values values-production.yaml \
  --set replicaCount=3

helm upgrade --install veza-stream veza/stream \
  --namespace veza \
  --values values-production.yaml \
  --set replicaCount=2
```

## CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  AWS_REGION: us-west-2
  EKS_CLUSTER: veza-production

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Setup Rust
        uses: actions/setup-rust@v1
        with:
          rust-version: '1.70'
      
      - name: Run Tests
        run: |
          cd veza-backend-api && go test ./...
          cd veza-chat-server && cargo test
          cd veza-stream-server && cargo test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v1
      
      - name: Build and push Backend API
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: veza-backend-api
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG ./veza-backend-api
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
      
      - name: Build and push Chat Server
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: veza-chat-server
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG ./veza-chat-server
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
      
      - name: Build and push Stream Server
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          ECR_REPOSITORY: veza-stream-server
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG ./veza-stream-server
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Update kubeconfig
        run: aws eks update-kubeconfig --region ${{ env.AWS_REGION }} --name ${{ env.EKS_CLUSTER }}
      
      - name: Deploy with Helm
        run: |
          helm upgrade --install veza-backend ./helm/backend \
            --namespace veza \
            --create-namespace \
            --set image.tag=${{ github.sha }}
          
          helm upgrade --install veza-chat ./helm/chat \
            --namespace veza \
            --set image.tag=${{ github.sha }}
          
          helm upgrade --install veza-stream ./helm/stream \
            --namespace veza \
            --set image.tag=${{ github.sha }}
      
      - name: Wait for deployment
        run: |
          kubectl rollout status deployment/veza-backend -n veza
          kubectl rollout status deployment/veza-chat -n veza
          kubectl rollout status deployment/veza-stream -n veza
```

## Monitoring et Observabilité

### Prometheus Configuration

```yaml
# monitoring/prometheus/config.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: kubernetes_pod_name

  - job_name: 'veza-backend-api'
    static_configs:
      - targets: ['veza-backend-api:8080']
    metrics_path: '/metrics'

  - job_name: 'veza-chat-server'
    static_configs:
      - targets: ['veza-chat-server:3001']
    metrics_path: '/metrics'

  - job_name: 'veza-stream-server'
    static_configs:
      - targets: ['veza-stream-server:3002']
    metrics_path: '/metrics'
```

### Grafana Dashboards

```json
// monitoring/grafana/dashboards/veza-overview.json
{
  "dashboard": {
    "title": "Veza Platform Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Active Connections",
        "type": "stat",
        "targets": [
          {
            "expr": "websocket_connections_active",
            "legendFormat": "{{service}}"
          }
        ]
      }
    ]
  }
}
```

## Sécurité

### Network Policies

```yaml
# k8s/network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: veza-backend-policy
  namespace: veza
spec:
  podSelector:
    matchLabels:
      app: veza-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - namespaceSelector:
        matchLabels:
          name: cache
    ports:
    - protocol: TCP
      port: 6379
```

### RBAC Configuration

```yaml
# k8s/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: veza-backend
  namespace: veza
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: veza-backend-role
  namespace: veza
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: veza-backend-rolebinding
  namespace: veza
subjects:
- kind: ServiceAccount
  name: veza-backend
  namespace: veza
roleRef:
  kind: Role
  name: veza-backend-role
  apiGroup: rbac.authorization.k8s.io
```

## Backup et Récupération

### Stratégie de Backup

```bash
#!/bin/bash
# scripts/backup.sh

# Configuration
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_HOST="veza-production.cluster-xyz.us-west-2.rds.amazonaws.com"
DB_NAME="veza_production"
S3_BUCKET="veza-backups"

# Backup PostgreSQL
pg_dump -h $DB_HOST -U postgres -d $DB_NAME \
  --format=custom --compress=9 \
  --file=$BACKUP_DIR/veza_db_$DATE.dump

# Backup Redis (si nécessaire)
redis-cli -h $REDIS_HOST BGSAVE

# Upload vers S3
aws s3 cp $BACKUP_DIR/veza_db_$DATE.dump s3://$S3_BUCKET/database/

# Nettoyage des anciens backups (garde 30 jours)
find $BACKUP_DIR -name "*.dump" -mtime +30 -delete
```

### Stratégie de Récupération

```bash
#!/bin/bash
# scripts/restore.sh

# Configuration
BACKUP_FILE=$1
DB_HOST="veza-production.cluster-xyz.us-west-2.rds.amazonaws.com"
DB_NAME="veza_production"

# Validation
if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Récupération
pg_restore -h $DB_HOST -U postgres -d $DB_NAME \
  --clean --if-exists \
  $BACKUP_FILE

echo "Récupération terminée"
```

## Maintenance

### Mise à Jour des Services

```bash
#!/bin/bash
# scripts/update.sh

# Mise à jour progressive
kubectl set image deployment/veza-backend \
  veza-backend=veza-backend-api:latest \
  -n veza

# Attendre le déploiement
kubectl rollout status deployment/veza-backend -n veza

# Vérifier la santé
kubectl get pods -n veza -l app=veza-backend

# Répéter pour les autres services
kubectl set image deployment/veza-chat \
  veza-chat=veza-chat-server:latest \
  -n veza

kubectl set image deployment/veza-stream \
  veza-stream=veza-stream-server:latest \
  -n veza
```

### Monitoring de la Santé

```bash
#!/bin/bash
# scripts/health-check.sh

# Vérification des services
SERVICES=("veza-backend" "veza-chat" "veza-stream")

for service in "${SERVICES[@]}"; do
    echo "Vérification de $service..."
    
    # Vérifier les pods
    kubectl get pods -n veza -l app=$service
    
    # Vérifier les endpoints
    kubectl get endpoints -n veza $service
    
    # Test de connectivité
    kubectl exec -n veza deployment/$service -- curl -f http://localhost:8080/health
done
```

---

**Dernière mise à jour** : $(date)
**Version du guide** : 1.0.0 