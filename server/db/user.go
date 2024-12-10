package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       int64  `bson:"ID"`
	UserName string `bson:"UserName"`
	PassWD   string `bson:"Passwd"`
	Email    string `bson:"Email"`
	Phone    int64  `bson:"Phone"`
}

// InsertUser inserts a new user into the database
func (user *User) InsertUser() error {
	collection := GetCollection("crypto-exchange", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, user)
	return err
}

// GetUserByID retrieves a user by ID
func (user *User) GetUserByID() error {
	collection := GetCollection("crypto-exchange", "users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"ID": user.ID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return err
	}
	return nil
}
