// file: frontend/js/users.js

function dmChat() {
    return {
      myUserId: null,
      otherUserId: parseInt(new URLSearchParams(window.location.search).get("user_id")),
      messages: [],
      contenu: '',
      socket: null,

      async init() {
        const token = localStorage.getItem('access_token');
        if (!token || isNaN(this.otherUserId)) return alert("Token manquant ou ID invalide");

        // Récupération de l'utilisateur courant
        try {
          const res = await fetch("/me", {
            headers: { Authorization: "Bearer " + token }
          });
          const data = await res.json();
          this.myUserId = data.id;
        } catch (e) {
          return alert("Impossible de charger les infos utilisateur");
        }

        // Initialisation WebSocket
        if (this.socket && this.socket.readyState === WebSocket.OPEN) this.socket.close();
        this.socket = new WebSocket(`ws://localhost:9001/?token=${token}`);

        this.socket.onopen = () => {
          this.socket.send(JSON.stringify({
            type: "dm_history",
            with: this.otherUserId,
            limit: 50
          }));
        };

        this.socket.onmessage = (event) => {
          const data = JSON.parse(event.data);

          if (data.type === "dm" && (data.data?.fromUser === this.otherUserId || data.data?.fromUser === this.myUserId)) {
            this.messages.push(data.data);
            this.scrollToBottom();
          } else if (Array.isArray(data)) {
            this.messages = data
              .filter(msg => msg.content)
              .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
            this.scrollToBottom();
          }
        };

        this.socket.onerror = (err) => {
          console.error("WebSocket error:", err);
        };
      },

      envoyerMessage() {
        const msg = this.contenu.trim();
        if (!msg || !this.socket || this.socket.readyState !== WebSocket.OPEN) return;

        const payload = {
          type: "dm",
          to: this.otherUserId,
          content: msg
        };

        // Envoi via WebSocket
        this.socket.send(JSON.stringify(payload));

        // Ajout local immédiat
        this.messages.push({
          fromUser: this.myUserId,
          to: this.otherUserId,
          content: msg,
          timestamp: new Date().toISOString(),
          username: 'Moi'
        });

        this.scrollToBottom();
        this.contenu = '';
      },

      scrollToBottom() {
        this.$nextTick(() => {
          const el = this.$refs.log;
          if (el) el.scrollTop = el.scrollHeight;
        });
      }
    };
  }