package godbot

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Constants for locked channels.
var (
	ErrChannelNotLocked = errors.New("channel is not locked")
	ErrChannelLocked    = errors.New("channel is locked")
	ErrNilChannelLock   = errors.New("provided a nil channel lock")
	ErrBadChannel       = errors.New("bad channel for operation")
	ErrBadGuild         = errors.New("bad guild for operation")
)

// GetMainChannel sets the main channel for the bot.
func (bot *Core) GetMainChannel(gID string) *discordgo.Channel {
	c := bot.Links[gID]
	for _, p := range c {
		if p.ID == gID {
			return p
		}
	}
	return nil
}

// SetMainChannel sets the channel to primarily sit in.
func (bot *Core) SetMainChannel(gID, cID string) error {
	for _, p := range bot.Links[gID] {
		if p.ID == cID {
			bot.ChannelMain = p
			return nil
		}
	}
	return ErrNotFound
}

// SetMainGuild assigns the guild and channel to the main server.
func (bot *Core) SetMainGuild(gID string) {
	g := bot.GetGuild(gID)
	c := bot.GetChannel(gID)
	bot.GuildMain = g.Guild
	bot.ChannelMain = c.Channel
}

// GetChannel gets a Channel struct based on Channel ID.
func (bot *Core) GetChannel(cID string) *Channel {
	for _, p := range bot.Channels {
		if p.ID == cID {
			return &Channel{Channel: p}
		}
	}
	return nil
}

// GetGuild gets a Guild structure from a Guild ID.
func (bot *Core) GetGuild(gID string) *Guild {
	for _, p := range bot.Guilds {
		if p.ID == gID {
			return &Guild{Guild: p}
		}
	}
	return nil
}

// GetGuildID gets the ID of a guild from a Channel ID.
func (bot *Core) GetGuildID(cID string) (string, error) {
	if bot.Links == nil {
		return "", ErrNilLinks
	}
	for guild, channels := range bot.Links {
		for _, c := range channels {
			if c.ID == cID {
				return guild, nil
			}
		}
	}
	return "", ErrNotFound
}

// ChannelLockCreate returns a ChannelLock struct.
func (bot *Core) ChannelLockCreate(cID string) (*ChannelLock, error) {
	s := bot.Session
	var cl = &ChannelLock{}
	//var f bool

	cl.Session = s
	cl.Channel = bot.GetChannel(cID)
	cl.Guild = bot.GetGuild(cl.Channel.GuildID)

	if cl.Channel.Type != "text" {
		return nil, ErrBadChannel
	}

	for _, p := range cl.Channel.PermissionOverwrites {
		cl.Overwrites = append(cl.Overwrites, p)
		r, err := s.State.Role(cl.Guild.ID, p.ID)
		if err != nil {
			return nil, err
		}

		cl.Roles = append(cl.Roles, r)
	}

	return cl, nil
}

// ChannelLock will lock a channel preventing @everyone typing.
func (cl *ChannelLock) ChannelLock(alert bool) error {
	//var timeoutRole, everyoneRole, everyoneRoleBak *discordgo.Role
	// Get current Roles permissions.
	if cl == nil {
		return ErrNilChannelLock
	}
	if cl.Locked {
		return nil
	}

	s := cl.Session
	for _, ow := range cl.Overwrites {
		r, err := cl.overwriteRole(ow.ID)
		if err != nil {
			fmt.Println("getting role from overwrite id", err)
			continue
		}

		nA := ow.Allow
		if ow.Allow&2048 == 2048 {
			nA = ow.Allow ^ 2048
		}

		err = s.ChannelPermissionSet(cl.Channel.ID, r.ID, ow.Type, nA, ow.Deny|2048)
		if err != nil {
			return err
		}
	}

	if alert {
		d := fmt.Sprintf("**%s** channel is temporarily __**locked**__ for maintenance.\n%4s message will disappear when it is available.", cl.Channel.Name, "This")

		// Embed create.
		em := &discordgo.MessageEmbed{
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       0x800000,
			Description: d,
			Fields:      []*discordgo.MessageEmbedField{},
		}
		var err error
		cl.Message, err = s.ChannelMessageSendEmbed(cl.Channel.ID, em)
		if err != nil {
			return err
		}
		// End Embed send.
	}

	cl.Locked = true
	return nil
}

// ChannelUnlock will unlock a channel allowing for @everyone to type.
func (cl *ChannelLock) ChannelUnlock() error {
	if cl.Locked != true {
		return ErrChannelNotLocked
	}

	s := cl.Session
	for _, ow := range cl.Overwrites {
		r, err := cl.overwriteRole(ow.ID)
		if err != nil {
			fmt.Println("getting role from overwrite id", err)
			continue
		}

		err = s.ChannelPermissionSet(cl.Channel.ID, r.ID, ow.Type, ow.Allow, ow.Deny)
		if err != nil {
			fmt.Printf("[ERROR] Could not unlock!\n Channel: %s -> Role: %s, ID: %s\n Overwrite Dump: %#v", cl.Channel.Name, r.Name, r.ID, ow)
			return err
		}
	}

	if cl.Message != nil {
		err := s.ChannelMessageDelete(cl.Channel.ID, cl.Message.ID)
		if err != nil {
			return err
		}
	}
	cl.Locked = false
	return nil
}

func (cl *ChannelLock) overwriteRole(oID string) (*discordgo.Role, error) {
	for _, r := range cl.Roles {
		if r.ID == oID {
			return r, nil
		}
	}
	return nil, ErrNotFound
}

// SetNickname will set the current name of the bot to the guild.
func (bot *Core) SetNickname(gID, name string, append bool) error {
	s := bot.Session
	if gID == "" {
		return ErrBadGuild
	}

	if append {
		name = fmt.Sprintf("%s %s", bot.Username, name)
	}

	err := s.GuildMemberNickname(gID, bot.User.ID, name)
	if err != nil {
		return err
	}

	return nil
}

// UserID turns a Username#Discriminator into an ID.
func (bot *Core) UserID(name string) (string, error) {
	s := bot.Session
	uname := strings.Split(name, "#")
	if len(uname) < 2 {
		return "", fmt.Errorf("invalid name provided: %s", name)
	}
	for {
		for _, g := range bot.Guilds {
			var uID string

			users, err := s.GuildMembers(g.ID, uID, 100)
			if err != nil {
				return "", err
			}
			ulen := len(users)
			for n, u := range users {
				uID = u.User.ID
				if u.User.Username == uname[0] && u.User.Discriminator == uname[1] {
					return uID, nil
				}
				if n+1 == ulen && ulen < 100 {
					return "", fmt.Errorf("user not found")
				}
			}

		}
	}
}

// GuildsString converts the entire guild list into string format.
func (bot *Core) GuildsString() string {
	var ret = fmt.Sprintf("%20s -> %s\n", "Guild ID", "Guild Name")
	for _, g := range bot.Guilds {
		ret += fmt.Sprintf("%20s -> %s\n", g.ID, g.Name)
	}
	return ret
}
