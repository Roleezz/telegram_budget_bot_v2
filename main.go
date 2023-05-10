package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func writeMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func dataAccumulation(text string, total *int) {
	number, err := strconv.Atoi(text)
	if err != nil {
		log.Fatal(err)
	}
	*total += number
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
	bot, err := tgbotapi.NewBotAPI("5505578325:AAE4sHnqPUc-VYi9rYQpuJsj4VPxAP_7_Uc")
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhook("https://3c7f-93-95-139-165.ngrok-free.app/" + bot.Token)

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

	total := 0

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
				text := strconv.Itoa(total)
				writeMessage(bot, update.Message.Chat.ID, text)
			}
		} else {
			re := regexp.MustCompile(`^\s*\d+\.?\d+`)
			answer := re.FindAllString(update.Message.Text, 1)

			if answer != nil {
				dataAccumulation(answer[0], &total)
			} else {
				writeMessage(bot, update.Message.Chat.ID, "Not matched")
			}
		}
	}
}
