package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "path/filepath"
    "text/template"
	"os"
	"regexp"

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

type Post struct {
    ID      int
    Title   string
    Content string
    Video   string
    UserID  int
}

type MainPageData struct {
    IsLoggedIn     bool
    ProfilePicture string
    Posts          []Post
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
    http.Handle("/newpost", &newPostHandler{})
    http.Handle("/posts", &postsHandler{})
    http.Handle("/erreur", &errorHandler{})

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
    http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./src/"))))
    http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))

    fmt.Println("Serveur écoutant sur le port 6969...")
    log.Fatal(http.ListenAndServe("localhost:6969", nil))
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
    t, err := template.ParseFiles(tmpl)
    if err != nil {
        log.Println(err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    if err := t.Execute(w, data); err != nil {
        log.Println(err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}



type mainPageHandler struct{}

func (h *mainPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        var data MainPageData
        sessionCookie, err := r.Cookie("session_id")
        if err == nil {
            email, ok := sessions[sessionCookie.Value]
            if ok {
                data.IsLoggedIn = true
                // Récupérer l'image de profil de l'utilisateur
                var profilePicture string
                err := db.QueryRow("SELECT profile_picture FROM utilisateurs WHERE email = ?", email).Scan(&profilePicture)
                if err == nil {
                    data.ProfilePicture = profilePicture
                }
            }
        }

        // Récupérer les posts
        rows, err := db.Query("SELECT title, content, video FROM posts")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des posts:", err)
            return
        }
        defer rows.Close()

        for rows.Next() {
            var post Post
            err := rows.Scan(&post.Title, &post.Content, &post.Video)
            if err != nil {
                http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
                log.Println("Erreur lors de la lecture des posts:", err)
                return
            }
            data.Posts = append(data.Posts, post)
        }

        renderTemplate(w, "./src/Main_page.html", data)
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
            setCookie(w, "error", "Erreur lors de la lecture du formulaire")
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
        email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")
        if email == "" || username == "" || password == "" {
            setCookie(w, "error", "Email, nom d'utilisateur ou mot de passe vide")
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
        emailPattern := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
        matched, err := regexp.MatchString(emailPattern, email)
        if err != nil || !matched {
            setCookie(w, "error", "Email invalide")
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
        if usernameExists(username) {
            setCookie(w, "error", "Nom d'utilisateur deja pris, veuillez en choisir un autre")
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
		if emailExists(email) {
            setCookie(w, "error", "Email deja existente, veuillez en choisir un autre")
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
        _, err = db.Exec("INSERT INTO utilisateurs (email, username, password) VALUES (?, ?, ?)", email, username, password)
        if err != nil {
            setCookie(w, "error", "Erreur lors de l'inscription")
            log.Println("Erreur lors de l'insertion dans la base de données:", err)
            http.Redirect(w, r, "/register", http.StatusSeeOther)
            return
        }
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    http.NotFound(w, r)
}

func setCookie(w http.ResponseWriter, name, value string) {
    cookie := &http.Cookie{
        Name:  name,
        Value: value,
        Path:  "/",
        MaxAge: 10, // The cookie will be valid for 10 seconds
    }
    http.SetCookie(w, cookie)
}

func usernameExists(username string) bool {
    var exists bool
    query := "SELECT EXISTS (SELECT 1 FROM utilisateurs WHERE username = ?)"
    err := db.QueryRow(query, username).Scan(&exists)
    if err != nil {
        log.Println("Erreur lors de la vérification du nom d'utilisateur :", err)
        return true
    }
    return exists
}

func emailExists(email string) bool {
    var exists bool
    query := "SELECT EXISTS (SELECT 1 FROM utilisateurs WHERE email = ?)"
    err := db.QueryRow(query, email).Scan(&exists)
    if err != nil {
        log.Println("Erreur lors de la vérification du nom d'utilisateur :", err)
        return true
    }
    return exists
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
            setErrorCookie(w, "Email ou mot de passe vide")
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        var dbPassword string
        err := db.QueryRow("SELECT password FROM utilisateurs WHERE email = ?", email).Scan(&dbPassword)
        if err != nil {
            if err == sql.ErrNoRows {
                setErrorCookie(w, "Email ou mot de passe incorrect")
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }
            setErrorCookie(w, "Erreur lors de la vérification de l'utilisateur")
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            log.Println("Erreur lors de la vérification de l'utilisateur:", err)
            return
        }
        if password != dbPassword {
            setErrorCookie(w, "Mot de passe incorrect")
            http.Redirect(w, r, "/login", http.StatusSeeOther)
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
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    http.NotFound(w, r)
}

func setErrorCookie(w http.ResponseWriter, errorMsg string) {
    cookie := &http.Cookie{
        Name:  "error",
        Value: errorMsg,
        Path:  "/",
    }
    http.SetCookie(w, cookie)
}

type newPostHandler struct{}

func (h *newPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        renderTemplate(w, "./src/new_post.html", nil)
        return
    }
    if r.Method == http.MethodPost {
        // Check if user is logged in
        sessionCookie, err := r.Cookie("session_id")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        email, ok := sessions[sessionCookie.Value]
        if !ok {
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }
        var userID int
        err = db.QueryRow("SELECT id FROM utilisateurs WHERE email = ?", email).Scan(&userID)
        if err != nil {
            http.Error(w, "Erreur lors de la vérification de l'utilisateur", http.StatusInternalServerError)
            log.Println("Erreur lors de la vérification de l'utilisateur:", err)
            return
        }
        // Handle form submission
        if err := r.ParseMultipartForm(10 << 20); err != nil {
            http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
            return
        }
        title := r.FormValue("title")
        content := r.FormValue("content")
        var videoPath string

        // Handle video and image upload

		files, header, err := r.FormFile("all")
        if err == nil {
            defer files.Close()
            videoPath = filepath.Join("./img_video", header.Filename)
            out, err := os.Create(videoPath)
            if err != nil {
                http.Error(w, "Erreur lors de la sauvegarde de la vidéo", http.StatusInternalServerError)
                return
            }
            defer out.Close()
            _, err = files.Seek(0, 0)
            if err != nil {
                http.Error(w, "Erreur lors de la lecture du fichier", http.StatusInternalServerError)
                return
            }
            _, err = out.ReadFrom(files)
            if err != nil {
                http.Error(w, "Erreur lors de la lecture du fichier", http.StatusInternalServerError)
                return
            }
        }

        // Insert the post into the database
        result, err := db.Exec("INSERT INTO posts (title, content, video, user_id) VALUES (?, ?, ?, ?)", title, content, videoPath, userID)
		if err != nil {
			http.Error(w, "Erreur lors de la création du post", http.StatusInternalServerError)
			log.Println("Erreur lors de l'insertion dans la base de données:", err)
			return
		}

        // Get the ID of the inserted post
        postID, _ := result.LastInsertId()

        // Redirect to the main page with the ID of the new post
        http.Redirect(w, r, fmt.Sprintf("/?postID=%d", postID), http.StatusSeeOther)
        return
    }
    http.NotFound(w, r)
}

type postsHandler struct{}

func (h *postsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        rows, err := db.Query("SELECT title, content, video FROM posts")
        if err != nil {
            http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
            log.Println("Erreur lors de la récupération des posts:", err)
            return
        }
        defer rows.Close()
        var posts []Post
        for rows.Next() {
            var post Post
            err := rows.Scan(&post.Title, &post.Content, &post.Video)
            if err != nil {
                http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
                log.Println("Erreur lors de la lecture des posts:", err)
                return
            }
            posts = append(posts, post)
        }
        if err = rows.Err(); err != nil {
            http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
            log.Println("Erreur lors de la lecture des posts:", err)
            return
        }
        renderTemplate(w, "./src/posts.html", posts)
        return
    }
    http.NotFound(w, r)
}

type errorHandler struct{}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "./src/erreur.html", nil)
}
