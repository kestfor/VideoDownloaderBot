package main

import (
	"log"
	"os"
	"videoDownloader/bot"
)

func main() {

	token := os.Getenv("BOT_TOKEN")
	urlApi := os.Getenv("COBALT_API_URL")
	app := NewApplication(token)
	updates := app.GetUpdateChan()

	botDownloadService := bot.NewBotDownloadService(app.Bot, urlApi)
	_ = app.AddObserver(botDownloadService)

	for update := range updates {
		err := app.HandleUpdate(update)
		if err != nil {
			log.Println(err)
		}
	}
}
