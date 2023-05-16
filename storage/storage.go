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
	dbClient *mongo.Client
}

func (client *Client) database() *mongo.Database {
	return client.dbClient.Database("budget-bot")
}

func (client *Client) collection() *mongo.Collection {
	return client.database().Collection("transactions")
}

func (client *Client) Connect() {
	println("Connect to the storage")

	clientOptions := options.Client().ApplyURI("mongodb://root:example@localhost:27017/")
	dbClient, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		panic(err)
	}

	client.dbClient = dbClient

	var result bson.M
	if err := client.database().RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}

	println("Successfully connected to the storage")
}

func (client *Client) Write(chatID int64, date int, value int) {
	// Insert a single document
	res, err := client.collection().InsertOne(context.TODO(), bson.M{
		"ChatID":    chatID,
		"CreatedAt": date,
		"Value":     value})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID %v\n", res.InsertedID)
}
