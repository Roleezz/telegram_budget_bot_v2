package storage

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client *mongo.Client
}

func (storage *Client) Connect() {
	println("Connect to the storage")

	clientOptions := options.Client().ApplyURI("mongodb://root:example@localhost:27017/")
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		panic(err)
	}

	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}

	println("Successfully connected to the storage")

	storage.client = client
}

func (storage *Client) Read() int {
	return 100
}

func (storage *Client) Write(chatID int64, date int, value int) {
	collection := storage.client.Database("budget-bot").Collection("transactions")

	// Insert a single document
	res, err := collection.InsertOne(context.TODO(), bson.M{
		"ChatID":    chatID,
		"CreatedAt": date,
		"Value":     value})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID %v\n", res.InsertedID)
}
