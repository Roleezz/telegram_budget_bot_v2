package telegram

import (
	"log"
	"net/http"
	"os"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func writeMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
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

func Update() {
	//storageClient := new(storage.Client)
	//storageClient.Connect()
	connect := connectToTelegramBot()
	updates := connect.ListenForWebhook("/" + connect.Token)
	go http.ListenAndServe("0.0.0.0:8443", nil)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "stats":
				// total := storageClient.CalculateTotal()
				// strconv.Itoa(total)
				writeMessage(connect, update.Message.Chat.ID, "55")
			}
		} else {
			re := regexp.MustCompile(`^\s*\d+\.?\d+`)
			answer := re.FindAllString(update.Message.Text, 1)

			if answer != nil {
				writeMessage(connect, update.Message.Chat.ID, "Nice")
				// number, err := strconv.Atoi(answer[0])

				// if err != nil {
				// 	log.Fatal(err)
				// }

				// storageClient.Write(update.Message.Chat.ID, update.Message.Date, number)
			} else {
				writeMessage(connect, update.Message.Chat.ID, "Not matched")
			}
		}
	}
}
