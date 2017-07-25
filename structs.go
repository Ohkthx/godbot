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
	Session     *discordgo.Session
	ChannelMain *discordgo.Channel
	Channels    []*discordgo.Channel
	GuildMain   *discordgo.Guild
	Guilds      []*discordgo.Guild

	Links   map[string][]*discordgo.Channel
	Private []*discordgo.Channel

	// Message handling function.
	mhAssigned  bool
	uahAssigned bool
	urhAssigned bool
	mh          func(*discordgo.Session, *discordgo.MessageCreate)
	uah         func(*discordgo.Session, *discordgo.GuildMemberAdd)
	urh         func(*discordgo.Session, *discordgo.GuildMemberRemove)

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

/*
Discordgo datatype Wrappers
*/

// User contains user data.
type User struct {
	*discordgo.User
}

// Guild contains guild data.
type Guild struct {
	*discordgo.Guild
}

// Channel contains channel data.
type Channel struct {
	*discordgo.Channel
}

// ChannelLock holds Locking information for a Channel.
type ChannelLock struct {
	Locked     bool
	Session    *discordgo.Session
	Guild      *Guild
	Channel    *Channel
	Roles      []*discordgo.Role
	Overwrites []*discordgo.PermissionOverwrite
	Message    *discordgo.Message
	//Type        string
	//Allow       int
	//Deny        int
	//Permissions int
}
