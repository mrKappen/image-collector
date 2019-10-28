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
type ImageObjRetrieve struct {
	Image        []byte `json:"images" bson:"Image"`
	FileType     string `json:"type" bson:"Type"`
	Size         int64  `json:"size" bson:"Size"`
	CollectionID string `json:"collectionID" bson:"CollectionID"`
	FileId       string `json:"imageId" bson:"ImageId"`
}

type ImageObjSend struct {
	Image        interface{} `json:"images" bson:"Image"`
	FileType     string      `json:"type" bson:"Type"`
	Size         int64       `json:"size" bson:"Size"`
	CollectionID string      `json:"collectionID" bson:"CollectionID"`
	FileId       string      `json:"imageId" bson:"ImageId"`
}

//Collection object
type Collection struct {
	CollectionID   string   `json:"collectionID" bson:"CollectionID"`
	CollectionName string   `json:"collectionName" bson:"CollectionName"`
	FileIds        []string `json:"fileIds" bson:"FileIds"`
}
