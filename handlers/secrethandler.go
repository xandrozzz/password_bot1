package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	wait--
	for wait > 0 {
		editMessage := tgbotapi.NewEditMessageText(chatID, editMsg.MessageID, "Для сохранения конфиденциальной информации ваши секретные данные будут удалены через "+
			strconv.Itoa(wait)+
			" секунд")
		_, err1 := bot.Send(editMessage)
		if err1 != nil {
			panic(err1)
		}
		wait--
		time.Sleep(1 * time.Second)
	}

	for _, deleteMessage := range secretIds {
		bot.Send(deleteMessage)
	}

	secretIds = nil
}
