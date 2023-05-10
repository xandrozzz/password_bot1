package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/exp/slices"
	"log"
	"strconv"
	"time"
)

func ClearSecrets(bot *tgbotapi.BotAPI, chatID int64, secretIds []tgbotapi.DeleteMessageConfig, wait int) {
	msg := tgbotapi.NewMessage(chatID, "Для сохранения конфиденциальной информации ваши секретные данные будут удалены через "+
		strconv.Itoa(wait)+" секунд")
	editMsg, err := bot.Send(msg)
	if err != nil {
		panic(err)
	}
	secretIds = append(secretIds, tgbotapi.NewDeleteMessage(chatID, editMsg.MessageID))
	go SetTimer(bot, chatID, editMsg.MessageID, secretIds, wait)

}

func SetTimer(bot *tgbotapi.BotAPI, chatID int64, editMessageID int, secretIds []tgbotapi.DeleteMessageConfig, wait int) {
	currentIds := make([]tgbotapi.DeleteMessageConfig, len(secretIds))
	copy(currentIds, secretIds)
	for wait >= 0 {
		editMessage := tgbotapi.NewEditMessageText(chatID, editMessageID, "Для сохранения конфиденциальной информации ваши секретные данные будут удалены через "+
			strconv.Itoa(wait)+
			" секунд")

		bot.Send(editMessage)
		wait--
		time.Sleep(1 * time.Second)
	}
	log.Print(currentIds)
	for _, deleteMessage := range currentIds {
		bot.Send(deleteMessage)
		i1 := slices.IndexFunc(secretIds, func(m tgbotapi.DeleteMessageConfig) bool { return m.MessageID == deleteMessage.MessageID })
		if i1 != -1 {
			slices.Delete(secretIds, i1, i1+1)
		}
	}
	currentIds = nil
}
