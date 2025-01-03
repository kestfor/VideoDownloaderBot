package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var apiUrl = ""

func main() {

	godotenv.Load()
	token := os.Getenv("BOT_TOKEN")
	app := NewApplication(token)
	updates := app.GetUpdateChan()

	for update := range updates {
		err := app.HandleUpdate(update)
		if err != nil {
			log.Println(err)
		}
	}
}
