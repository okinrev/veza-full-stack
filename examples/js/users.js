// file: frontend/js/users.js

function userList() {
    return {
        users: [],
        async init() {
            const token = localStorage.getItem("access_token");
            if (!token) return alert("Non connecté");
  
            const res = await fetch("/users/except-me", {
                headers: { Authorization: "Bearer " + token }
            });
  
            if (res.ok) {
                this.users = await res.json();
            } else {
                alert("Erreur lors du chargement des utilisateurs");
            }
        }
    }
}