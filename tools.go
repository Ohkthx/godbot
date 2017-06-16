package godbot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// Constants for locked channels.
var (
	ErrChannelNotLocked = errors.New("channel is not locked")
	ErrChannelLocked    = errors.New("channel is locked")
)

// getMainChannel sets the main channel for the bot.
func (bot *Core) getMainChannel(gID string) *discordgo.Channel {
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

	cl.Session = s
	cl.Channel = bot.GetChannel(cID)
	cl.Guild = bot.GetGuild(cl.Channel.GuildID)
	for _, p := range cl.Channel.PermissionOverwrites {
		r, err := s.State.Role(cl.Guild.ID, p.ID)
		if err != nil {
			return nil, err
		}

		if r.Name == "@everyone" {
			cl.Role = r
			cl.Allow = p.Allow
			cl.Deny = p.Deny
			cl.Permissions = r.Permissions
			cl.Type = p.Type
		}
	}

	return cl, nil

}

// ChannelLock will lock a channel preventing @everyone typing.
func (cl *ChannelLock) ChannelLock() error {
	//var timeoutRole, everyoneRole, everyoneRoleBak *discordgo.Role
	// Get current Roles permissions.
	if cl.Locked {
		return nil
	}

	s := cl.Session
	nA := cl.Allow
	if cl.Allow&2048 == 2048 {
		nA = cl.Allow ^ 2048
	}

	err := s.ChannelPermissionSet(cl.Channel.ID, cl.Role.ID, cl.Type, nA, cl.Deny|2048)
	if err != nil {
		return err
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
	err := s.ChannelPermissionSet(cl.Channel.ID, cl.Role.ID, cl.Type, cl.Allow, cl.Deny)
	if err != nil {
		return err
	}
	cl.Locked = false
	return nil
}
