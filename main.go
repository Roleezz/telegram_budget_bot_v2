package main

import (
	"github.com/Roleezz/telegram_budget_bot_v2/storage"
	"github.com/Roleezz/telegram_budget_bot_v2/telegram"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func main() {
	storageClient := new(storage.Client)
	storageClient.Connect()
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
				total := storageClient.Read()
				telegram.WriteMessage(bot, update.Message.Chat.ID, strconv.Itoa(total))
			}
		} else {
			re := regexp.MustCompile(`^\s*\d+\.?\d+`)
			answer := re.FindAllString(update.Message.Text, 1)

			if answer != nil {
				number, err := strconv.Atoi(answer[0])

				if err != nil {
					log.Fatal(err)
				}

				storageClient.Write(update.Message.Chat.ID, update.Message.Date, number)
			} else {
				telegram.WriteMessage(bot, update.Message.Chat.ID, "Not matched")
			}
		}
	}
}
