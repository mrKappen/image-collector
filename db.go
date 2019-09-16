package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database string = "imageCollector"

func setUpDb() *mongo.Client {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	db, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = db.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func getCollection(collectionName string) *mongo.Collection {
	collection := db.Database(database).Collection(collectionName)
	return collection
}

func filterDb(collectionName, filterKey string) {

}
