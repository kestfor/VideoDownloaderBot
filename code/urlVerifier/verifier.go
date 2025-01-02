package urlVerifier

import "strings"

const (
	NotFound  = ""
	TikTok    = "tiktok"
	YouTube   = "youtu"
	Instagram = "insta"
)

var urls = []string{TikTok, Instagram}

func GetUrlType(url string) string {

	for _, domain := range urls {
		if strings.Contains(url, domain) {
			return domain
		}
	}
	return NotFound
}
