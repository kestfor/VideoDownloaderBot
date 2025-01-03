package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mvdan/xurls"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
	"videoDownloader/cobalt"
)

const URL_NUM_LIMIT = 10

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

//func sendVideo(bot *tgbotapi.BotAPI, file *os.File, msg *tgbotapi.Message, deleteOld bool) {
//	start := time.Now()
//	filePath := tgbotapi.FilePath(file.Name())
//	dataToSend := tgbotapi.NewInputMediaVideo(filePath)
//	tgbotapi.NewMediaGroup(msg.Chat.ID)
//	dataToSend.Caption = fmt.Sprintf("отправил @%s", msg.From.UserName)
//	_, err := bot.Send(dataToSend)
//
//	if err != nil {
//		log.Println(err)
//		return
//	}
//
//	if deleteOld {
//		_, _ = bot.Send(tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID))
//	}
//
//	log.Println(fmt.Sprintf("sending video done, time taken: %v", time.Since(start)))
//}

func processVideo(fileUrl string, group *sync.WaitGroup, files chan<- *os.File) {
	defer group.Done()
	file := downloadVideo(fileUrl)
	if file != nil {
		files <- file
	}
}

func createMediaGroup(files []*os.File, message *tgbotapi.Message, addCaption bool) tgbotapi.MediaGroupConfig {
	caption := fmt.Sprintf("отправил @%s", message.From.UserName)
	filesToSend := make([]any, 0)
	for index, file := range files {
		media := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(file.Name()))
		if index == 0 && addCaption {
			media.Caption = caption
		}
		filesToSend = append(filesToSend, media)
	}

	return tgbotapi.NewMediaGroup(message.Chat.ID, filesToSend)
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	rxStrict := xurls.Strict
	urls := rxStrict.FindAllString(message.Text, -1)

	found := make(map[string]bool)
	urlTotalLen := 0

	for index, url := range urls {
		if index >= URL_NUM_LIMIT {
			break
		}
		log.Println(fmt.Sprintf("found url in message %d from user @%s", message.MessageID, message.From.UserName))
		urlTotalLen += len(url)
		found[url] = true
	}

	//create tasks for downloading extracted videos
	group := sync.WaitGroup{}
	filesChan := make(chan *os.File, len(found))

	for url, _ := range found {
		group.Add(1)
		go processVideo(url, &group, filesChan)
	}

	group.Wait()
	close(filesChan)

	start := time.Now()

	//create and send media group
	files := make([]*os.File, 0)
	for file := range filesChan {
		files = append(files, file)
	}

	r, _ := regexp.Compile(`\S+`)
	words := r.FindAllString(message.Text, -1)
	messageLen := 0
	for _, word := range words {
		messageLen += len(word)
	}

	deleteOld := messageLen == urlTotalLen

	mediaGroup := createMediaGroup(files, message, deleteOld)
	if !deleteOld {
		mediaGroup.ReplyToMessageID = message.MessageID
	}
	_, _ = bot.Send(mediaGroup)

	//delete old message if it only contains links
	if deleteOld {
		_, _ = bot.Send(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
	}

	for _, file := range files {
		closeFile(file)
	}

	log.Println(fmt.Sprintf("sending video done, time taken: %v", time.Since(start)))
}
