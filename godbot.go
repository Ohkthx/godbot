package godbot

import (
	"errors"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Error constants
var (
	_version      = "0.1.9"
	ErrNilToken   = errors.New("token is not set")
	ErrNilHandler = errors.New("message handler not assigned")
)

// New creates a new instance of the bot.
func New(token string) (*Core, error) {
	return &Core{Token: token}, nil
}

// MessageCreateHandler assigns a function to handle messages.
func (bot *Core) MessageCreateHandler(msgHandler func(*discordgo.Session, *discordgo.MessageCreate)) {
	bot.mch = msgHandler
}

// MessageUpdateHandler assigns a function to handle messages.
func (bot *Core) MessageUpdateHandler(msgHandler func(*discordgo.Session, *discordgo.MessageUpdate)) {
	bot.muh = msgHandler
}

// GuildMemberAddHandler assigns a function to deal with newly joining users.
func (bot *Core) GuildMemberAddHandler(userHandler func(*discordgo.Session, *discordgo.GuildMemberAdd)) {
	bot.uah = userHandler
}

// GuildMemberRemoveHandler assigns a function to deal with leaving users.
func (bot *Core) GuildMemberRemoveHandler(userHandler func(*discordgo.Session, *discordgo.GuildMemberRemove)) {
	bot.urh = userHandler
}

// GuildCreateHandler assigns a function to deal with newly create guilds.
func (bot *Core) GuildCreateHandler(createHandler func(*discordgo.Session, *discordgo.GuildCreate)) {
	bot.gah = createHandler
}

// Start initiates the bot, attempts to connect to Discord.
func (bot *Core) Start() error {
	var err error

	if bot.Token == "" {
		return ErrNilToken
	} else if bot.mch == nil {
		return ErrNilHandler
	} else if bot.muh == nil {
		return ErrNilHandler
	}

	err = bot.setupLogger()
	if err != nil {
		return err
	}

	bot.Session, err = discordgo.New("Bot " + bot.Token)
	if err != nil {
		return err
	}

	// Ready callback for when application is ready.
	bot.Session.AddHandler(bot.ready)

	// Message handler for MessageCreate and MessageUpdate
	bot.Session.AddHandler(bot.mch)
	bot.Session.AddHandler(bot.muh)

	// Handlers for channel changes
	bot.Session.AddHandler(bot.channelCreated)
	bot.Session.AddHandler(bot.channelDeleted)
	bot.Session.AddHandler(bot.channelUpdated)

	// Optional handlers.
	if bot.uah != nil {
		bot.Session.AddHandler(bot.uah)
	}
	if bot.urh != nil {
		bot.Session.AddHandler(bot.urh)
	}
	if bot.gah != nil {
		bot.Session.AddHandler(bot.gah)
	}

	err = bot.Session.Open()
	if err != nil {
		bot.errorlog(err)
		bot.Stop()
		return err
	}

	for bot.done == false {
		if bot.done == true {
			break
		}
	}

	return nil
}

// Stop shuts down the bot.
func (bot *Core) Stop() error {
	//bot.Unlock()
	bot.Session.Close()
	return nil
}

func (bot *Core) ready(s *discordgo.Session, event *discordgo.Ready) {
	//s := bot.session
	bot.Lock()
	defer bot.Unlock()

	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}

	bot.User, err = s.User("@me")
	if err != nil {
		bot.errorlog(err)
		return
	}

	bot.GuildMain = bot.Guilds[0]
	bot.ChannelMain = bot.GetMainChannel(bot.GuildMain.ID)

	if bot.Game != "" {
		err = s.UpdateStatus(0, bot.Game)
		if err != nil {
			bot.errorlog(err)
			return
		}
	}

	bot.done = true
}

func (bot *Core) channelCreated(s *discordgo.Session, cc *discordgo.ChannelCreate) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}

func (bot *Core) channelDeleted(s *discordgo.Session, cd *discordgo.ChannelDelete) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}

func (bot *Core) channelUpdated(s *discordgo.Session, cu *discordgo.ChannelUpdate) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}

func (bot *Core) setupLogger() error {
	bot.errlog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	f, err := os.OpenFile("stderr.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	bot.errlog.SetOutput(f)
	return nil
}

func (bot *Core) errorlog(err error) {
	bot.muLog.Lock()
	defer bot.muLog.Unlock()
	bot.errlog.Println(err)
}
