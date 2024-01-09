package alerts

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

func TelegramSend(mnemonic string) {
	godotenv.Load()
	botToken := os.Getenv("TELEGRAM_TOKEN")
	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Println(err)
	}
	msg := tgbotapi.NewMessage(chatID, mnemonic)
	_, err = bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
}
