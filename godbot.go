package godbot

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Error constants
var (
	_version      = "0.2.0"
	ErrNilToken   = errors.New("token is not set")
	ErrNilHandler = errors.New("message handler not assigned")
)

// New creates a new instance of the bot.
func New(token string) (*Core, error) {
	return &Core{Token: token, LiteMode: false}, nil
}

// Start initiates the bot, attempts to connect to Discord.
func (bot *Core) Start() error {
	var err error

	if bot.Token == "" {
		return ErrNilToken
	} else if !bot.LiteMode {
		if bot.mch == nil {
			return ErrNilHandler
		} else if bot.muh == nil {
			return ErrNilHandler
		}
	}

	// Acknowledge the bot is starting.
	fmt.Print("Bot: Core is attempting normal startup... ")

	err = bot.setupLogger()
	if err != nil {
		return err
	}

	bot.Session, err = discordgo.New("Bot " + bot.Token)
	if err != nil {
		return err
	}

	// Ready callback for when application is ready.
	bot.ready = make(chan string)
	bot.Ready = nil
	bot.Session.AddHandler(bot.readyHandler)

	if !bot.LiteMode {
		// Message handler for MessageCreate and MessageUpdate
		bot.Session.AddHandler(bot.mch)
		bot.Session.AddHandler(bot.muh)

		// Handlers for channel changes
		bot.Session.AddHandler(bot.channelCreated)

		// Channel Update Handler.
		if bot.cuh != nil {
			bot.Session.AddHandler(bot.cuh)
		} else {
			bot.Session.AddHandler(bot.channelUpdated)
		}

		// Channel Delete Handler.
		if bot.cdh != nil {
			bot.Session.AddHandler(bot.cdh)
		} else {
			bot.Session.AddHandler(bot.channelDeleted)
		}

		// Member handlers
		bot.Session.AddHandler(bot.gmah)
		bot.Session.AddHandler(bot.gmuh)
		bot.Session.AddHandler(bot.gmrh)

		// Guild operation handlers
		bot.Session.AddHandler(bot.gah)
		bot.Session.AddHandler(bot.gruh)
		bot.Session.AddHandler(bot.grdh)
	}

	err = bot.Session.Open()
	if err != nil {
		bot.errorlog(err)
		bot.Stop()
		return err
	}

	// Wait for the ready to continue.
	for msg := range bot.ready {
		// If the message is ok, return nil
		if msg == "ok" {
			fmt.Println(msg)
			break
		} else {
			// Something wrong happened, returning message.
			fmt.Println("Failed.")
			return errors.New(msg)
		}
	}

	return nil
}

// Stop shuts down the bot.
func (bot *Core) Stop() error {
	//bot.Unlock()
	close(bot.ready)
	bot.Session.Close()
	return nil
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
