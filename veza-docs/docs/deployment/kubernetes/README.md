---
id: kubernetes-readme
title: Kubernetes
sidebar_label: Kubernetes
---

# Déploiement Kubernetes - Veza

## Vue d'ensemble

Ce document décrit le déploiement de Veza sur Kubernetes.

## Manifests Kubernetes

### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veza-backend-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: veza-backend-api
```

### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: veza-backend-service
spec:
  ports:
  - port: 8080
    targetPort: 8080
```

## Commandes

```bash
# Déployer
kubectl apply -f k8s/

# Vérifier le statut
kubectl get pods

# Logs
kubectl logs -f deployment/veza-backend-api
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 