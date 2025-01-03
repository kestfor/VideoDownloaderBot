package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"videoDownloader/subs"
)

type Application struct {
	observers []subs.Observer
	Bot       *tgbotapi.BotAPI
}

func (app *Application) NotifyObservers(event any) {
	for index := range app.observers {
		observer := app.observers[index]
		go observer.Update(event)
	}
}

func (app *Application) AddObserver(observer subs.Observer) error {
	app.observers = append(app.observers, observer)
	return nil
}

func (app *Application) DetachObserver(observer subs.Observer) error {
	for index, obs := range app.observers {
		if obs == observer {
			app.observers = append(app.observers[0:index], app.observers[index:]...)
			break
		}
	}
	return nil
}

func NewApplication(token string) *Application {
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		panic(err)
	}

	return &Application{Bot: bot, observers: make([]subs.Observer, 0)}
}

func (app *Application) GetUpdateChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10
	updates := app.Bot.GetUpdatesChan(u)
	return updates
}

func (app *Application) HandleUpdate(update tgbotapi.Update) error {
	app.NotifyObservers(update)
	return nil
}
