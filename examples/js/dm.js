function logout() {
  localStorage.removeItem("access_token");
  window.location.href = "/login";
}

document.addEventListener('alpine:init', () => {
  Alpine.store('dmChat', {
    users: [],
    selectedUser: null,
    myUserId: null,
    messages: [],
    contenu: '',
    socket: null,
    logEl: null,

    async init() {
      const token = localStorage.getItem("access_token");
      if (!token) return window.location.href = "/login";

      const me = await fetch("/me", {
        headers: { Authorization: "Bearer " + token }
      });
      this.myUserId = (await me.json()).id;

      const res = await fetch("/users/except-me", {
        headers: { Authorization: "Bearer " + token }
      });
      this.users = await res.json();

      this.socket = new WebSocket(`ws://localhost:9001/?token=${token}`);
      this.socket.onmessage = (event) => {
        const data = JSON.parse(event.data);

        if (data.type === "dm" &&
          this.selectedUser &&
          (data.data.fromUser === this.selectedUser.id || data.data.to === this.selectedUser.id)) {
          this.messages.push(data.data);
          this.scrollToBottom();
        } else if (data.type === "dm_history") {
          this.messages = data.data
            .filter(msg => msg.content)
            .sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));
          this.scrollToBottom();
        }
      };
    },

    selectUser(user) {
      this.selectedUser = user;
      this.messages = [];
      if (this.socket && this.socket.readyState === WebSocket.OPEN) {
        this.socket.send(JSON.stringify({
          type: "dm_history",
          with: user.id,
          limit: 50
        }));
      }
    },

    envoyerMessage() {
      const content = this.contenu.trim();
      if (!content || !this.selectedUser || this.socket.readyState !== WebSocket.OPEN) return;

      this.socket.send(JSON.stringify({
        type: "dm",
        to: this.selectedUser.id,
        content
      }));

      this.messages.push({
        fromUser: this.myUserId,
        to: this.selectedUser.id,
        content,
        timestamp: new Date().toISOString(),
        username: "Moi"
      });

      this.scrollToBottom();
      this.contenu = '';
    },

    scrollToBottom() {
      setTimeout(() => {
        if (this.logEl) {
          this.logEl.scrollTop = this.logEl.scrollHeight;
        }
      }, 0);
    }
  });

  Alpine.effect(() => {
    const el = document.querySelector('[x-ref="log"]');
    if (el) Alpine.store('dmChat').logEl = el;
  });

  Alpine.store('dmChat').init();
});