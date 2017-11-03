package godbot

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Core is the basic structure of the bot.
type Core struct {
	sync.Mutex
	muUpdate sync.Mutex
	// User Information
	User     *discordgo.User
	Username string
	Status   string
	ID       int
	Token    string

	Stream bool
	Game   string

	// Ready channel
	ready chan string

	// Connection Information.
	Session     *discordgo.Session
	ChannelMain *discordgo.Channel
	Channels    []*discordgo.Channel
	GuildMain   *discordgo.Guild
	Guilds      []*discordgo.Guild

	// Link map: [guild ID] []*Channels
	Links   map[string][]*discordgo.Channel
	Private []*discordgo.Channel

	// Message handling functions.
	mch func(*discordgo.Session, *discordgo.MessageCreate)
	muh func(*discordgo.Session, *discordgo.MessageUpdate)

	// Member handlers
	gmah func(*discordgo.Session, *discordgo.GuildMemberAdd)
	gmuh func(*discordgo.Session, *discordgo.GuildMemberUpdate)
	gmrh func(*discordgo.Session, *discordgo.GuildMemberRemove)

	// Guild handlers
	gah  func(*discordgo.Session, *discordgo.GuildCreate)
	gruh func(*discordgo.Session, *discordgo.GuildRoleUpdate)
	grdh func(*discordgo.Session, *discordgo.GuildRoleDelete)

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
}
