package twitter

import (
	"fmt"
	"log"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/whyrusleeping/hellabot"
)

const (
	twitterTag  = "@twitter"
	tweetLength = 140
)

type Twitter struct {
	// unexported fields
	api       *anaconda.TwitterApi
	lastTweet string
}

var t Twitter

func init() {
	t = Twitter{}
}

func Configure(appKey, appSecret, oauthToken, oauthTokenSecret string) (err error) {
	anaconda.SetConsumerKey(appKey)
	anaconda.SetConsumerSecret(appSecret)
	t.api = anaconda.NewTwitterApi(oauthToken, oauthTokenSecret)
	ok, err := t.api.VerifyCredentials()
	if !ok {
		return fmt.Errorf("Configuration not correct: %s\n", err)
	}
	return nil
}

func GetTrigger() (trigger *hbot.Trigger) {
	trigger = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if strings.Contains(mes.Content, twitterTag) {
				if t.lastTweet != mes.Content && mes.Command == "PRIVMSG" {
					t.lastTweet = mes.Content
					return true
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			t.postTweet(mes.Content)
			return true
		},
	}
	return trigger
}

func (t *Twitter) postTweet(message string) {
	if len(message) > tweetLength {
		log.Println("Trimming tweet")
		message = message[:140]
	}
	log.Printf("Tweeting: %s\n", message)
	_, _ = t.api.PostTweet(message, nil)
}
