package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "text/template"
)

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("mysql", "user:password@/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Gestion des routes
    http.HandleFunc("/register", RegisterHandler)
    http.HandleFunc("/login", LoginHandler)

    // Gestion des fichiers statiques
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
    http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

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


func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        RenderTemplate(w, "/src/register.html", nil)
        return
    }

    // Processus d'inscription
    // Appelez la fonction RegisterHandler de votre package forum
    RegisterHandler(w, r)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        RenderTemplate(w, "/src/login.html", nil)
        return
    }

    // Processus de connexion
    // Appelez la fonction LoginHandler de votre package forum
    LoginHandler(w, r)
}
