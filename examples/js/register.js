// file: frontend/js/register.js

function registerForm() {
    return {
      email: '',
      username: '',
      password: '',
      confirm: '',
      message: '',

      submit() {
        if (this.password !== this.confirm) {
          this.message = 'Les mots de passe ne correspondent pas';
          return;
        }

        fetch('/signup', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email: this.email, username: this.username, password: this.password })
        })
        .then(res => res.ok ? res.json() : Promise.reject(res))
        .then(data => {
          localStorage.setItem('access_token', data.access_token);
          localStorage.setItem('refresh_token', data.refresh_token);
          this.message = '✅ Inscription réussie !';
          window.location.href = '/dashboard.html';
        })
        .catch(() => this.message = '❌ Erreur lors de l\'inscription');
      }
    };
  }