package main

import (
    "fmt"
    "log"
    "net/http"
    "text/template"
)

func main() {
    // Gestion des routes
    http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./"))))

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
        log.Fatal(err)
    }

    err = t.Execute(w, data)
    if err != nil {
        log.Fatal(err)
    }
}
