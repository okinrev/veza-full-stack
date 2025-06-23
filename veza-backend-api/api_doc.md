## 🔐 AUTHENTIFICATION

### `POST /login`

* **But** : Connexion utilisateur et récupération des tokens JWT.
* **Headers** : `Content-Type: application/json`
* **Body JSON** :

  ```json
  {
    "email": "test@test.com",
    "password": "supersecret"
  }
  ```
* **Réponse** :

  ```json
  {
    "access_token": "JWT...",
    "refresh_token": "UUID..."
  }
  ```

---

## 🎧 PRODUITS

### `GET /products`

* **But** : Récupérer la liste des produits.
* **Auth** : facultatif ou requis selon ta logique.
* **Réponse** : Liste de produits (`id`, `name`, `description`, `price`, `stock`, etc.)

### `POST /products`

* **But** : Ajouter un nouveau produit.
* **Headers** : `Authorization: Bearer <JWT>`, `Content-Type: application/json`
* **Body JSON** :

  ```json
  {
    "name": "Nom du produit",
    "description": "Description",
    "price": 99.99,
    "stock": 5
  }
  ```

---

## 📂 FICHIERS

### `GET /files`

* **But** : Récupérer la liste des fichiers (uploadés ? internes ?)
* **Auth** : Requiert `Authorization: Bearer <JWT>`

### `POST /files`

* **But** : Envoyer un fichier (probablement multipart/form-data)
* **Payload** : `form-data` avec un champ `file`

---

## 📚 RESSOURCES INTERNES

### `GET /ressources`

* **But** : Lister les ressources internes (PDFs ? tutos ?)
* **Auth** : Oui

---

## 💬 CHAT / MESSAGERIE

### `GET /chat/rooms`

* **But** : Lister tous les salons publics disponibles.
* **Headers** : `Authorization: Bearer <JWT>`
* **Réponse** :

  ```json
  [
    { "id": 1, "name": "afterworks", "created_at": "..." }
  ]
  ```

### `POST /chat/rooms`

* **But** : Créer un nouveau salon.
* **Headers** : `Authorization: Bearer <JWT>`, `Content-Type: application/json`
* **Body JSON** :

  ```json
  {
    "name": "general",
    "is_private": false
  }
  ```

### `GET /chat/rooms/{room}/messages`

* **But** : Récupérer les 50 derniers messages d’un salon.
* **Headers** : `Authorization: Bearer <JWT>`

### `GET /chat/dm/{user_id}`

* **But** : Récupérer les 50 derniers messages privés entre le user connecté et `user_id`.
* **Headers** : `Authorization: Bearer <JWT>`
* **Réponse** :

  ```json
  [
    { "from_user": 3, "to_user": 2, "content": "...", "timestamp": "..." }
  ]
  ```

---

## 🔄 INTERACTIONS WEBSOCKET

* Le WebSocket est à connecter sur `ws://localhost:9001`
* Le token JWT doit être envoyé dans le header `Authorization: Bearer <JWT>`
* Les messages envoyés sont de forme :

  ```json
  { "type": "join", "room": "general" }
  { "type": "message", "room": "general", "content": "salut" }
  { "type": "dm", "to": 2, "content": "yo" }
  { "type": "room_history", "room": "general", "limit": 50 }
  { "type": "dm_history", "with": 2, "limit": 50 }
  ```
