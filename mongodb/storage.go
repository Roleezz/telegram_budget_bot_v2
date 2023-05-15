package mongodb

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
}

func (storage *Storage) Сonnect() {
	clinentOptions := options.Client().ApplyURI("mongodb://root:example@mongo:27017/")
	client, err := mongo.Connect(context.TODO(), clinentOptions)

	if err != nil {
		panic(err)
	}

	storage.client = client
}

func (storage *Storage) read() {

}

func (storage *Storage) Write(chatID int64, date int, value int) {
	collection := storage.client.Database("budget-bot").Collection("transactions")

	// Insert a single document
	res, err := collection.InsertOne(context.Background(), bson.M{
		"ChatID":    chatID,
		"CreatedAt": date,
		"Value":     value})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID %v\n", res.InsertedID)
}
