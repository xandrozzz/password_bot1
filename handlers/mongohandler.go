package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Password struct {
	Service  string `bson:"service,omitempty"`
	Login    string `bson:"login,omitempty"`
	Password string `bson:"password,omitempty"`
}

func SetMongo(client *mongo.Client, user string, service string, login string, password string) {
	filter1 := bson.D{}
	names, err := client.Database("users").ListCollectionNames(context.TODO(), filter1)
	if err != nil {
		panic(err)
	}

	t := false
	for _, name := range names {
		if name == user {
			t = true
			break
		}
	}
	if !t {
		client.Database("users").CreateCollection(context.TODO(), user)
	}

	coll := client.Database("users").Collection(user)

	var result bson.M
	err1 := coll.FindOne(context.TODO(), bson.D{{"service", service}}).Decode(&result)
	if err1 != nil {
		newPassword := Password{Service: service, Login: login, Password: password}
		coll.InsertOne(context.TODO(), newPassword)
	} else {
		filter := bson.D{{"service", service}}
		update := bson.D{{"$set", bson.D{{"login", login}, {"password", password}}}}
		coll.UpdateOne(context.TODO(), filter, update)
	}
	err2 := coll.FindOne(context.TODO(), bson.D{{"service", service}}).Decode(&result)
	if err2 != nil {
		panic(err2)
	}
}

func GetMongo(client *mongo.Client, user string, service string) (r string) {
	filter1 := bson.D{}
	names, err := client.Database("users").ListCollectionNames(context.TODO(), filter1)
	if err != nil {
		panic(err)
	}
	t := false
	for _, name := range names {
		if name == user {
			t = true
			break
		}
	}
	if !t {
		return "Вы ещё не задали ни одного логина и пароля"
	}

	coll := client.Database("users").Collection(user)

	var result bson.M
	err1 := coll.FindOne(context.TODO(), bson.D{{"service", service}}).Decode(&result)
	if err1 != nil {
		return "Для данного сервиса логин и пароль не найдены"
	}
	jsonData, err2 := json.MarshalIndent(result, "", "    ")
	if err2 != nil {
		panic(err2)
	}

	m := make(map[string]string)
	err3 := json.Unmarshal(jsonData, &m)
	if err3 != nil {
		log.Fatal(err3)
	}
	r = "Данные для сервиса " + m["service"] + "\nЛогин: " + m["login"] + "\nПароль: " + m["password"]
	return r
}

func DelMongo(client *mongo.Client, user string, service string) (r string) {

	filter1 := bson.D{}
	names, err := client.Database("users").ListCollectionNames(context.TODO(), filter1)
	if err != nil {
		panic(err)
	}
	t := false
	for _, name := range names {
		if name == user {
			t = true
			break
		}
	}
	if !t {
		return "Вы ещё не задали ни одного логина и пароля"
	}

	coll := client.Database("users").Collection(user)

	var result bson.M
	err1 := coll.FindOne(context.TODO(), bson.D{{"service", service}}).Decode(&result)
	if err1 != nil {
		return "Для данного сервиса логин и пароль не найдены"
	}

	filter := bson.D{{"service", service}}
	coll.DeleteOne(context.TODO(), filter)

	return "Логин и пароль успешно удалены"
}
