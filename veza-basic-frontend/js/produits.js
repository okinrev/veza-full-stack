// file: frontend/js/produits.js

function produitsApp() {
    return {
      produits: [],
      message: '',
      form: {
        id: null,
        name: '',
        version: '',
        purchase_date: '',
        warranty_expires: ''
      },

      async getValidToken() {
        return localStorage.getItem('access_token') || null
      },

      async chargerProduits() {
        const token = await this.getValidToken()
        if (!token) return this.message = "Non connecté"

        try {
          const r = await fetch('/products', {
            headers: { 'Authorization': 'Bearer ' + token }
          })
          if (!r.ok) throw new Error("Erreur " + r.status)
          this.produits = await r.json()
        } catch (e) {
          this.message = e.message
        }
      },

      formatDate(str) {
        const d = new Date(str)
        return isNaN(d) ? '—' : d.toLocaleDateString('fr-FR')
      },

      remplirFormulaire(p) {
        this.form = { ...p }
        this.form.purchase_date = p.purchase_date.split('T')[0]
        this.form.warranty_expires = p.warranty_expires.split('T')[0]
      },

      resetForm() {
        this.form = {
          id: null,
          name: '',
          version: '',
          purchase_date: '',
          warranty_expires: ''
        }
      },

      async enregistrerProduit() {
        const token = await this.getValidToken()
        if (!token) return this.message = "Non connecté"

        const method = this.form.id ? 'PUT' : 'POST'
        const url = this.form.id ? `/products/${this.form.id}` : '/products'

        const payload = {
          name: this.form.name,
          version: this.form.version,
          purchase_date: new Date(this.form.purchase_date).toISOString(),
          warranty_expires: new Date(this.form.warranty_expires).toISOString()
        }

        try {
          const r = await fetch(url, {
            method,
            headers: {
              'Authorization': 'Bearer ' + token,
              'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
          })

          if (!r.ok) throw new Error("Erreur " + r.status)
          this.message = this.form.id ? "✅ Produit mis à jour" : "✅ Produit créé"
          this.resetForm()
          this.chargerProduits()
        } catch (e) {
          this.message = e.message
        }
      },

      async supprimerProduit(id) {
        const token = await this.getValidToken()
        if (!token) return this.message = "Non connecté"

        try {
          const r = await fetch(`/products/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': 'Bearer ' + token }
          })
          if (!r.ok) throw new Error("Erreur " + r.status)
          this.message = "🗑️ Produit supprimé"
          this.chargerProduits()
        } catch (e) {
          this.message = e.message
        }
      }
    }
  }