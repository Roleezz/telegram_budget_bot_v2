package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Roleezz/telegram_budget_bot_v2/db"
	"github.com/Roleezz/telegram_budget_bot_v2/telegram"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TableRow struct {
	ChatID    string `json:"ChatID"`
	CreatedAt string `json:"CreatedAt"`
	Value     string `json:"Value"`
}

func readFromTable(chatID int64, dbClient *dynamodb.Client) int {
	// Make the query to DynamoDB
	result, err := dbClient.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("budget-transactions"),
		KeyConditionExpression: aws.String("ChatID = :v_condition"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v_condition": &types.AttributeValueMemberS{Value: strconv.FormatInt(chatID, 10)},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Items)
	total := 0
	var row TableRow
	for item := range result.Items {
		err = attributevalue.UnmarshalMap(result.Items[item], &row)

		if err != nil {
			fmt.Println("Failed to unmarshal DynamoDB item", err)
			return 0
		}
		number, err := strconv.Atoi(row.Value)

		if err != nil {
			log.Fatal(err)
		}
		total += number
	}
	return total
}

func writeMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func writeToTable(chatID int64, date int, value int, dbClient *dynamodb.Client) {
	out, err := dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("budget-transactions"),
		Item: map[string]types.AttributeValue{
			"ChatID":    &types.AttributeValueMemberS{Value: strconv.FormatInt(chatID, 10)},
			"CreatedAt": &types.AttributeValueMemberS{Value: strconv.Itoa(date)},
			"Value":     &types.AttributeValueMemberS{Value: strconv.Itoa(value)},
		},
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(out.Attributes)

}

func main() {
	dbClient := db.ConnectToDB()
	bot := telegram.ConnectToTelegramBot()
	updates := bot.ListenForWebhook("/" + bot.Token)

	go http.ListenAndServe("0.0.0.0:8443", nil)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "stats":
				total := readFromTable(update.Message.Chat.ID, dbClient)
				writeMessage(bot, update.Message.Chat.ID, strconv.Itoa(total))
			}
		} else {
			re := regexp.MustCompile(`^\s*\d+\.?\d+`)
			answer := re.FindAllString(update.Message.Text, 1)

			if answer != nil {
				number, err := strconv.Atoi(answer[0])

				if err != nil {
					log.Fatal(err)
				}
				writeToTable(update.Message.Chat.ID, update.Message.Date, number, dbClient)
			} else {
				writeMessage(bot, update.Message.Chat.ID, "Not matched")
			}
		}
	}
}
