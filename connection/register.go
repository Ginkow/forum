package forum

import (
	"database/sql"
	// "fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier que la méthode est POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer les données du formulaire
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Vérifier si l'email est déjà pris
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already exists", http.StatusBadRequest)
		return
	}

	// Encrypter le mot de passe (à remplacer par un algorithme sécurisé)
	// Note : il est recommandé d'utiliser bcrypt ou un autre algorithme de hachage sécurisé
	encryptedPassword := password // Placeholder for encryption

	// Insérer l'utilisateur dans la base de données
	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, encryptedPassword)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Créer un cookie de session avec une date d'expiration
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: "session", Value: email, Expires: expiration}
	http.SetCookie(w, &cookie)

	// Rediriger vers une page de succès ou envoyer une réponse de succès
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}
