package commands

import (
	"fmt"
	"log"
	"strings"
	"time"
	"flag"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whyrusleeping/hellabot"

	"code.google.com/p/google-api-go-client/googleapi/transport"
	"code.google.com/p/google-api-go-client/youtube/v3"
)

type Commands struct {
	Identifier string
	commands   map[string]interface{}
	db         *gorm.DB
	self       string
	ytapikey   string
}

var (
	c Commands
	ytservice *youtube.Service
)

func init() {
	c = Commands{}
	c.commands = map[string]interface{}{
		"help": help,
		"seen": lastSeen,
		"tell": tell,
		"ack":  ack,
		"youtube": youtube,
	}
}

type LastSeen struct {
	Id   int64
	Nick string
	Seen time.Time
}

type TellMessage struct {
	Id          int64
	Sender      string
	Destination string
	Message     string
	Channel     string
}

func ack(irc *hbot.IrcCon, msg *hbot.Message) bool {
	c.db.Where("destination like ?", msg.Name).Delete(TellMessage{})
	irc.Channels[msg.To].Say(fmt.Sprintf("Ok %s all your pending messages deleted", msg.Name))
	return true
}

func tell(irc *hbot.IrcCon, msg *hbot.Message) bool {
	splitted := strings.SplitN(msg.Content, " ", 3)
	if len(splitted) == 3 {
		tm := TellMessage{
			Destination: splitted[1],
			Message:     splitted[2],
			Sender:      msg.Name,
			Channel:     msg.To,
		}
		c.db.Save(&tm)
		irc.Channels[msg.To].Say("Message saved")
	} else {
		irc.Channels[msg.To].Say(fmt.Sprintf("Command TELL: %stell <nickname> <message>", c.Identifier))
	}
	return true
}

func lastSeen(irc *hbot.IrcCon, msg *hbot.Message) bool {
	log.Printf("Running lastSeen command")
	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	fields := strings.Fields(msg.Content)
	if len(fields) < 2 {
		irc.Channels[msg.To].Say(fmt.Sprintf("Command TELL: %sseen <nickname>", c.Identifier))
		return true
	}
	nick := fields[1]
	seen := LastSeen{}
	c.db.Where("nick = ?", nick).First(&seen)
	if seen.Id == 0 {
		irc.Channels[msg.To].Say(fmt.Sprintf("Never seen %s, sorry", nick))
	} else {
		irc.Channels[msg.To].Say(fmt.Sprintf("Last seen %s at %s", nick, seen.Seen.Format(layout)))
	}
	return true
}

func youtube(irc *hbot.IrcCon, msg *hbot.Message) bool {
	log.Printf("Running yt command")
	splitted := strings.SplitN(msg.Content, " ", 2)
	if len(splitted) < 2 {
		irc.Channels[msg.To].Say(fmt.Sprintf("Youtube search: %syt <query>", c.Identifier))
		return true
	} else {
		var query = flag.String("query", splitted[1], "Query string")
		var maxResults = flag.Int64("max-results", 1, "Max results")
		flag.Parse()

		// API call
		call := ytservice.Search.List("id,snippet").
				Q(*query).
				MaxResults(*maxResults)
		response, err := call.Do()
		if err != nil { log.Printf("Error making search API call: %v", err); return true }

		// Parse response
		var found bool = false
		for _, item := range response.Items {
			if item.Id.Kind == "youtube#video" {
				var returnMsg string = item.Snippet.Title + " - https://youtu.be/" + item.Id.VideoId
				found = true
				log.Printf(returnMsg)
				irc.Channels[msg.To].Say(fmt.Sprintf("[YouTube] %s", returnMsg))
				continue
			}
		}

		if !found {
			irc.Channels[msg.To].Say(fmt.Sprintf("No videos found!"))
			return true
		}
	}
	return true
}

func help(irc *hbot.IrcCon, msg *hbot.Message) bool {
	log.Printf("Running help command")
	keys := []string{}
	for k := range c.commands {
		keys = append(keys, k)
	}
	irc.Channels[msg.To].Say(fmt.Sprintf("Available commands: %s", strings.Join(keys, ",")))
	return true
}

func Configure(identifier string, dblocation string, ytapikey string, self string, bot *hbot.IrcCon) (err error) {
	c.Identifier = identifier
	c.self = self
	log.Printf("Command module running with command identifier %s", identifier)
	db, err := gorm.Open("sqlite3", dblocation)
	if err != nil {
		return err
	}
	c.db = &db
	c.db.AutoMigrate(&LastSeen{})
	c.db.AutoMigrate(&TellMessage{})
	c.ytapikey = ytapikey
	ytclient := &http.Client{ Transport: &transport.APIKey{Key: c.ytapikey}, }
	ytservice, err = youtube.New(ytclient)
	if err != nil {
		return err
	}
	lastSeenTrigger := &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Name != self && mes.Command == "PRIVMSG" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			lastseen := LastSeen{}
			err := c.db.Where("nick = ?", mes.Name).Find(&lastseen).Error
			if err != nil {
				log.Printf("Cannot find last seen for user %s: %s", mes.Name, err)
			}
			if lastseen.Nick == "" {
				lastseen.Nick = mes.Name
			}
			lastseen.Seen = time.Now()
			c.db.Save(&lastseen)
			return false
		},
	}
	tellTrigger := &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Command == "JOIN" {
				if mes.Name != c.self {
					return true
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			rows, err := c.db.Model(TellMessage{}).Where("destination = ?", mes.Name).Select("sender, message, channel").Rows()
			if err != nil {
				log.Println("ERROR in tell message: %s", err)
			}
			defer rows.Close()
			var message, channel, from string
			for rows.Next() {
				rows.Scan(&from, &message, &channel)
				irc.Channels[channel].Say(fmt.Sprintf("Hey %s! %s told me to tell you: \"%s\"", mes.Name, from, message))
			}
			irc.Channels[channel].Say("Use !ack to acnowledge all messages or i will repeat them each time you log in")
			return true
		},
	}
	bot.AddTrigger(tellTrigger)
	bot.AddTrigger(lastSeenTrigger)
	return nil
}

func GetTrigger() *hbot.Trigger {
	trigger := &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Command == "PRIVMSG" {
				if mes.Content != "" && string(mes.Content[0]) == c.Identifier {
					return true
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			return handleCommand(irc, mes)
		},
	}
	return trigger
}

func handleCommand(irc *hbot.IrcCon, mes *hbot.Message) bool {
	log.Printf("Handle command invoked on: %s", mes.Content)
	fields := strings.Fields(mes.Content)
	handler, ok := c.commands[string(fields[0][1:])]
	if !ok {
		irc.Channels[mes.To].Say(fmt.Sprintf("Command %s unknown", fields[0]))
		return false
	}
	return handler.(func(*hbot.IrcCon, *hbot.Message) bool)(irc, mes)
}
