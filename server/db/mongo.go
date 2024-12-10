package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

// InitializeMongo connects to MongoDB and initializes the global client
func InitializeMongo(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database to ensure connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	log.Println("Connected to MongoDB")
	MongoClient = client
}

// GetCollection returns a MongoDB collection for a given database and collection name
func GetCollection(database, collection string) *mongo.Collection {
	if MongoClient == nil {
		log.Fatal("MongoClient is not initialized")
	}
	return MongoClient.Database(database).Collection(collection)
}
