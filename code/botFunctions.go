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

func downloadVideo(url string) *os.File {

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
		return nil
	}

	file, err := cobaltApi.DownLoadVideo(res)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(fmt.Sprintf("fetching video done, time taken: %v", time.Since(start)))
	return file
}

func closeFile(file *os.File) {
	_ = file.Close()
	_ = os.Remove(file.Name())
}

func sendVideo(bot *tgbotapi.BotAPI, file *os.File, msg *tgbotapi.Message, deleteOld bool) {
	start := time.Now()
	filePath := tgbotapi.FilePath(file.Name())
	dataToSend := tgbotapi.NewVideo(msg.Chat.ID, filePath)
	dataToSend.Caption = fmt.Sprintf("отправил @%s", msg.From.UserName)
	_, err := bot.Send(dataToSend)

	if err != nil {
		log.Println(err)
		return
	}

	if deleteOld {
		_, _ = bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
	}

	log.Println(fmt.Sprintf("sending video done, time taken: %v", time.Since(start)))
}

func processVideo(bot *tgbotapi.BotAPI, fileUrl string, message *tgbotapi.Message, group *sync.WaitGroup) {
	defer group.Done()
	file := downloadVideo(fileUrl)
	if file != nil {
		sendVideo(bot, file, message, message.Text == fileUrl)
		closeFile(file)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rxStrict := xurls.Strict
	urls := rxStrict.FindAllString(message.Text, -1)
	group := sync.WaitGroup{}

	found := make(map[string]bool)

	for _, url := range urls {

		log.Println(fmt.Sprintf("found url in message %d from user %s", message.MessageID, message.From.UserName))
		found[url] = true
	}

	for url, _ := range found {
		group.Add(1)
		go processVideo(bot, url, message, &group)
	}

	group.Wait()
}
