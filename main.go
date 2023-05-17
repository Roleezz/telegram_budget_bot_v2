package main

import (
	"github.com/Roleezz/telegram_budget_bot_v2/storage"
	"github.com/Roleezz/telegram_budget_bot_v2/telegram"
)

func main() {
	storageClient := new(storage.Client)
	storageClient.Connect()

	telegramClient := new(telegram.Client)
	telegramClient.ConnectToTelegramBot()

	telegramClient.Update(storageClient.CalculateTotal, storageClient.Write)

}
