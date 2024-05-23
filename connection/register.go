package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// Utilisateur structure pour stocker les informations d'utilisateur
type Users struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler structure pour la page d'inscription
type RegisterHandler struct {
	DB *sql.DB
}

// ServeHTTP méthode pour gérer les requêtes d'inscription
func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var creds Users
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Requête mal formée", http.StatusBadRequest)
		return
	}

	// Insertion de l'utilisateur dans la base de données
	stmt, err := h.DB.Prepare("INSERT INTO utilisateurs(username, password) VALUES(?, ?)")
	if err != nil {
		http.Error(w, "Erreur de préparation de la requête", http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(creds.Username, creds.Password)
	fmt.Println(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, "Erreur lors de l'insertion de l'utilisateur", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Utilisateur créé avec succès")
}
