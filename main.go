package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

func connectToDB() *dynamodb.Client {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB client
	return dynamodb.NewFromConfig(cfg)
}

func connectToTelegramBot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhook(os.Getenv("NGROK_URL") + bot.Token)

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	return bot
}
func main() {
	dbClient := connectToDB()
	bot := connectToTelegramBot()
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
