package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
)

func signUp(w http.ResponseWriter, r *http.Request) {
	//TODO: ensure created user is unique
	var user User
	var createdUser User
	_ = json.NewDecoder(r.Body).Decode(&user)
	uniqueEmail := user.Email
	usersCollection := getCollection("users")
	userDataCollection := getCollection("userData")
	usersCollection.InsertOne(context.TODO(), user)
	filter := bson.D{primitive.E{"Email", uniqueEmail}}
	err := usersCollection.FindOne(context.TODO(), filter).Decode(&createdUser)
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
	if user.Password == sentUser.Password {
		fmt.Println("here!")
		w.Header().Add("Content-type", "application/json")
		returnData["userID"] = user.UserID.Hex()
		v, _ := json.Marshal(returnData)
		w.Write(v)
	} else {
		http.Error(w, "incorrect password", 400)
	}
}
