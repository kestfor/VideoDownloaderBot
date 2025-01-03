package bot

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

var apiUrl = os.Getenv("COBALT_API_URL")

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
		fi, err := file.Stat()
		if err != nil || fi.Size() == 0 {
			continue
		}

		media := tgbotapi.NewInputMediaVideo(tgbotapi.FilePath(file.Name()))
		if index == 0 && addCaption {
			media.Caption = caption
		}
		filesToSend = append(filesToSend, media)
	}

	return tgbotapi.NewMediaGroup(message.Chat.ID, filesToSend)
}

func extractUrls(text string) []string {
	rxStrict := xurls.Strict
	urls := rxStrict.FindAllString(text, -1)
	return urls
}

func getUnique(data []string) []string {
	found := make(map[string]bool)

	for index, url := range data {
		if index >= URL_NUM_LIMIT {
			break
		}
		found[url] = true
	}

	res := make([]string, 0, len(found))
	for url, _ := range found {
		res = append(res, url)
	}

	return res
}

func getUrlsTotalLen(urls []string) int {
	res := 0
	for _, url := range urls {
		res += len(url)
	}
	return res
}

func getMessageLen(msg string) int {
	r, _ := regexp.Compile(`\S+`)
	words := r.FindAllString(msg, -1)
	return getUrlsTotalLen(words)
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	urls := extractUrls(message.Text)
	uniqueUrls := getUnique(urls)
	if len(uniqueUrls) == 0 {
		return
	}
	onlyUrlsInMsg := getMessageLen(message.Text) == getUrlsTotalLen(urls)

	log.Println(fmt.Sprintf("found %d urls in message %d by @%s", len(uniqueUrls), message.MessageID, message.From.UserName))

	//create tasks for downloading extracted videos
	group := sync.WaitGroup{}
	filesChan := make(chan *os.File, len(uniqueUrls))

	for _, url := range uniqueUrls {
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

	mediaGroup := createMediaGroup(files, message, onlyUrlsInMsg)
	onlyUrlsInMsg = onlyUrlsInMsg && len(files) == len(mediaGroup.Media)

	if !onlyUrlsInMsg {
		mediaGroup.ReplyToMessageID = message.MessageID
	}
	_, _ = bot.Send(mediaGroup)

	//delete old message if it only contains links
	if onlyUrlsInMsg {
		_, _ = bot.Send(tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID))
	}

	for _, file := range files {
		closeFile(file)
	}

	log.Println(fmt.Sprintf("sending video done, time taken: %v", time.Since(start)))
}
