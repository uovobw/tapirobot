package tumblr

import (
	"log"
	"regexp"
	"strings"

	"github.com/MariaTerzieva/gotumblr"
	"github.com/uovobw/tapiro/utils/linklist"
	"github.com/whyrusleeping/hellabot"
)

const (
	urlRegexp = "(?:([^:/?#]+):)?(?://([^/?#]*))?([^?#]*\\.(?:jpg|gif|png|jpeg|bmp))(?:\\?([^#]*))?(?:#(.*))?"
)

type Tumblr struct {
	api       *gotumblr.TumblrRestClient
	tumblrUrl string
}

var t Tumblr

func init() {
	t = Tumblr{}
}

func Configure(appKey, appSecret, oauthToken, oauthTokenSecret, tumblrUrl string) (err error) {
	t.api = gotumblr.NewTumblrRestClient(appKey, appSecret, oauthToken, oauthTokenSecret, tumblrUrl, "http://api.tumblr.com")
	t.tumblrUrl = tumblrUrl
	//TODO: error handling?
	return nil
}

func getUrlsFromMessage(msg string) (result []string) {
	imgRegexp := regexp.MustCompile(urlRegexp)
	imagesUrls := imgRegexp.FindAllString(msg, -1)
	for _, url := range imagesUrls {
		if strings.Contains(url, " ") {
			for _, splitted := range strings.Split(url, " ") {
				if splitted != "" {
					result = append(result, splitted)
				}
			}
		} else {
			result = append(result, url)
		}
	}
	return
}

func GetTrigger() (trigger *hbot.Trigger) {
	trigger = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Command == "PRIVMSG" {
				if len(getUrlsFromMessage(mes.Content)) != 0 {
					return true
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			tagsRe := regexp.MustCompile("\\[(.+?)\\]")
			dirtyTags := tagsRe.FindAllString(mes.Content, -1)
			tags := make([]string, 0, len(dirtyTags))
			for _, tag := range dirtyTags {
				if strings.Contains(tag, "[img]") == false && strings.Contains(tag, "[/img]") == false {
					tags = append(tags, strings.Replace(strings.Trim(tag, "[] "), " ", "_", -1))
				}
			}
			tagList := strings.Join(tags, ", ")
			caption := strings.Join(tags, " ")
			for _, image := range getUrlsFromMessage(mes.Content) {
				if !linklist.Uniq(image) {
					continue
				}
				caption = caption + " from: " + image
				t.postImage(tagList, image, caption)
			}
			return true
		},
	}
	return trigger
}

func (t *Tumblr) postImage(taglist, url, caption string) {
	options := map[string]string{
		"tags":    taglist,
		"source":  url,
		"caption": caption,
	}
	log.Printf("Posting image: %s with tags %s\n", url, taglist)
	err := t.api.CreatePhoto(t.tumblrUrl, options)
	if err != nil {
		log.Printf("Failed to post image: %s\n", err)
	}
}
