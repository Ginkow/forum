package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "text/template"
    _ "github.com/go-sql-driver/mysql"
)


type Main_page struct{}
type Register struct{}
type Login struct{}

var db *sql.DB
func main() {
    var err error
    db, err = sql.Open("mysql", "user:password@/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Gestion des routes
    http.Handle("/", Main_page{})
    http.Handle("/register", Register{})
    http.Handle("/login", Login{})

    // Gestion des fichiers statiques
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
    http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("./src/"))))

    // Démarrer le serveur
    fmt.Println("Serveur écoutant sur le port 6969...")
    log.Fatal(http.ListenAndServe("localhost:6969", nil))
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
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

func (h Main_page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        http.ServeFile(w, r, "./src/Main_page.html")
        return
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h Register) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        http.ServeFile(w, r, "./src/register.html")
        return
    }

    saveAccount := r.FormValue("saveAccount")
    if saveAccount == "on" {
        fmt.Fprintf(w, "Votre compte a été sauvegardé pour une connexion plus facile la prochaine fois.")
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        http.ServeFile(w, r, "./src/login.html")
        return
    }
    http.Redirect(w, r, "/", http.StatusSeeOther)
}