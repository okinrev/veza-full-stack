// public/js/app.js

function unifiedChatApp() {
    return {
        // État principal
        username: '',
        myUserId: null,
        isConnected: false,
        socket: null,
        
        // Navigation
        activeTab: 'rooms', // 'rooms' ou 'dm'
        
        // Messages et interface
        messages: [],
        messageContent: '',
        loadingMessages: false,
        sending: false,
        notificationsEnabled: true,
        
        // Salons publics
        rooms: [],
        currentRoom: '',
        connectedUsers: [],
        newRoomName: '',
        creating: false,
        isTyping: [],
        typingTimeout: null,
        
        // Messages privés
        allUsers: [],
        filteredUsers: [],
        userSearch: '',
        loadingUsers: false,
        otherUserId: null,
        otherUserInfo: {
            username: '',
            isOnline: false
        },
        recentConversations: [],
        
        // Stats
        roomStats: {
            totalRooms: 0,
            activeUsers: 0,
            todayMessages: 0
        },
        dmStats: {
            totalMessages: 0,
            todayMessages: 0
        },
        
        // Notifications
        notifications: [],

        // Initialisation
        async init() {
            console.log('Initialisation Chat Unifié...');
            
            // Récupérer l'onglet depuis l'URL
            const urlParams = new URLSearchParams(window.location.search);
            const tab = urlParams.get("tab");
            const userId = urlParams.get("user_id");
            
            if (tab === 'dm' || userId) {
                this.activeTab = 'dm';
                if (userId) {
                    this.otherUserId = parseInt(userId);
                }
            }
            
            const authSuccess = await this.checkAuth();
            if (!authSuccess) return;
            
            await this.loadUserInfo();
            await this.loadRooms();
            
            if (this.activeTab === 'dm') {
                await this.loadUsers();
            }
            
            await this.connectWebSocket();
            this.requestNotificationPermission();
            
            // Watcher pour changement d'onglet
            this.$watch('activeTab', (newTab) => {
                this.switchTab(newTab);
            });
            
            console.log('Initialisation terminée');
        },

        // Authentification
        async checkAuth() {
            const token = localStorage.getItem('access_token');
            if (!token) {
                window.location.href = '/login.html';
                return false;
            }

            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                this.username = payload.username;
                
                if (payload.exp && payload.exp < Date.now() / 1000) {
                    localStorage.removeItem('access_token');
                    window.location.href = '/login.html';
                    return false;
                }
                
                return true;
            } catch (e) {
                this.showNotification('Erreur d\'authentification', 'error');
                window.location.href = '/login.html';
                return false;
            }
        },

        logout() {
            if (this.socket) {
                this.socket.onclose = null;
                this.socket.onerror = null;
                this.socket.close();
                this.socket = null;
            }
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            window.location.href = '/login.html';
        },

        goBack() {
            if (window.history.length > 1) {
                window.history.back();
            } else {
                window.location.href = '/';
            }
        },

        // Chargement des informations utilisateur
        async loadUserInfo() {
            try {
                const token = localStorage.getItem('access_token');
                
                try {
                    const meResponse = await fetch('/me', {
                        headers: { 'Authorization': 'Bearer ' + token }
                    });
                    
                    if (meResponse.ok) {
                        const meData = await meResponse.json();
                        this.myUserId = meData.id;
                    } else {
                        const payload = JSON.parse(atob(token.split('.')[1]));
                        this.myUserId = payload.user_id || payload.id || payload.sub;
                    }
                } catch (error) {
                    const payload = JSON.parse(atob(token.split('.')[1]));
                    this.myUserId = payload.user_id || payload.id || payload.sub;
                }

                // Charger les infos de l'autre utilisateur si en mode DM
                if (this.activeTab === 'dm' && this.otherUserId) {
                    await this.loadOtherUserInfo();
                }
            } catch (error) {
                console.error('Erreur chargement utilisateur:', error);
            }
        },

        async loadOtherUserInfo() {
            if (!this.otherUserId) return;
            
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch(`/users/${this.otherUserId}`, {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                
                if (response.ok) {
                    const userData = await response.json();
                    this.otherUserInfo.username = userData.username;
                    this.otherUserInfo.isOnline = userData.isOnline || false;
                } else {
                    this.otherUserInfo.username = `Utilisateur #${this.otherUserId}`;
                }
            } catch (error) {
                this.otherUserInfo.username = `Utilisateur #${this.otherUserId}`;
            }
        },

        // Changement d'onglet
        async switchTab(newTab) {
            // Nettoyer l'état précédent
            this.messages = [];
            this.loadingMessages = false;
            this.messageContent = '';
            
            if (newTab === 'rooms') {
                this.currentRoom = '';
                this.connectedUsers = [];
                this.otherUserId = null;
                this.otherUserInfo = { username: '', isOnline: false };
            } else if (newTab === 'dm') {
                this.currentRoom = '';
                this.connectedUsers = [];
                if (this.allUsers.length === 0) {
                    await this.loadUsers();
                }
            }
            
            // Mettre à jour l'URL
            const url = new URL(window.location);
            url.searchParams.set('tab', newTab);
            window.history.pushState({}, '', url);
        },

        // ===== FONCTIONS SALONS =====
        
        async loadRooms() {
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch('/chat/rooms', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });

                if (response.ok) {
                    this.rooms = await response.json();
                    this.roomStats.totalRooms = this.rooms.length;
                }
            } catch (error) {
                this.showNotification('Erreur lors du chargement des salons', 'error');
            }
        },

        async refreshRooms() {
            await this.loadRooms();
            this.showNotification('Liste des salons actualisée', 'success');
        },

        async joinRoom(roomName) {
            if (this.currentRoom === roomName) return;
            
            this.currentRoom = roomName;
            this.messages = [];
            this.connectedUsers = [];
            this.loadingMessages = true;
            
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({ 
                    type: "join", 
                    room: roomName 
                }));
                
                setTimeout(() => {
                    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                        this.socket.send(JSON.stringify({ 
                            type: "room_history", 
                            room: roomName, 
                            limit: 50 
                        }));
                    }
                }, 100);
                
                setTimeout(() => {
                    if (this.loadingMessages) {
                        this.loadingMessages = false;
                    }
                }, 10000);
            }
            
            this.showNotification(`Salon "${roomName}" rejoint`, 'success');
        },

        leaveRoom() {
            if (!this.currentRoom) return;
            
            const roomName = this.currentRoom;
            this.currentRoom = '';
            this.messages = [];
            this.connectedUsers = [];
            
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({ 
                    type: "leave", 
                    room: roomName 
                }));
            }
            
            this.showNotification(`Salon "${roomName}" quitté`, 'info');
        },

        async createRoom() {
            if (!this.newRoomName.trim()) return;
            
            this.creating = true;
            
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch('/chat/rooms', {
                    method: 'POST',
                    headers: {
                        'Authorization': 'Bearer ' + token,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ 
                        name: this.newRoomName.trim(),
                        description: `Salon créé par ${this.username}`
                    })
                });
                
                if (response.ok) {
                    const newRoom = await response.json();
                    this.rooms.push(newRoom);
                    this.newRoomName = '';
                    this.showNotification('Salon créé avec succès', 'success');
                    await this.joinRoom(newRoom.name);
                } else {
                    throw new Error('Erreur lors de la création');
                }
            } catch (error) {
                this.showNotification('Erreur lors de la création du salon', 'error');
            } finally {
                this.creating = false;
            }
        },

        // ===== FONCTIONS MESSAGES PRIVÉS =====

        async loadUsers() {
            this.loadingUsers = true;
            try {
                const token = localStorage.getItem('access_token');
                const response = await fetch('/users/except-me', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                
                if (response.ok) {
                    this.allUsers = await response.json();
                    this.filteredUsers = [...this.allUsers];
                    this.loadRecentConversations();
                }
            } catch (error) {
                this.showNotification('Erreur lors du chargement des utilisateurs', 'error');
            } finally {
                this.loadingUsers = false;
            }
        },

        async refreshUsers() {
            await this.loadUsers();
            this.showNotification('Liste des utilisateurs actualisée', 'success');
        },

        searchUsers() {
            const query = this.userSearch.toLowerCase();
            if (!query) {
                this.filteredUsers = [...this.allUsers];
            } else {
                this.filteredUsers = this.allUsers.filter(user => 
                    user.username.toLowerCase().includes(query) ||
                    user.email.toLowerCase().includes(query)
                );
            }
        },

        async selectUser(userId) {
            if (this.otherUserId === userId) return;
            
            // Fermer la connexion WebSocket actuelle proprement
            if (this.socket) {
                this.socket.onclose = null;
                this.socket.onerror = null;
                this.socket.close();
                this.socket = null;
            }
            
            // Mettre à jour l'état
            this.otherUserId = userId;
            this.messages = [];
            this.messageContent = '';
            this.isConnected = false;
            this.loadingMessages = false;
            
            // Mettre à jour l'URL
            const url = new URL(window.location);
            url.searchParams.set('tab', 'dm');
            url.searchParams.set('user_id', userId);
            window.history.pushState({}, '', url);
            
            // Charger les infos du nouvel utilisateur
            await this.loadOtherUserInfo();
            
            // Se connecter au WebSocket
            await this.connectWebSocket();
            
            // Sauvegarder dans les conversations récentes
            this.addToRecentConversations(userId);
            
            this.showNotification(`Conversation avec ${this.otherUserInfo.username} ouverte`, 'success');
        },

        loadRecentConversations() {
            const recent = JSON.parse(localStorage.getItem('recent_conversations') || '[]');
            this.recentConversations = recent.slice(0, 5);
        },

        addToRecentConversations(userId) {
            const user = this.allUsers.find(u => u.id === userId);
            if (!user) return;
            
            let recent = JSON.parse(localStorage.getItem('recent_conversations') || '[]');
            recent = recent.filter(conv => conv.userId !== userId);
            
            recent.unshift({
                userId: userId,
                username: user.username,
                lastMessage: 'Conversation ouverte',
                timestamp: new Date().toISOString()
            });
            
            recent = recent.slice(0, 10);
            localStorage.setItem('recent_conversations', JSON.stringify(recent));
            this.loadRecentConversations();
        },

        clearChat() {
            if (confirm('Êtes-vous sûr de vouloir effacer cette conversation ?')) {
                this.messages = [];
                this.updateStats();
                this.showNotification('Conversation effacée', 'info');
            }
        },

        // ===== WEBSOCKET =====

        async connectWebSocket() {
            try {
                const token = localStorage.getItem('access_token');
                
                if (this.socket) {
                    this.socket.onclose = null;
                    this.socket.onerror = null;
                    this.socket.close();
                    this.socket = null;
                }
                
                this.socket = new WebSocket(`ws://localhost:9001/?token=${token}`);
                
                this.socket.onopen = () => {
                    this.isConnected = true;
                    this.showNotification('Connexion établie', 'success');
                    
                    // Si on est en mode DM avec un utilisateur sélectionné
                    if (this.activeTab === 'dm' && this.otherUserId) {
                        this.loadingMessages = true;
                        this.socket.send(JSON.stringify({
                            type: "dm_history",
                            with: this.otherUserId,
                            limit: 50
                        }));
                        
                        setTimeout(() => {
                            if (this.loadingMessages) {
                                this.loadingMessages = false;
                            }
                        }, 10000);
                    }
                };
                
                this.socket.onclose = (event) => {
                    this.isConnected = false;
                    
                    if (event.code !== 1000) {
                        setTimeout(() => {
                            if ((this.activeTab === 'rooms') || (this.activeTab === 'dm' && this.otherUserId)) {
                                this.connectWebSocket();
                            }
                        }, 5000);
                    }
                };
                
                this.socket.onerror = (error) => {
                    this.showNotification('Erreur de connexion', 'error');
                };
                
                this.socket.onmessage = (event) => {
                    this.handleWebSocketMessage(JSON.parse(event.data));
                };
            } catch (error) {
                this.showNotification('Impossible de se connecter au serveur', 'error');
            }
        },

        handleWebSocketMessage(data) {
            console.log("📥 WS reçu :", data);
            
            if (this.activeTab === 'rooms') {
                this.handleRoomMessage(data);
            } else if (this.activeTab === 'dm') {
                this.handleDMMessage(data);
            }
        },

        handleRoomMessage(data) {
            if (data.type === "message" && data.data?.room === this.currentRoom) {
                this.messages.push(data.data);
                this.scrollToBottom();
                
                if (this.notificationsEnabled && document.hidden) {
                    this.showBrowserNotification(data.data.username, data.data.content);
                }
            } else if (data.username && data.content) {
                if (data.room === this.currentRoom || !data.room) {
                    this.messages.push(data);
                    this.scrollToBottom();
                }
            } else if (Array.isArray(data)) {
                this.messages = data
                    .filter(m => m.room === this.currentRoom)
                    .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
                this.loadingMessages = false;
                this.scrollToBottom();
            } else if (data.type === "room_history") {
                if (data.messages && Array.isArray(data.messages)) {
                    this.messages = data.messages
                        .filter(m => m.room === this.currentRoom)
                        .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
                }
                this.loadingMessages = false;
                this.scrollToBottom();
            } else if (data.type === "user_joined") {
                if (data.user && !this.connectedUsers.find(u => u.id === data.user.id)) {
                    this.connectedUsers.push(data.user);
                }
            } else if (data.type === "user_left") {
                this.connectedUsers = this.connectedUsers.filter(u => u.id !== data.user.id);
            } else if (data.type === "typing") {
                this.handleTypingIndicator(data);
            } else if (data.type === "room_users") {
                this.connectedUsers = data.users || [];
                this.roomStats.activeUsers = this.connectedUsers.length;
            }
        },

        handleDMMessage(data) {
            if (data.type === "dm" && (data.data?.fromUser === this.otherUserId || data.data?.to === this.otherUserId)) {
                this.messages.push({
                    ...data.data,
                    id: Date.now(),
                    status: 'delivered'
                });
                this.scrollToBottom();
                
                if (this.notificationsEnabled && document.hidden) {
                    this.showBrowserNotification(this.otherUserInfo.username, data.data.content);
                }
            } else if (data.type === "dm_history") {
                if (data.data && Array.isArray(data.data)) {
                    this.messages = data.data
                        .filter(msg => msg.content)
                        .map(msg => ({
                            ...msg,
                            id: msg.id || Date.now() + Math.random(),
                            status: 'delivered'
                        }))
                        .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
                } else {
                    this.messages = [];
                }
                this.loadingMessages = false;
                this.updateStats();
                this.scrollToBottom();
            } else if (data.type === "dm_sent") {
                const lastMessage = this.messages[this.messages.length - 1];
                if (lastMessage && lastMessage.fromUser === this.myUserId) {
                    lastMessage.status = 'sent';
                }
            } else if (data.type === "user_status") {
                if (data.userId === this.otherUserId) {
                    this.otherUserInfo.isOnline = data.isOnline;
                }
            } else if (data.type === "typing" && data.fromUser === this.otherUserId) {
                // Gérer l'indicateur de frappe pour les DM
            }
        },

        // ===== ENVOI DE MESSAGES =====

        async sendMessage() {
            const content = this.messageContent.trim();
            if (!content || this.sending || !this.isConnected) return;
            
            this.sending = true;
            
            try {
                if (this.activeTab === 'rooms' && this.currentRoom) {
                    await this.sendRoomMessage(content);
                } else if (this.activeTab === 'dm' && this.otherUserId) {
                    await this.sendDMMessage(content);
                }
            } finally {
                this.sending = false;
            }
        },

        async sendRoomMessage(content) {
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({
                    type: "message",
                    room: this.currentRoom,
                    content: content
                }));
                
                this.messageContent = '';
                this.roomStats.todayMessages++;
            }
        },

        async sendDMMessage(content) {
            const outgoingMessage = {
                id: Date.now(),
                fromUser: this.myUserId,
                to: this.otherUserId,
                content: content,
                timestamp: new Date().toISOString(),
                username: this.username,
                status: 'sending'
            };
            
            if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                this.socket.send(JSON.stringify({
                    type: "dm",
                    to: this.otherUserId,
                    content: content
                }));
                
                this.messages.push(outgoingMessage);
                this.messageContent = '';
                this.updateStats();
                this.scrollToBottom();
                
                setTimeout(() => {
                    outgoingMessage.status = 'sent';
                }, 1000);
            }
        },

        // ===== UTILITAIRES =====

        handleTyping() {
            if (!this.socket || !this.socket.readyState === WebSocket.OPEN) return;
            
            if (this.activeTab === 'rooms' && this.currentRoom) {
                this.socket.send(JSON.stringify({
                    type: "typing",
                    room: this.currentRoom,
                    isTyping: true
                }));
                
                if (this.typingTimeout) clearTimeout(this.typingTimeout);
                
                this.typingTimeout = setTimeout(() => {
                    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                        this.socket.send(JSON.stringify({
                            type: "typing",
                            room: this.currentRoom,
                            isTyping: false
                        }));
                    }
                }, 3000);
            } else if (this.activeTab === 'dm' && this.otherUserId) {
                this.socket.send(JSON.stringify({
                    type: "typing",
                    to: this.otherUserId,
                    isTyping: true
                }));
                
                if (this.typingTimeout) clearTimeout(this.typingTimeout);
                
                this.typingTimeout = setTimeout(() => {
                    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
                        this.socket.send(JSON.stringify({
                            type: "typing",
                            to: this.otherUserId,
                            isTyping: false
                        }));
                    }
                }, 3000);
            }
        },

        handleTypingIndicator(data) {
            if (data.username === this.username) return;
            
            if (data.isTyping) {
                if (!this.isTyping.includes(data.username)) {
                    this.isTyping.push(data.username);
                }
            } else {
                this.isTyping = this.isTyping.filter(u => u !== data.username);
            }
        },

        scrollToBottom() {
            this.$nextTick(() => {
                const container = this.$refs.messagesContainer;
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            });
        },

        handleScroll() {
            // Future feature: charger plus de messages
        },

        updateStats() {
            if (this.activeTab === 'dm') {
                this.dmStats.totalMessages = this.messages.length;
                
                const today = new Date().toDateString();
                this.dmStats.todayMessages = this.messages.filter(msg => 
                    new Date(msg.timestamp).toDateString() === today
                ).length;
            }
        },

        formatTime(timestamp) {
            const date = new Date(timestamp);
            const now = new Date();
            const diffInHours = (now - date) / (1000 * 60 * 60);
            
            if (diffInHours < 24) {
                return date.toLocaleTimeString('fr-FR', { 
                    hour: '2-digit', 
                    minute: '2-digit' 
                });
            } else {
                return date.toLocaleDateString('fr-FR', { 
                    day: 'numeric', 
                    month: 'short',
                    hour: '2-digit', 
                    minute: '2-digit'
                });
            }
        },

        // ===== NOTIFICATIONS =====

        async requestNotificationPermission() {
            if ('Notification' in window && Notification.permission === 'default') {
                await Notification.requestPermission();
            }
        },

        showBrowserNotification(username, content) {
            if ('Notification' in window && Notification.permission === 'granted') {
                const title = this.activeTab === 'rooms' ? `${username} dans ${this.currentRoom}` : `Message de ${username}`;
                new Notification(title, {
                    body: content,
                    icon: '/favicon.ico'
                });
            }
        },

        toggleNotifications() {
            this.notificationsEnabled = !this.notificationsEnabled;
            this.showNotification(
                this.notificationsEnabled ? 'Notifications activées' : 'Notifications désactivées',
                'info'
            );
        },

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
        }
    }
}