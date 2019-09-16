package main

import (
	"context"
	"net/http"
	"text/template"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User struct for each user
type User struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

//user data
type UserData struct {
	userID primitive.ObjectID
}

func signUp(w http.ResponseWriter, r *http.Request) {
	var newUserData UserData
	t, _ := template.ParseFiles("static/html/index.html")
	t.Execute(w, nil)
	var user User
	if r.Body == nil {
		http.Error(w, "Send a body", 400)
		return
	}
	user.FirstName = r.FormValue("firstName")
	user.LastName = r.FormValue("lastName")
	user.Email = r.FormValue("email")
	user.Password = r.FormValue("password")
	collectionUsers := getCollection("users")
	collectionData := getCollection("userData")
	collectionUsers.InsertOne(context.TODO(), user)
	collectionUsers.FindOne(context.TODO(), bson.D{{"email", user.Email}}).Decode(&newUserData)
	collectionData.InsertOne(context.TODO(), newUserData)
}

func checkLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	collectionUsers := getCollection("users")
	collectionUsers.FindOne(context.TODO(), bson.D{{"email", r.FormValue("email")}}).Decode(&user)
	if user.Password == r.FormValue("password") {
		t, _ := template.ParseFiles("static/html/user-page.html")
		t.Execute(w, nil)
	} else {
		http.Error(w, "incorrect password", 400)
	}
}
