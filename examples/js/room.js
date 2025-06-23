// file: frontend/js/room.js

document.addEventListener('alpine:init', () => {
    Alpine.store('roomChat', {
      rooms: [],
      roomActuelle: '',
      messages: [],
      contenu: '',
      socket: null,
      logEl: null,

      async init() {
        const token = localStorage.getItem('access_token');
        if (!token) return alert("Non connecté");

        const res = await fetch('/chat/rooms', {
          headers: { Authorization: "Bearer " + token }
        });
        this.rooms = await res.json();

        this.socket = new WebSocket(`ws://localhost:9001/?token=${token}`);
        this.socket.onmessage = (event) => {
          const data = JSON.parse(event.data);
          //console.log("📥 WS reçu :", data);

          if (data.type === "message" && data.data?.room === this.roomActuelle) {
            this.messages.push(data.data);
            this.scrollToBottom();
          } else if (data.username && data.content) {
            this.messages.push(data);
            this.scrollToBottom();
          } else if (Array.isArray(data)) {
            this.messages = data
              .filter(m => m.room === this.roomActuelle)
              .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
            this.scrollToBottom();
          }
        };
      },

      scrollToBottom() {
        setTimeout(() => {
          if (this.logEl) {
            this.logEl.scrollTop = this.logEl.scrollHeight;
          }
        }, 0);
      },

      rejoindre(room) {
        this.roomActuelle = room;
        this.messages = [];
        this.socket.send(JSON.stringify({ type: "join", room }));
        this.socket.send(JSON.stringify({ type: "room_history", room, limit: 50 }));
      },

      envoyerMessage() {
        const content = this.contenu.trim();
        if (!content || !this.roomActuelle) return;

        this.socket.send(JSON.stringify({
          type: "message",
          room: this.roomActuelle,
          content
        }));
        this.contenu = '';
      }
    });

    Alpine.effect(() => {
      const el = document.querySelector('[x-ref="log"]');
      if (el) Alpine.store('roomChat').logEl = el;
    });

    Alpine.store('roomChat').init();
});