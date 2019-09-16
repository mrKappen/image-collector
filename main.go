package main

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var db *mongo.Client

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/html/index.html")
	t.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/html/login.html")
	t.Execute(w, nil)
}
func register(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/html/register.html")
	t.Execute(w, nil)
}
func main() {
	db = setUpDb()
	r := mux.NewRouter()
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/register", register).Methods("GET")
	r.HandleFunc("/register", signUp).Methods("POST")
	r.HandleFunc("/checkLogin", checkLogin).Methods("POST")
	fmt.Println("**************STARTING THE SERVER**************")
	http.ListenAndServe(":8080", r)
}
