package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "html/template"
    "log"
    //"github.com/gorilla/mux"
)

type Page struct {
    Title string
    Body  []byte
}
func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    fmt.Println(filename)
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    t := template.New("index.html")
    t, err := t.ParseFiles("template/index.html")
    if err != nil {
        fmt.Println("view parser error")
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, nil)
    if err != nil {
        fmt.Println("Execute parser error")
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func httpService() {
    http.Handle("/resources/",http.StripPrefix("/resources/", http.FileServer(http.Dir("resources")))) 
    http.Handle("/libs/",http.StripPrefix("/libs/", http.FileServer(http.Dir("resources/libs")))) 
    http.Handle("/",http.StripPrefix("/", http.FileServer(http.Dir("resources/")))) 
    http.Handle("/controller/",http.StripPrefix("/controller/", http.FileServer(http.Dir("resources/controller")))) 
    http.Handle("/service/",http.StripPrefix("/service/", http.FileServer(http.Dir("resources/service")))) 
    http.Handle("/jsonlog/",http.StripPrefix("/jsonlog/", http.FileServer(http.Dir("jsonlog")))) 
    http.HandleFunc("/monitor", viewHandler)
    log.Println("Listening...")
    http.ListenAndServe(":8080",nil)
}
/*
func main() {
    go httpService()
    log.Println("Listening...")
    http.ListenAndServe(":8080", nil)
    //r := mux.NewRouter()
    //r.HandleFunc("/", viewHandler)
    //http.Handle("/template/",http.StripPrefix("/template/", http.FileServer(http.Dir("template")))) 
    //http.Handle("/jsonlog/",http.StripPrefix("/jsonlog/", http.FileServer(http.Dir("jsonlog")))) 
    //http.Handle("/", r)
}
*/