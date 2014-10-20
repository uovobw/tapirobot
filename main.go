package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/rakyll/globalconf"
	"github.com/uovobw/tapiro/commands"
	"github.com/uovobw/tapiro/logger"
	"github.com/uovobw/tapiro/tumblr"
	"github.com/uovobw/tapiro/twitter"
	"github.com/whyrusleeping/hellabot"
)

const configFile = "tapiro.cfg"

var (
	ircNetwork  = flag.String("network", "chat.freenode.net", "The irc channel to connect to")
	ircPort     = flag.Int("port", 6697, "Port of the host to connect to")
	ircSsl      = flag.Bool("ssl", true, "Use ssl")
	ircNickname = flag.String("nick", "tapiro", "Nickname of the bot")
	ircPassword = flag.String("password", "secret", "Password of the bot")
	ircChannels = flag.String("channels", "##somechannel", "Comma separated list of channels to join")
	debug       = flag.Bool("debug", false, "Turn debug on")

	tumblrConf        = flag.NewFlagSet("tumblr", flag.ExitOnError)
	tumblrAppKey      = tumblrConf.String("appkey", "_", "Tumblr Application Key")
	tumblrAppSecret   = tumblrConf.String("appsecret", "_", "Tumblr Application Secret")
	tumblrOauthToken  = tumblrConf.String("oauthtoken", "_", "Tumblr OAuth Token")
	tumblrOauthSecret = tumblrConf.String("oauthsecret", "_", "Tumblr OAuth Token Secret")
	tumblrUrl         = tumblrConf.String("url", "_", "Tumblr url")

	twitterConf        = flag.NewFlagSet("twitter", flag.ExitOnError)
	twitterAppKey      = twitterConf.String("appkey", "_", "Twitter Application Key")
	twitterAppSecret   = twitterConf.String("appsecret", "_", "Twitter Application Secret")
	twitterOauthToken  = twitterConf.String("oauthtoken", "_", "Twitter OAuth Token")
	twitterOauthSecret = twitterConf.String("oauthsecret", "_", "Twitter OAuth Token Secret")

	loggerConf     = flag.NewFlagSet("logger", flag.ExitOnError)
	loggerLocation = loggerConf.String("location", "_", "Logfile location")
	loggerChannel  = loggerConf.String("channel", "_", "Channel for which to log messages")

	commandsConf       = flag.NewFlagSet("commands", flag.ExitOnError)
	commandsIdentifier = commandsConf.String("identifier", "_", "Command prefix identifier")
	commandsDb         = commandsConf.String("db", "_", "Db location")
)

var config globalconf.GlobalConf
var chanmap map[string]*hbot.IrcChannel

func init() {
	chanmap = make(map[string]*hbot.IrcChannel)
	config, err := globalconf.NewWithOptions(&globalconf.Options{
		Filename: configFile,
	})
	if err != nil {
		log.Fatalf("Cannot parse configuration file: %s", err)
	}
	globalconf.Register("tumblr", tumblrConf)
	globalconf.Register("twitter", twitterConf)
	globalconf.Register("logger", loggerConf)
	globalconf.Register("commands", commandsConf)
	config.ParseAll()
}

func main() {
	if *debug {
		log.Println(*ircNetwork)
		log.Println(*ircChannels)
		log.Println(*ircPort)
		log.Println(*ircSsl)
		log.Println(*ircNickname)
		log.Println(*ircPassword)
		log.Println(*ircChannels)
		log.Println(*tumblrAppKey)
		log.Println(*tumblrAppSecret)
		log.Println(*tumblrOauthToken)
		log.Println(*tumblrOauthSecret)
		log.Println(*tumblrUrl)
		log.Println(*twitterAppKey)
		log.Println(*twitterAppSecret)
		log.Println(*twitterOauthToken)
		log.Println(*twitterOauthSecret)
		log.Println(*loggerLocation)
		log.Println(*loggerChannel)
		log.Println(*commandsDb)
		log.Println(*commandsIdentifier)
	}

	bot, err := hbot.NewIrcConnection(fmt.Sprintf("%s:%d", *ircNetwork, *ircPort), *ircNickname, *ircSsl)
	if err != nil {
		log.Fatalf("Cannot create irc bot: %s", err)
	}
	if *ircPassword != "" {
		bot.Password = *ircPassword
	}

	// Configure plugins
	err = twitter.Configure(*twitterAppKey, *twitterAppSecret, *twitterOauthToken, *twitterOauthSecret)
	if err != nil {
		log.Fatalf("Configuration error: %s", err)
	}
	err = tumblr.Configure(*tumblrAppKey, *tumblrAppSecret, *tumblrOauthToken, *tumblrOauthSecret, *tumblrUrl)
	if err != nil {
		log.Fatalf("Configuration error: %s", err)
	}
	err = logger.Configure(*loggerChannel, *loggerLocation)
	if err != nil {
		log.Fatalf("Configuration error: %s", err)
	}
	err = commands.Configure(*commandsIdentifier, *commandsDb, *ircNickname, bot)
	if err != nil {
		log.Fatalf("Configuration error: %s", err)
	}

	bot.AddTrigger(commands.GetTrigger())
	bot.AddTrigger(twitter.GetTrigger())
	bot.AddTrigger(tumblr.GetTrigger())
	bot.AddTrigger(logger.GetTrigger())

	bot.Start()

	// parse the list of channels
	for _, channel := range strings.Split(*ircChannels, ",") {
		log.Printf("Joining channel: %s\n", channel)
		chanmap[channel] = bot.Join(channel)
	}

	for msg := range bot.Incoming {
		if msg == nil {
			log.Println("Disconnected")
			return
		}
	}

}
