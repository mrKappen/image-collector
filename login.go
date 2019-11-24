package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	bcrypt "golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
)

func logout(w http.ResponseWriter, r *http.Request) {
	log.Println("IN logout")
	session, err := store.Get(r, "auth-cookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("failed to logout")
		return
	}

	session.Values["authenticated"] = false
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("error: " + err.Error())
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
func signUp(w http.ResponseWriter, r *http.Request) {
	//TODO: ensure created user is unique
	var user User
	var createdUser User
	_ = json.NewDecoder(r.Body).Decode(&user)
	uniqueEmail := user.Email
	usersCollection := getCollection("users")
	userDataCollection := getCollection("userData")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		log.Println("ERROR hashing password: " + err.Error())
	}
	user.Password = string(hashedPassword)
	usersCollection.InsertOne(context.TODO(), user)
	filter := bson.D{primitive.E{"Email", uniqueEmail}}
	err = usersCollection.FindOne(context.TODO(), filter).Decode(&createdUser)
	userDataID := createdUser.UserID.Hex()
	_, err = userDataCollection.InsertOne(context.TODO(), UserData{UserID: userDataID})
	if err != nil {
		http.Error(w, "failed to register user", 400)
		return
	}
	returnData := make(map[string]string)
	returnData["userID"] = userDataID
	v, err := json.Marshal(returnData)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	session, err := store.Get(r, "auth-cookie")
	if err != nil {
		log.Println("Failed to set auth cookie: " + err.Error())
	}
	session.Values["authenticated"] = true
	session.Values["userID"] = userDataID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.Write(v)
}

func checkLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	var sentUser User
	collectionUsers := getCollection("users")
	returnData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&sentUser)
	fmt.Println(sentUser)
	collectionUsers.FindOne(context.TODO(), bson.D{{"Email", sentUser.Email}}).Decode(&user)
	fmt.Println(user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(sentUser.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		session, err := store.Get(r, "auth-cookie")
		if err != nil {
			log.Println("Failed to set auth cookie: " + err.Error())
		}
		session.Values["authenticated"] = true
		session.Values["userID"] = user.UserID.Hex()
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-type", "application/json")
		returnData["userID"] = user.UserID.Hex()
		v, _ := json.Marshal(returnData)
		w.Write(v)
	}
}
