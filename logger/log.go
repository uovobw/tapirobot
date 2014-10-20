package logger

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/whyrusleeping/hellabot"
)

type Message struct {
	Nickname  string
	Timestamp time.Time
	Content   string
}

type Logger struct {
	channelname string
	db          *gorm.DB
}

var l Logger

func init() {
	l = Logger{}
}

func Configure(channelname, location string) (err error) {
	l.channelname = channelname
	db, err := gorm.Open("sqlite3", location)
	if err != nil {
		return err
	}
	l.db = &db
	l.db.AutoMigrate(&Message{})
	return err
}

func GetTrigger() (trigger *hbot.Trigger) {
	trigger = &hbot.Trigger{
		func(mes *hbot.Message) bool {
			if mes.Command == "PRIVMSG" {
				if mes.To == l.channelname {
					return true
				}
			}
			return false
		},
		func(irc *hbot.IrcCon, mes *hbot.Message) bool {
			l.log(mes)
			return true
		},
	}
	return trigger
}

func (l *Logger) log(msg *hbot.Message) {
	m := Message{
		Nickname:  msg.From,
		Timestamp: msg.TimeStamp,
		Content:   msg.Content,
	}
	l.db.Save(&m)
}
