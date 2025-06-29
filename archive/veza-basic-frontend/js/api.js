// public/js/api.js

function apiDocsApp() {
    return {
        searchQuery: '',
        methodFilters: [],
        availableMethods: ['GET', 'POST', 'PUT', 'DELETE', 'WEBSOCKET'],
        notifications: [],
        sectorExpanded: {},
        
        // Statistiques
        stats: {
            totalEndpoints: 0,
            totalSectors: 0,
            websocketEndpoints: 0,
            authEndpoints: 0
        },

        sectors: {
            "🔐 Authentification & Utilisateurs": [
                { 
                    method: "POST", 
                    path: "/signup", 
                    description: "Inscrit un **nouvel utilisateur** sur la plateforme Talas. Cette route crée un compte unique avec un nom d'utilisateur, une adresse email et un mot de passe.", 
                    notes: "Requiert un corps JSON avec les champs obligatoires : `username`, `email`, et `password`. En cas de succès, elle retourne des tokens d'accès pour l'authentification future.",
                    auth: false,
                    new: false,
                    parameters: [
                        { name: "username", type: "string", required: true, description: "Nom d'utilisateur unique" },
                        { name: "email", type: "string", required: true, description: "Adresse email valide" },
                        { name: "password", type: "string", required: true, description: "Mot de passe (min 8 caractères)" }
                    ],
                    example: `curl -X POST "http://localhost:8080/signup" \\
-H "Content-Type: application/json" \\
-d '{
"username": "john_doe",
"email": "john@example.com",
"password": "motdepasse123"
}'`,
                    testable: true
                },
                { 
                    method: "POST", 
                    path: "/login", 
                    description: "Authentifie un utilisateur existant. Cette route vérifie les identifiants fournis et, si valides, émet des tokens de session.", 
                    notes: "Requiert un corps JSON avec les champs `email` et `password`. Retourne un `access_token` et un `refresh_token` nécessaires pour les requêtes authentifiées.",
                    auth: false,
                    parameters: [
                        { name: "email", type: "string", required: true, description: "Adresse email du compte" },
                        { name: "password", type: "string", required: true, description: "Mot de passe du compte" }
                    ],
                    example: `curl -X POST "http://localhost:8080/login" \\
-H "Content-Type: application/json" \\
-d '{
"email": "john@example.com",
"password": "motdepasse123"
}'`,
                    testable: true
                },
                { 
                    method: "POST", 
                    path: "/refresh", 
                    description: "Rafraîchit un token d'accès expiré. Cela permet de maintenir une session utilisateur active sans avoir à se reconnecter.", 
                    notes: "Requiert un corps JSON contenant un `refresh_token` valide. Retourne un nouveau `access_token`.",
                    auth: true,
                    parameters: [
                        { name: "refresh_token", type: "string", required: true, description: "Token de rafraîchissement valide" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/users/me", 
                    description: "Récupère toutes les informations du profil de l'utilisateur actuellement authentifié.", 
                    notes: "Nécessite un `access_token` valide dans l'en-tête `Authorization`.",
                    auth: true,
                    example: `curl -X GET "http://localhost:8080/users/me" \\
-H "Authorization: Bearer YOUR_ACCESS_TOKEN"`
                },
                { 
                    method: "PUT", 
                    path: "/users/me", 
                    description: "Met à jour les informations du profil de l'utilisateur connecté, telles que l'adresse email ou le nom d'utilisateur.", 
                    notes: "Nécessite un `access_token`. Peut mettre à jour les champs `email` et/ou `username` via un corps JSON.",
                    auth: true,
                    parameters: [
                        { name: "username", type: "string", required: false, description: "Nouveau nom d'utilisateur" },
                        { name: "email", type: "string", required: false, description: "Nouvelle adresse email" }
                    ]
                },
                { 
                    method: "PUT", 
                    path: "/users/password", 
                    description: "Permet à l'utilisateur authentifié de changer son mot de passe actuel. Une vérification du mot de passe existant est requise.", 
                    notes: "Nécessite un `access_token`. Requiert un corps JSON avec `old_password` et `new_password`.",
                    auth: true,
                    parameters: [
                        { name: "old_password", type: "string", required: true, description: "Mot de passe actuel" },
                        { name: "new_password", type: "string", required: true, description: "Nouveau mot de passe" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/users", 
                    description: "Liste tous les utilisateurs enregistrés sur la plateforme. Accessible uniquement par les administrateurs ou les utilisateurs avec les permissions adéquates.", 
                    notes: "Nécessite un `access_token` valide. Peut inclure des paramètres de pagination ou de filtrage.",
                    auth: true
                },
                { 
                    method: "GET", 
                    path: "/users/except-me", 
                    description: "Récupère une liste de tous les utilisateurs à l'exception de l'utilisateur actuellement connecté. Utile pour les fonctionnalités de messagerie directe.", 
                    notes: "Nécessite un `access_token`. Souvent utilisé pour les listes de sélection pour l'envoi de DM.",
                    auth: true
                },
                { 
                    method: "GET", 
                    path: "/users/search", 
                    description: "Recherche des utilisateurs par une correspondance partielle ou complète avec leur nom d'utilisateur ou leur adresse email.", 
                    notes: "Nécessite un `access_token`. Le paramètre de requête `q` est obligatoire.",
                    auth: true,
                    parameters: [
                        { name: "q", type: "string", required: true, description: "Terme de recherche" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/users/{id}/avatar", 
                    description: "Récupère l'image d'avatar d'un utilisateur spécifique. L'ID de l'utilisateur est passé dans l'URL.", 
                    notes: "L'ID de l'utilisateur est un paramètre de route (`{id}`). Retourne l'image directement.",
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de l'utilisateur" }
                    ]
                },
            ],
            "🎵 Pistes Musicales": [
                { 
                    method: "GET", 
                    path: "/tracks/all", 
                    description: "Récupère une liste complète de toutes les pistes musicales disponibles sur la plateforme. Idéal pour l'affichage initial ou la navigation.", 
                    notes: "Cet endpoint est fréquemment appelé par des interfaces utilisateur pour peupler les listes de lecture."
                },
                { 
                    method: "POST", 
                    path: "/tracks", 
                    description: "Permet à un utilisateur de téléverser une nouvelle piste musicale. Les métadonnées et le fichier audio sont envoyés simultanément.", 
                    notes: "Requiert un formulaire multipart (`multipart/form-data`) avec les champs `title`, `artist`, `tags` (séparés par des virgules), et `audio` (le fichier MP3/WAV, etc.). Nécessite un token d'accès.",
                    auth: true,
                    parameters: [
                        { name: "title", type: "string", required: true, description: "Titre de la piste" },
                        { name: "artist", type: "string", required: true, description: "Nom de l'artiste" },
                        { name: "tags", type: "string", required: false, description: "Tags séparés par des virgules" },
                        { name: "audio", type: "file", required: true, description: "Fichier audio (MP3, WAV, FLAC)" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/stream/{filename}", 
                    description: "Diffuse un fichier audio directement depuis le serveur. Cette route est conçue pour le streaming public ou non-sécurisé.", 
                    notes: "Le nom du fichier audio est passé en paramètre de route. Aucune authentification n'est requise.",
                    parameters: [
                        { name: "filename", type: "string", required: true, description: "Nom du fichier audio" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/stream/signed/{filename}", 
                    description: "Diffuse un fichier audio via une URL sécurisée et signée. La signature garantit l'accès autorisé et limite la durée de validité.", 
                    notes: "Pour le streaming protégé. L'URL contient un horodatage d'expiration (`expires`) et une signature cryptographique (`signature`) pour valider l'accès.",
                    parameters: [
                        { name: "filename", type: "string", required: true, description: "Nom du fichier audio" },
                        { name: "expires", type: "timestamp", required: true, description: "Timestamp d'expiration" },
                        { name: "signature", type: "string", required: true, description: "Signature cryptographique" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/tracks/{id}", 
                    description: "Récupère les détails spécifiques d'une piste musicale individuelle, incluant son titre, artiste, tags, et informations de l'uploader.", 
                    notes: "L'ID de la piste est un paramètre de route (`{id}`).",
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la piste" }
                    ]
                },
                { 
                    method: "PUT", 
                    path: "/tracks/{id}", 
                    description: "Met à jour les informations d'une piste musicale existante. Seul l'uploader ou un administrateur peut modifier la piste.", 
                    notes: "Nécessite un `access_token`. L'ID de la piste est un paramètre de route (`{id}`). Les champs modifiables sont envoyés dans le corps JSON.",
                    auth: true,
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la piste" },
                        { name: "title", type: "string", required: false, description: "Nouveau titre" },
                        { name: "artist", type: "string", required: false, description: "Nouveau nom d'artiste" },
                        { name: "tags", type: "string", required: false, description: "Nouveaux tags" }
                    ]
                },
                { 
                    method: "DELETE", 
                    path: "/tracks/{id}", 
                    description: "Supprime une piste musicale de la plateforme. Cette action est irréversible et nécessite des droits d'accès suffisants.", 
                    notes: "Nécessite un `access_token`. L'ID de la piste est un paramètre de route (`{id}`).",
                    auth: true,
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la piste à supprimer" }
                    ]
                },
            ],
            "📁 Ressources Partagées": [
                { 
                    method: "GET", 
                    path: "/shared_ressources", 
                    description: "Liste toutes les ressources partagées publiquement disponibles sur la plateforme. Cela inclut des documents, images, vidéos, etc.", 
                    notes: "Cet endpoint est souvent appelé par des interfaces utilisateur pour afficher le contenu partagé."
                },
                { 
                    method: "POST", 
                    path: "/shared_ressources", 
                    description: "Uploade une nouvelle ressource à partager publiquement. Cette route permet d'ajouter des fichiers avec leurs métadonnées associées.", 
                    notes: "Requiert un formulaire multipart (`multipart/form-data`) avec les champs `title`, `type`, `tags`, `description`, et `file`. Nécessite un token d'accès.",
                    auth: true,
                    parameters: [
                        { name: "title", type: "string", required: true, description: "Titre de la ressource" },
                        { name: "type", type: "string", required: true, description: "Type de ressource (sample, preset, project, etc.)" },
                        { name: "description", type: "string", required: false, description: "Description de la ressource" },
                        { name: "tags", type: "string", required: false, description: "Tags séparés par des virgules" },
                        { name: "file", type: "file", required: true, description: "Fichier à partager" },
                        { name: "is_public", type: "boolean", required: false, description: "Visibilité publique" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/shared_ressources/{id}", 
                    description: "Récupère les détails et métadonnées d'une ressource partagée spécifique par son identifiant unique.", 
                    notes: "L'ID de la ressource est un paramètre de route (`{id}`).",
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la ressource" }
                    ]
                },
                { 
                    method: "PUT", 
                    path: "/shared_ressources/{id}", 
                    description: "Met à jour les informations d'une ressource partagée existante. Seul l'uploader ou un administrateur peut effectuer cette modification.", 
                    notes: "Nécessite un `access_token`. L'ID de la ressource est un paramètre de route (`{id}`). Les champs à modifier sont envoyés dans le corps JSON.",
                    auth: true,
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la ressource" }
                    ]
                },
                { 
                    method: "DELETE", 
                    path: "/shared_ressources/{id}", 
                    description: "Supprime une ressource partagée de la plateforme. Cette action est irréversible et nécessite des permissions d'administrateur ou d'uploader.", 
                    notes: "Nécessite un `access_token`. L'ID de la ressource est un paramètre de route (`{id}`).",
                    auth: true,
                    parameters: [
                        { name: "id", type: "integer", required: true, description: "ID de la ressource à supprimer" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/shared_ressources/{filename}", 
                    description: "Télécharge un fichier de ressource partagée directement. Cette route est conçue pour faciliter l'accès aux fichiers partagés.", 
                    notes: "Le nom du fichier est passé en paramètre de route. Peut nécessiter une authentification selon les permissions.",
                    parameters: [
                        { name: "filename", type: "string", required: true, description: "Nom du fichier" },
                        { name: "download", type: "boolean", required: false, description: "Force le téléchargement" }
                    ]
                },
            ],
            "🏷️ Tags & Suggestions": [
                { 
                    method: "GET", 
                    path: "/suggestions", 
                    description: "Fournit des suggestions basées sur une requête, utile pour l'autocomplétion ou la découverte de contenu.", 
                    notes: "Le paramètre de requête peut être `tag`, `title`, `uploader`, `doc`, `user`, ou `track` pour des suggestions filtrées.",
                    parameters: [
                        { name: "tag", type: "string", required: false, description: "Recherche de tags" },
                        { name: "title", type: "string", required: false, description: "Recherche de titres" },
                        { name: "user", type: "string", required: false, description: "Recherche d'utilisateurs" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/tags", 
                    description: "Récupère une liste de tous les tags existants sur la plateforme, utilisés pour categoriser les pistes, ressources, etc.", 
                    notes: "Utilisé pour afficher toutes les options de tags disponibles pour la recherche ou le filtrage."
                },
                { 
                    method: "GET", 
                    path: "/tags/search", 
                    description: "Recherche des tags spécifiques en fonction d'une chaîne de caractères. Utile pour filtrer les tags ou trouver des correspondances exactes/partielles.", 
                    notes: "Le paramètre de requête `q` est utilisé pour la recherche de tags.",
                    parameters: [
                        { name: "q", type: "string", required: true, description: "Terme de recherche" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/tags/popular", 
                    description: "Récupère une liste des tags les plus fréquemment utilisés ou les plus populaires sur la plateforme, classés par pertinence ou nombre d'occurrences.", 
                    notes: "Idéal pour identifier les sujets ou catégories tendances."
                },
            ],
            "🏷️ Listings & Offers": [
                { 
                    method: "POST", 
                    path: "/listings", 
                    description: "Permet de créer une nouvelle offfre de troc.", 
                    notes: "Le paramètre de requête peut être `tag`, `title`, `uploader`, `doc`, `user`, ou `track` pour des suggestions filtrées.",
                    parameters: [
                        { name: "tag", type: "string", required: false, description: "Recherche de tags" },
                        { name: "title", type: "string", required: false, description: "Recherche de titres" },
                        { name: "user", type: "string", required: false, description: "Recherche d'utilisateurs" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/listings", 
                    description: "Récupère une liste de tous les tags existants sur la plateforme, utilisés pour categoriser les pistes, ressources, etc.", 
                    notes: "Utilisé pour afficher toutes les options de tags disponibles pour la recherche ou le filtrage."
                },
                { 
                    method: "GET", 
                    path: "/listings/{id:[0-9]+}", 
                    description: "Recherche des tags spécifiques en fonction d'une chaîne de caractères. Utile pour filtrer les tags ou trouver des correspondances exactes/partielles.", 
                    notes: "Le paramètre de requête `q` est utilisé pour la recherche de tags.",
                    parameters: [
                        { name: "q", type: "string", required: true, description: "Terme de recherche" }
                    ]
                },
                { 
                    method: "DELETE", 
                    path: "/listings/{id:[0-9]+}", 
                    description: "Récupère une liste des tags les plus fréquemment utilisés ou les plus populaires sur la plateforme, classés par pertinence ou nombre d'occurrences.", 
                    notes: "Idéal pour identifier les sujets ou catégories tendances."
                },
                { 
                    method: "POST", 
                    path: "/listings/{id:[0-9]+}/offer", 
                    description: "Recherche des tags spécifiques en fonction d'une chaîne de caractères. Utile pour filtrer les tags ou trouver des correspondances exactes/partielles.", 
                    notes: "Le paramètre de requête `q` est utilisé pour la recherche de tags.",
                    parameters: [
                        { name: "q", type: "string", required: true, description: "Terme de recherche" }
                    ]
                },
                { 
                    method: "POST", 
                    path: "/offers/{id:[0-9]+}/accept", 
                    description: "Récupère une liste des tags les plus fréquemment utilisés ou les plus populaires sur la plateforme, classés par pertinence ou nombre d'occurrences.", 
                    notes: "Idéal pour identifier les sujets ou catégories tendances."
                },
            ],
            "💬 Chat & WebSockets": [
                { 
                    method: "WEBSOCKET", 
                    path: "ws://localhost:9001/?token={token}", 
                    description: "Établit une connexion WebSocket pour les messages directs (DM) et les salons de discussion. Permet l'envoi et la réception de messages en temps réel.", 
                    notes: "Nécessite un token JWT valide passé comme paramètre de requête `token`. Les messages sont échangés via cette connexion persistante.",
                    auth: true,
                    parameters: [
                        { name: "token", type: "string", required: true, description: "Token JWT d'authentification" }
                    ],
                    example: `// Connexion WebSocket
const token = localStorage.getItem('access_token');
const socket = new WebSocket(\`ws://localhost:9001/?token=\${token}\`);

// Envoyer un message privé
socket.send(JSON.stringify({
type: "dm",
to: 123,
content: "Hello!"
}));

// Rejoindre un salon
socket.send(JSON.stringify({
type: "join",
room: "general"
}));`
                },
                { 
                    method: "GET", 
                    path: "/chat/rooms", 
                    description: "Liste tous les salons de discussion publics disponibles sur la plateforme. Utile pour permettre aux utilisateurs de rejoindre ou de découvrir des salons.", 
                    notes: "Utilisé pour afficher les salons auxquels un utilisateur peut participer.",
                    auth: true
                },
                { 
                    method: "POST", 
                    path: "/chat/rooms", 
                    description: "Crée un nouveau salon de discussion, qui peut être public ou privé. L'utilisateur qui crée le salon en devient l'administrateur.", 
                    notes: "Requiert un corps JSON avec les champs `name` (nom du salon) et optionnellement `description`. Nécessite un token d'accès.",
                    auth: true,
                    parameters: [
                        { name: "name", type: "string", required: true, description: "Nom du salon" },
                        { name: "description", type: "string", required: false, description: "Description du salon" },
                        { name: "is_private", type: "boolean", required: false, description: "Salon privé ou public" }
                    ]
                },
                { 
                    method: "GET", 
                    path: "/generate-stream-url", 
                    description: "Génère une URL sécurisée et signée pour le streaming de fichiers audio. Utilisé pour l'écoute sécurisée.", 
                    notes: "Nécessite un token d'accès. Le paramètre `filename` spécifie le fichier à streamer.",
                    auth: true,
                    parameters: [
                        { name: "filename", type: "string", required: true, description: "Nom du fichier à streamer" }
                    ]
                },
            ],
        },

        // Secteurs filtrés (calculé dynamiquement)
        filteredSectors: {},
        filteredCount: 0,

        init() {
            // Initialiser l'état d'expansion des secteurs
            for (const sectorName in this.sectors) {
                this.sectorExpanded[sectorName] = true; // Tous ouverts par défaut
                
                // Initialiser showDescription pour chaque route
                this.sectors[sectorName].forEach(route => {
                    route.showDescription = false;
                });
            }
            
            // Calculer les statistiques
            this.calculateStats();
            
            // Initialiser les secteurs filtrés
            this.filterRoutes();
        },

        calculateStats() {
            let totalEndpoints = 0;
            let websocketEndpoints = 0;
            let authEndpoints = 0;

            for (const sectorName in this.sectors) {
                const routes = this.sectors[sectorName];
                totalEndpoints += routes.length;
                
                routes.forEach(route => {
                    if (route.method === 'WEBSOCKET') {
                        websocketEndpoints++;
                    }
                    if (route.auth) {
                        authEndpoints++;
                    }
                });
            }

            this.stats = {
                totalEndpoints,
                totalSectors: Object.keys(this.sectors).length,
                websocketEndpoints,
                authEndpoints
            };
        },

        filterRoutes() {
            this.filteredSectors = {};
            this.filteredCount = 0;

            for (const sectorName in this.sectors) {
                const filteredRoutes = this.sectors[sectorName].filter(route => {
                    // Filtre par méthode
                    if (this.methodFilters.length > 0 && !this.methodFilters.includes(route.method)) {
                        return false;
                    }

                    // Filtre par recherche textuelle
                    if (this.searchQuery) {
                        const query = this.searchQuery.toLowerCase();
                        return (
                            route.path.toLowerCase().includes(query) ||
                            route.description.toLowerCase().includes(query) ||
                            route.method.toLowerCase().includes(query) ||
                            (route.notes && route.notes.toLowerCase().includes(query))
                        );
                    }

                    return true;
                });

                if (filteredRoutes.length > 0) {
                    this.filteredSectors[sectorName] = filteredRoutes;
                    this.filteredCount += filteredRoutes.length;
                }
            }
        },

        toggleMethodFilter(method) {
            const index = this.methodFilters.indexOf(method);
            if (index > -1) {
                this.methodFilters.splice(index, 1);
            } else {
                this.methodFilters.push(method);
            }
            this.filterRoutes();
        },

        clearFilters() {
            this.searchQuery = '';
            this.methodFilters = [];
            this.filterRoutes();
        },

        toggleSector(sectorName) {
            this.sectorExpanded[sectorName] = !this.sectorExpanded[sectorName];
        },

        expandAll() {
            for (const sectorName in this.filteredSectors) {
                this.sectorExpanded[sectorName] = true;
                this.filteredSectors[sectorName].forEach(route => {
                    route.showDescription = true;
                });
            }
        },

        collapseAll() {
            for (const sectorName in this.filteredSectors) {
                this.sectorExpanded[sectorName] = false;
                this.filteredSectors[sectorName].forEach(route => {
                    route.showDescription = false;
                });
            }
        },

        toggleDescription(sectorName, index) {
            const route = this.filteredSectors[sectorName][index];
            route.showDescription = !route.showDescription;
        },

        formatDescription(text) {
            if (!text) return '';
            
            // Convertir le markdown basique en HTML
            return text
                .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                .replace(/`(.*?)`/g, '<code class="bg-gray-200 px-1 rounded">$1</code>')
                .replace(/\n/g, '<br>');
        },

        async copyToClipboard(text) {
            try {
                await navigator.clipboard.writeText(text);
                this.showNotification('URL copiée dans le presse-papier !');
            } catch (err) {
                console.error('Erreur lors de la copie:', err);
            }
        },

        showNotification(message) {
            const id = Date.now();
            const notification = {
                id,
                message,
                show: true
            };
            this.notifications.push(notification);

            setTimeout(() => {
                const index = this.notifications.findIndex(n => n.id === id);
                if (index > -1) {
                    this.notifications[index].show = false;
                    setTimeout(() => {
                        this.notifications = this.notifications.filter(n => n.id !== id);
                    }, 300);
                }
            }, 3000);
        }
    }
}