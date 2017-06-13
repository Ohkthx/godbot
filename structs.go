package godbot

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Core is the basic structure of the bot.
type Core struct {
	sync.Mutex
	done     bool
	muUpdate sync.Mutex
	// User Information
	User     *discordgo.User
	Username string
	Status   string
	ID       int
	Token    string

	Stream bool
	Game   string

	// Connection Information.
	session     *discordgo.Session
	channelMain *discordgo.Channel
	channels    []*discordgo.Channel
	guildMain   *discordgo.Guild
	guilds      []*discordgo.Guild

	links   map[string][]*discordgo.Channel
	private []*discordgo.Channel

	// Message handling function.
	mhAssigned bool
	mh         func(*discordgo.Session, *discordgo.MessageCreate)

	// Logging for Errors.
	muLog  sync.Mutex
	errlog *log.Logger
}

// Connections holds all connection data.
type Connections struct {
	Links    map[string][]*discordgo.Channel
	Guilds   []*discordgo.Guild
	Channels []*discordgo.Channel
}

// User contains user data.
type User struct {
	*discordgo.User
}
