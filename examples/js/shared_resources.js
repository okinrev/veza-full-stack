// file: frontend/js/shared_resources.js

function sharedResourcesApp() {
    return {
        // État principal
        username: '',
        resources: [],
        filteredResources: [],
        favorites: JSON.parse(localStorage.getItem('favorites') || '[]'),
        activeTab: 'all',
        viewMode: 'grid',
        showUploadForm: false,
        showPreview: false,
        previewResource: null,

        // Upload
        uploadForm: {
            title: '',
            description: '',
            type: '',
            selectedTags: [],
            isPublic: true,
            file: null
        },
        uploading: false,
        uploadProgress: 0,
        dragOver: false,

        // Recherche et filtres
        search: {
            query: '',
            filters: {
                type: '',
                tag: ''
            },
            sort: 'recent'
        },

        // Tags
        tagInput: '',
        tagSuggestions: [],
        popularTags: [],

        // Pagination
        currentPage: 1,
        itemsPerPage: 12,

        // Stats
        stats: {
            totalResources: 0,
            myUploads: 0,
            totalDownloads: 0,
            popularTag: ''
        },

        // Notifications
        notifications: [],

        // Types de ressources
        resourceTypes: [{
                value: 'sample',
                label: 'Sample',
                icon: '🎵'
            },
            {
                value: 'preset',
                label: 'Preset',
                icon: '🎛️'
            },
            {
                value: 'project',
                label: 'Projet',
                icon: '📁'
            },
            {
                value: 'mix',
                label: 'Mix/Master',
                icon: '🎚️'
            },
            {
                value: 'stem',
                label: 'Stems',
                icon: '🎙️'
            },
            {
                value: 'midi',
                label: 'MIDI',
                icon: '🎹'
            }
        ],

        // Initialisation
        async init() {
            await this.checkAuth();
            await this.loadResources();
            await this.loadPopularTags();
            await this.loadStats();

            // Écouter les changements d'onglet
            this.$watch('activeTab', () => this.filterResources());
            this.$watch('search', () => this.performSearch(), {
                deep: true
            });
        },

        // Authentification
        async checkAuth() {
            const token = localStorage.getItem('access_token');
            if (!token) {
                window.location.href = '/login.html';
                return;
            }

            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                this.username = payload.username;
            } catch (e) {
                this.showNotification('Erreur d\'authentification', 'error');
                window.location.href = '/login.html';
            }
        },

        logout() {
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/login.html';
        },

        // Chargement des ressources
        async loadResources() {
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch('/shared_ressources', {
                    headers: {
                        'Authorization': 'Bearer ' + token
                    }
                });

                if (!response.ok) throw new Error('Erreur de chargement');

                this.resources = await response.json();
                this.filterResources();
            } catch (error) {
                this.showNotification('Erreur lors du chargement des ressources', 'error');
            }
        },

        // Chargement des tags populaires
        async loadPopularTags() {
            try {
                const response = await fetch('/tags');
                if (response.ok) {
                    const tags = await response.json();
                    this.popularTags = tags.slice(0, 10);
                }
            } catch (error) {
                console.error('Erreur chargement tags:', error);
            }
        },

        // Chargement des statistiques
        async loadStats() {
            try {
                const token = localStorage.getItem('access_token');
                const payload = JSON.parse(atob(token.split('.')[1]));

                this.stats.totalResources = this.resources.length;
                this.stats.myUploads = this.resources.filter(r => r.uploader_username === payload.username).length;
                this.stats.totalDownloads = this.resources.reduce((sum, r) => sum + (r.download_count || 0), 0);

                // Tag le plus populaire
                const tagCounts = {};
                this.resources.forEach(r => {
                    r.tags.forEach(tag => {
                        tagCounts[tag] = (tagCounts[tag] || 0) + 1;
                    });
                });
                const sortedTags = Object.entries(tagCounts).sort((a, b) => b[1] - a[1]);
                this.stats.popularTag = sortedTags[0]?.[0] || 'N/A';
            } catch (error) {
                console.error('Erreur stats:', error);
            }
        },

        // Filtrage des ressources
        filterResources() {
            let filtered = [...this.resources];

            // Filtrage par onglet
            if (this.activeTab === 'my') {
                const payload = JSON.parse(atob(localStorage.getItem('access_token').split('.')[1]));
                filtered = filtered.filter(r => r.uploader_username === payload.username);
            } else if (this.activeTab === 'favorites') {
                filtered = filtered.filter(r => this.favorites.includes(r.id));
            } else if (this.activeTab === 'recent') {
                const oneWeekAgo = new Date();
                oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);
                filtered = filtered.filter(r => new Date(r.uploaded_at) > oneWeekAgo);
            }

            // Recherche textuelle
            if (this.search.query) {
                const query = this.search.query.toLowerCase();
                filtered = filtered.filter(r =>
                    r.title.toLowerCase().includes(query) ||
                    r.description?.toLowerCase().includes(query) ||
                    r.uploader_username.toLowerCase().includes(query) ||
                    r.tags.some(tag => tag.toLowerCase().includes(query))
                );
            }

            // Filtrage par type
            if (this.search.filters.type) {
                filtered = filtered.filter(r => r.type === this.search.filters.type);
            }

            // Filtrage par tag
            if (this.search.filters.tag) {
                filtered = filtered.filter(r => r.tags.includes(this.search.filters.tag));
            }

            // Tri
            switch (this.search.sort) {
                case 'popular':
                    filtered.sort((a, b) => (b.download_count || 0) - (a.download_count || 0));
                    break;
                case 'alphabetical':
                    filtered.sort((a, b) => a.title.localeCompare(b.title));
                    break;
                case 'recent':
                default:
                    filtered.sort((a, b) => new Date(b.uploaded_at) - new Date(a.uploaded_at));
            }

            this.filteredResources = filtered;
            this.currentPage = 1;
        },

        // Recherche
        performSearch() {
            this.filterResources();
        },

        toggleFilter(type, value) {
            if (this.search.filters[type] === value) {
                this.search.filters[type] = '';
            } else {
                this.search.filters[type] = value;
            }
        },

        // Gestion des tags
        async searchTags() {
            if (this.tagInput.length < 2) {
                this.tagSuggestions = [];
                return;
            }

            try {
                const response = await fetch(`/tags/search?q=${encodeURIComponent(this.tagInput)}`);
                if (response.ok) {
                    const tags = await response.json();
                    this.tagSuggestions = tags.filter(tag => !this.uploadForm.selectedTags.includes(tag));
                }
            } catch (error) {
                console.error('Erreur recherche tags:', error);
            }
        },

        addTag(tag) {
            if (this.uploadForm.selectedTags.length < 5 && !this.uploadForm.selectedTags.includes(tag)) {
                this.uploadForm.selectedTags.push(tag);
                this.tagInput = '';
                this.tagSuggestions = [];
            }
        },

        addTagFromInput() {
            if (this.tagInput && this.uploadForm.selectedTags.length < 5) {
                this.addTag(this.tagInput.trim());
            }
        },

        removeTag(tag) {
            this.uploadForm.selectedTags = this.uploadForm.selectedTags.filter(t => t !== tag);
        },

        // Upload
        handleDrop(event) {
            this.dragOver = false;
            const files = event.dataTransfer.files;
            if (files.length > 0) {
                this.uploadForm.file = files[0];
            }
        },

        handleFileSelect(event) {
            const files = event.target.files;
            if (files.length > 0) {
                this.uploadForm.file = files[0];
            }
        },

        async uploadResource() {
            if (!this.uploadForm.file || !this.uploadForm.title || !this.uploadForm.type) {
                this.showNotification('Veuillez remplir tous les champs obligatoires', 'error');
                return;
            }

            this.uploading = true;
            this.uploadProgress = 0;

            const formData = new FormData();
            formData.append('file', this.uploadForm.file);
            formData.append('title', this.uploadForm.title);
            formData.append('description', this.uploadForm.description);
            formData.append('type', this.uploadForm.type);
            formData.append('tags', this.uploadForm.selectedTags.join(','));
            formData.append('is_public', this.uploadForm.isPublic);

            try {
                const token = localStorage.getItem('access_token');

                // Simulation de progression
                const progressInterval = setInterval(() => {
                    if (this.uploadProgress < 90) {
                        this.uploadProgress += 10;
                    }
                }, 200);

                const response = await fetch('/shared_ressources', {
                    method: 'POST',
                    headers: {
                        'Authorization': 'Bearer ' + token
                    },
                    body: formData
                });

                clearInterval(progressInterval);
                this.uploadProgress = 100;

                if (response.ok) {
                    this.showNotification('Ressource partagée avec succès!', 'success');
                    this.resetUploadForm();
                    await this.loadResources();
                    this.showUploadForm = false;
                } else {
                    throw new Error('Erreur lors de l\'upload');
                }
            } catch (error) {
                this.showNotification('Erreur lors de l\'upload', 'error');
            } finally {
                this.uploading = false;
                setTimeout(() => this.uploadProgress = 0, 1000);
            }
        },

        resetUploadForm() {
            this.uploadForm = {
                title: '',
                description: '',
                type: '',
                selectedTags: [],
                isPublic: true,
                file: null
            };
            this.tagInput = '';
            this.tagSuggestions = [];
        },

        // Favoris
        toggleFavorite(resourceId) {
            if (this.isFavorite(resourceId)) {
                this.favorites = this.favorites.filter(id => id !== resourceId);
            } else {
                this.favorites.push(resourceId);
            }
            localStorage.setItem('favorites', JSON.stringify(this.favorites));

            if (this.activeTab === 'favorites') {
                this.filterResources();
            }
        },

        isFavorite(resourceId) {
            return this.favorites.includes(resourceId);
        },

        // Prévisualisation
        previewResource(resource) {
            this.previewResource = resource;
            this.showPreview = true;
        },

        // Téléchargement
        async downloadResource(resource) {
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch(`/shared_ressources/${resource.filename}?download=true`, {
                    headers: {
                        'Authorization': 'Bearer ' + token
                    }
                });

                if (response.ok) {
                    const blob = await response.blob();
                    const url = window.URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = resource.filename;
                    a.click();
                    window.URL.revokeObjectURL(url);

                    // Recharger pour mettre à jour le compteur
                    setTimeout(() => this.loadResources(), 1000);
                }
            } catch (error) {
                this.showNotification('Erreur lors du téléchargement', 'error');
            }
        },

        // Partage
        async shareResource(resource) {
            const url = `${window.location.origin}/shared_ressources/${resource.filename}`;

            if (navigator.share) {
                try {
                    await navigator.share({
                        title: resource.title,
                        text: `Découvrez cette ressource: ${resource.title}`,
                        url: url
                    });
                } catch (error) {
                    console.log('Partage annulé');
                }
            } else {
                // Copier dans le presse-papier
                await navigator.clipboard.writeText(url);
                this.showNotification('Lien copié dans le presse-papier!', 'success');
            }
        },

        // Utilitaires
        formatDate(date) {
            return new Date(date).toLocaleDateString('fr-FR', {
                day: 'numeric',
                month: 'short',
                year: 'numeric'
            });
        },

        formatFileSize(bytes) {
            if (!bytes) return 'N/A';
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(1024));
            return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
        },

        isAudioFile(filename) {
            const audioExtensions = ['mp3', 'wav', 'ogg', 'flac', 'm4a'];
            const ext = filename.split('.').pop().toLowerCase();
            return audioExtensions.includes(ext);
        },

        getAudioMimeType(filename) {
            const ext = filename.split('.').pop().toLowerCase();
            const mimeTypes = {
                'mp3': 'audio/mpeg',
                'wav': 'audio/wav',
                'ogg': 'audio/ogg',
                'flac': 'audio/flac',
                'm4a': 'audio/mp4'
            };
            return mimeTypes[ext] || 'audio/mpeg';
        },

        getTypeIcon(type) {
            const typeObj = this.resourceTypes.find(t => t.value === type);
            return typeObj ? typeObj.icon : '📄';
        },

        // Notifications
        showNotification(message, type = 'success') {
            const id = Date.now();
            const notification = {
                id,
                message,
                type,
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
        },

        // Getters calculés
        get paginatedResources() {
            const start = (this.currentPage - 1) * this.itemsPerPage;
            const end = start + this.itemsPerPage;
            return this.filteredResources.slice(start, end);
        },

        get totalPages() {
            return Math.ceil(this.filteredResources.length / this.itemsPerPage);
        },

        get visiblePages() {
            const pages = [];
            const maxVisible = 5;
            let start = Math.max(1, this.currentPage - Math.floor(maxVisible / 2));
            let end = Math.min(this.totalPages, start + maxVisible - 1);

            if (end - start + 1 < maxVisible) {
                start = Math.max(1, end - maxVisible + 1);
            }

            for (let i = start; i <= end; i++) {
                pages.push(i);
            }

            return pages;
        }
    }
} 