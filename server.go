package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	data "forum/Data"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID       int
	Email    string
	Username string
	Password string
	DB       *sql.DB
}

var (
	db       *sql.DB
	sessions = map[string]string{} // map to store session IDs and corresponding usernames
)

func main() {
	var err error
	db, err = data.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &mainPageHandler{})
	http.Handle("/register", &registerHandler{})
	http.Handle("/login", &loginHandler{})
	http.Handle("/protected", &protectedHandler{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./src/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

	fmt.Println("Serveur écoutant sur le port 6969...")
	log.Fatal(http.ListenAndServe("localhost:6969", nil))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
	}
}

type mainPageHandler struct{}

func (h *mainPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/Main_page.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type registerHandler struct{}

func (h *registerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/register.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")
		if email == "" || username == "" || password == "" {
			http.Error(w, "Email, nom d'utilisateur ou mot de passe vide", http.StatusBadRequest)
			return
		}
		_, err := db.Exec("INSERT INTO utilisateurs (email, username, password) VALUES (?, ?, ?)", email, username, password)
		if err != nil {
			http.Error(w, "Erreur lors de l'inscription", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la base de données:", err)
			return
		}
		// fmt.Fprintln(w, "Inscription réussie")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type loginHandler struct{}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/login.html", nil)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")
		if email == "" || password == "" {
			http.Error(w, "Nom d'utilisateur ou mot de passe vide", http.StatusBadRequest)
			return
		}
		var dbUsername, dbPassword string
		err := db.QueryRow("SELECT email, password FROM utilisateurs WHERE email = ?", email).Scan(&dbUsername, &dbPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Erreur lors de la vérification de l'utilisateur", http.StatusInternalServerError)
			log.Println("Erreur lors de la vérification de l'utilisateur:", err)
			return
		}
		if password != dbPassword {
			http.Error(w, "Nom d'utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
			return
		}
		// Créer une session
		sessionID := uuid.New().String()
		sessions[sessionID] = email
		cookie := &http.Cookie{
			Name:  "session_id",
			Value: sessionID,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/protected", http.StatusSeeOther)
		return
	}
	http.NotFound(w, r)
}

type protectedHandler struct{}

func (h *protectedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Vérifier la session
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	sessionID := cookie.Value
	username, ok := sessions[sessionID]
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	fmt.Printf("User %s is accessing the protected page\n", username)
	if r.Method == http.MethodGet {
		renderTemplate(w, "./src/protected.html", nil)
		return
	}
	http.NotFound(w, r)
}