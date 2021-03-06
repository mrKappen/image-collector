package main

import (
	"context"
	"errors"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const database string = "imageCollector"

func setUpDb() (*mongo.Client, error) {
	// Set client options
	DB_URI := os.Getenv("DB_STRING")
	log.Println(DB_URI)
	if DB_URI == "" {
		log.Println("Connection string not found!")
		return nil, errors.New("connection string not found")
	}
	clientOptions := options.Client().ApplyURI(DB_URI)
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
