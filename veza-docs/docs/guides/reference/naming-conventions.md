---
id: naming-conventions
title: Conventions de Nommage
sidebar_label: Conventions
---

# Conventions de Nommage - Veza

## Vue d'ensemble

Ce document définit les conventions de nommage utilisées dans Veza.

## Conventions

### Variables
```go
// Go
userName := "john"
userID := 123

// Rust
let user_name = "john";
let user_id = 123;
```

### Fonctions
```go
// Go
func GetUserByID(id int64) (*User, error)

// Rust
fn get_user_by_id(id: u64) -> Result<User, Error>
```

### Constantes
```go
// Go
const MaxRetries = 3
const DefaultTimeout = 30 * time.Second
```

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 