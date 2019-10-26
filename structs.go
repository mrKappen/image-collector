package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User struct for each user
type User struct {
	UserID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstName" bson:"FirstName"`
	LastName  string             `json:"lastName" bson:"LastName"`
	Email     string             `json:"email" bson:"Email"`
	Password  string             `json:"password" bson:"Password"`
}

//UserData stores user data
type UserData struct {
	UserID      string       `json:"userId" bson:"UserID"`
	Collections []Collection `json:"collections" bson:"Collections"`
}

//ImageObj object stored in the images collection
type ImageObj struct {
	Image        interface{} `json:"images" bson:"Image"`
	CollectionID string      `json:"collectionID" bson:"CollectionID"`
}

//Collection object
type Collection struct {
	CollectionID   string `json:"collectionID" bson:"CollectionID"`
	CollectionName string `json:"collectionName" bson:"CollectionName"`
}
