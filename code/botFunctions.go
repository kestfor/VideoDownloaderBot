package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mvdan/xurls"
	"log"
	"os"
	"sync"
	"time"
	"videoDownloader/cobalt"
)

func downloadVideo(url string, files chan<- *os.File, group *sync.WaitGroup) {
	defer group.Done()

	cobaltApi := cobalt.NewCobaltInstance(apiUrl)

	start := time.Now()

	var err error = nil
	var res cobalt.Response
	for _ = range 5 {
		res, err = cobaltApi.FindVideo(url)
		if err != nil {
			continue
		} else {
			break
		}
	}

	if err != nil {
		log.Println(err)
		return
	}

	file, err := cobaltApi.DownLoadVideo(res)
	if err != nil {
		log.Println(err)
		return
	}
	files <- file
	log.Println(fmt.Sprintf("fetching video done, time taken: %v", time.Since(start)))
}

func closeFile(file *os.File) {
	_ = file.Close()
	_ = os.Remove(file.Name())
}

func sendVideos(bot *tgbotapi.BotAPI, files <-chan *os.File, msg *tgbotapi.Message, deleteOld bool) {
	start := time.Now()
	for file := range files {
		filePath := tgbotapi.FilePath(file.Name())
		dataToSend := tgbotapi.NewVideo(msg.Chat.ID, filePath)
		_, err := bot.Send(dataToSend)

		if err != nil {
			log.Println(err)
			closeFile(file)
			continue
		}

		if deleteOld {
			_, _ = bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
		}

		closeFile(file)

	}
	log.Println(fmt.Sprintf("sending video done, time taken: %v", time.Since(start)))
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rxStrict := xurls.Strict
	urls := rxStrict.FindAllString(message.Text, -1)
	group := sync.WaitGroup{}
	files := make(chan *os.File, 10)

	found := make(map[string]bool)

	for _, url := range urls {

		log.Println(fmt.Sprintf("found url in message %d from user %s", message.MessageID, message.From.UserName))
		found[url] = true
	}

	for url, _ := range found {
		group.Add(1)
		go downloadVideo(url, files, &group)
		go sendVideos(bot, files, message, url == message.Text)
	}

	group.Wait()
	close(files)
}
