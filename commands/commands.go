package commands

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whyrusleeping/hellabot"
)

type Commands struct {
	Identifier string
	commands   map[string]interface{}
	db         *gorm.DB
}

var (
	c Commands
)

func init() {
	c = Commands{}
	c.commands = map[string]interface{}{
		"help": help,
		"seen": lastSeen,
	}
}

type LastSeen struct {
	Id   int64
	Nick string
	Seen time.Time
}

func lastSeen(irc *hbot.IrcCon, msg *hbot.Message) bool {
	log.Printf("Running lastSeen command")
	const layout = "Jan 2, 2006 at 3:04pm (MST)"
	nick := strings.Fields(msg.Content)[1]
	seen := LastSeen{}
	c.db.Where("nick = ?", nick).First(&seen)
	irc.Channels[msg.To].Say(fmt.Sprintf("Last seen %s at %s", nick, seen.Seen.Format(layout)))
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

func Configure(identifier string, dblocation string, self string, bot *hbot.IrcCon) (err error) {
	c.Identifier = identifier
	log.Printf("Command module running with command identifier %s", identifier)
	db, err := gorm.Open("sqlite3", dblocation)
	if err != nil {
		return err
	}
	c.db = &db
	c.db.AutoMigrate(&LastSeen{})
	lastSeenTrigger := &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.From != self && mes.Command == "PRIVMSG" {
				return true
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			lastseen := LastSeen{}
			err := c.db.Where("nick = ?", mes.From).Find(&lastseen).Error
			if err != nil {
				log.Printf("Cannot find last seen for user %s: %s", mes.From, err)
			}
			if lastseen.Nick == "" {
				lastseen.Nick = mes.From
			}
			lastseen.Seen = time.Now()
			c.db.Save(&lastseen)
			return false
		},
	}
	bot.AddTrigger(lastSeenTrigger)
	return nil
}

func GetTrigger() *hbot.Trigger {
	trigger := &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Command == "PRIVMSG" {
				if string(mes.Content[0]) == c.Identifier {
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
