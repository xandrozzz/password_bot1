package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"password_bot1/handlers"
)

type State struct {
	State   string
	Service string
	Login   string
}

func (u *State) SetState(state string) {
	u.State = state
}

func (u *State) SetService(service string) {
	u.Service = service
}

func (u *State) SetLogin(login string) {
	u.Login = login
}

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err1 := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err1 != nil {
		panic(err1)
	}

	uri := os.Getenv("MONGODB_URI")

	client, err2 := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err2 != nil {
		panic(err2)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err2)
		}
	}()

	users := client.Database("users")

	log.Printf("Connected to MongoDB")

	bot.Debug = true

	log.Printf("Logged in as %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 300

	updates := bot.GetUpdatesChan(updateConfig)

	states := make(map[string]*State)

	var secretIds []tgbotapi.DeleteMessageConfig

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if _, ok := states[update.Message.From.UserName]; !ok {
			newState := State{State: "null", Service: "", Login: ""}
			states[update.Message.From.UserName] = &newState
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch states[update.Message.From.UserName].State {
		case "null":
			switch update.Message.Command() {
			case "help":
				msg.Text = "/set : Сохранить логин и пароль для выбранного сервиса \n" +
					"/get : Получить логин и пароль для выбранного сервиса \n" +
					"/del : Удалить логин и пароль для выбранного сервиса"
			case "hi":
				msg.Text = "Прив :)"
			case "status":
				msg.Text = "Все хорошо!"
			case "start":
				msg.Text = "Я умею сохранять пароли для различных сервисов, для ознакомления с командами напишите /help"
			case "set":
				msg.Text = "Введите название сервиса, для которого вы хотите установить логин и пароль"
				states[update.Message.From.UserName].SetState("set1")
			case "get":
				msg.Text = "Введите название сервиса, логин и пароль для которого вы хотите увидеть"
				states[update.Message.From.UserName].SetState("get")
			case "del":
				msg.Text = "Введите название сервиса, логин и пароль для которого вы бы хотели удалить"
				states[update.Message.From.UserName].SetState("del")
			default:
				msg.Text = "Команда не распознана, введите /help, чтобы получить список доступных команд"
			}
		case "set1":
			states[update.Message.From.UserName].SetService(update.Message.Text)
			msg.Text = "Теперь введите логин"
			states[update.Message.From.UserName].SetState("set2")
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		case "set2":
			states[update.Message.From.UserName].SetLogin(update.Message.Text)
			msg.Text = "Теперь введите пароль"
			states[update.Message.From.UserName].SetState("set3")
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		case "set3":
			handlers.SetMongo(users, update.Message.From.UserName, states[update.Message.From.UserName].Service, states[update.Message.From.UserName].Login, update.Message.Text)
			msg.Text = "Логин и пароль успешно заданы"
			states[update.Message.From.UserName].SetState("null")
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		case "get":
			r := handlers.GetMongo(users, update.Message.From.UserName, update.Message.Text)
			msg.Text = r
			states[update.Message.From.UserName].SetState("get1")
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		case "del":
			r := handlers.DelMongo(users, update.Message.From.UserName, update.Message.Text)
			msg.Text = r
			states[update.Message.From.UserName].SetState("null")
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID))
		}

		sentMsg, err3 := bot.Send(msg)
		if err3 != nil {
			panic(err3)
		}
		switch states[update.Message.From.UserName].State {
		case "set1", "set2", "set3", "del", "get":
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentMsg.MessageID))
		case "get1":
			secretIds = append(secretIds, tgbotapi.NewDeleteMessage(update.Message.Chat.ID, sentMsg.MessageID))
			states[update.Message.From.UserName].SetState("null")
		}

		if states[update.Message.From.UserName].State == "null" && len(secretIds) != 0 {
			handlers.ClearSecrets(bot, update.Message.Chat.ID, secretIds, 30)
		}
	}
}
