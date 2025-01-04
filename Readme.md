# VideoDownloaderBot

This repository is a simple source code for tg bot, that listens on messages in chats, detects urls and replace this messages with video files


- Third-party API is used to download videos: https://github.com/imputnet/cobalt

- ```Application - entrypoint struct, it designed with observable pattern, so any additional service that implements observer can be added```

```go
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
```

``` In this example botDownloadService - main service, that downloads videos, new event - Update struct```

``` Each individual service has its own goroutine```