package main

import (
	"context"
	"encoding/json"
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
	http.Error(w, "Success", 200)
}

func checkLogin(w http.ResponseWriter, r *http.Request) {
	// var user User
	// collectionUsers := getCollection("users")
	// collectionUsers.FindOne(context.TODO(), bson.D{{"email", r.FormValue("email")}}).Decode(&user)
	// if user.Password == r.FormValue("password") {
	// 	//Send get request for user data based on id
	// 	idStr := user.userID.String()
	// 	fmt.Println(user)
	// 	_, err := http.Get("/user/" + idStr)
	// 	if err != nil {
	// 		//handle error
	// 	}
	// } else {
	// 	http.Error(w, "incorrect password", 400)
	// }
}
