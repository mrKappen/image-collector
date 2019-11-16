package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database string = "imageCollector"
const URI string = "mongodb+srv://tkappen:Jesus999*@cluster0-jcnwn.mongodb.net/test"

func setUpDb() (*mongo.Client, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(URI)
	// Connect to MongoDB
	db, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// Check the connection
	err = db.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}

func getCollection(collectionName string) *mongo.Collection {
	collection := db.Database(database).Collection(collectionName)
	return collection
}

func filterDb(collectionName, filterKey string) {

}
