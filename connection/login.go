package forum

import (
    // "database/sql"
    "net/http"
	"time"

    _ "github.com/go-sql-driver/mysql"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Vérifier que la méthode est POST
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Récupérer les données du formulaire
    email := r.FormValue("email")
    password := r.FormValue("password")

    // Vérifier si l'email existe dans la base de données
    var storedPassword string
    err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&storedPassword)
    if err != nil {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Comparer les mots de passe (à remplacer par une vérification sécurisée)
    if storedPassword != password {
        http.Error(w, "Invalid email or password", http.StatusUnauthorized)
        return
    }

    // Créer un cookie de session avec une date d'expiration
    expiration := time.Now().Add(24 * time.Hour)
    cookie := http.Cookie{Name: "session", Value: email, Expires: expiration}
    http.SetCookie(w, &cookie)

    // Rediriger vers une page de succès ou envoyer une réponse de succès
    http.Redirect(w, r, "/success", http.StatusSeeOther)
}
