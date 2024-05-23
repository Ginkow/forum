package forum

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// Utilisateur structure pour stocker les informations d'utilisateur
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Base de données simulée pour les utilisateurs (en mémoire)
var utilisateurs = map[string]string{
	"utilisateurs1": "utilisateurs1",
	"utilisateurs2": "utilisateurs2",
}

// LoginHandler structure pour la page de login
type LoginHandler struct{}
type ProtectedHandler struct{}


// ServeHTTP méthode pour gérer les requêtes de login
func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var creds User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Requête mal formée", http.StatusBadRequest)
		return
	}

	// Vérification des identifiants
	storedPassword, ok := utilisateurs[creds.Username]
	if !ok || storedPassword != creds.Password {
		http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Authentification réussie")
}

// ServeHTTP méthode pour gérer les requêtes de login
func (h ProtectedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	storedPassword, ok := utilisateurs[username]
	if !ok || storedPassword != password {
		http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Accès autorisé à la page protégée")
}
