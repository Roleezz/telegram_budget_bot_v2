package telegram

import (
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Client struct {
	Bot *tgbotapi.BotAPI
}

func (client *Client) writeMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := client.Bot.Send(msg); err != nil {
		log.Panic(err)
	}
}

func (client *Client) ConnectToTelegramBot() {
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
	client.Bot = bot
}

func (client *Client) Update(totalSum func() int, write func(chatID int64, date int, value int)) {

	updates := client.Bot.ListenForWebhook("/" + client.Bot.Token)
	go http.ListenAndServe("0.0.0.0:8443", nil)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if update.Message.IsCommand() { // ignore any non-command Messages
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "stats":
				total := totalSum()
				strconv.Itoa(total)
				client.writeMessage(update.Message.Chat.ID, strconv.Itoa(total))
			}
		} else {
			re := regexp.MustCompile(`^\s*\d+\.?\d+`)
			answer := re.FindAllString(update.Message.Text, 1)

			if answer != nil {
				number, err := strconv.Atoi(answer[0])

				if err != nil {
					log.Fatal(err)
				}

				write(update.Message.Chat.ID, update.Message.Date, number)
			} else {
				client.writeMessage(update.Message.Chat.ID, "Not matched")
			}
		}
	}
}
