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

	if cl.Channel.Type != 0 {
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
	var username = name
	s := bot.Session
	if gID == "" {
		return ErrBadGuild
	}

	if append {
		username = fmt.Sprintf("%s %s", bot.User.Username, name)
	}

	err := s.GuildMemberNickname(gID, "@me", username)
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

// GuildToSlice gives a list of current accessible Guild
// infoType defaults to Guild ID, if infoType is set to "name" it will return names.
// otherwise it will return the IDs.
func (bot *Core) GuildToSlice(infoType string) (guilds []string) {
	for _, g := range bot.Guilds {
		if strings.ToLower(infoType) == "name" {
			guilds = append(guilds, g.Name)
		} else {
			guilds = append(guilds, g.ID)
		}
	}
	return
}

// ChannelToSlice grants a list of currently accessible channels for a particular guild.
// By default, it will return a list of IDs. Giving "name" for infoType will return a list
// of channel names.
func (bot *Core) ChannelToSlice(guildID, infoType string) (channels []string) {
	if guildID == "" {
		return
	}

	// Make sure we have that guild ID in our links.
	if _, ok := bot.Links[guildID]; !ok {
		return
	}

	for _, c := range bot.Links[guildID] {
		if strings.ToLower(infoType) == "name" {
			channels = append(channels, c.Name)
		} else {
			channels = append(channels, c.ID)
		}
	}
	return
}

// ChannelMemoryDelete will remove a channel from the array of channels in memory.
func (bot *Core) ChannelMemoryDelete(channel *discordgo.Channel) {
	bot.muUpdate.Lock()
	defer bot.muUpdate.Unlock()

	if len(bot.Channels) <= 1 {
		bot.Channels = nil
	} else {
		// Search the list of ALL channels.
		for n, c := range bot.Channels {
			// Find our channel...
			if c.ID == channel.ID {
				bot.Channels[n] = bot.Channels[len(bot.Channels)-1]
				bot.Channels[len(bot.Channels)-1] = nil
				bot.Channels = bot.Channels[:len(bot.Channels)-1]
				break
			}
		}
	}

	if len(bot.Links[channel.GuildID]) <= 1 {
		bot.Links[channel.GuildID] = nil
	} else {
		// Search links.
		for n, c := range bot.Links[channel.GuildID] {
			// Find our channel...
			if c.ID == channel.ID {
				bot.Links[channel.GuildID][n] = bot.Links[channel.GuildID][len(bot.Links[channel.GuildID])-1]
				bot.Links[channel.GuildID][len(bot.Links[channel.GuildID])-1] = nil
				bot.Links[channel.GuildID] = bot.Links[channel.GuildID][:len(bot.Links[channel.GuildID])-1]
				break
			}
		}
	}

	return
}

// ChannelMemoryAdd will Add/Replace a channels structure in memory.
func (bot *Core) ChannelMemoryAdd(channel *discordgo.Channel) {
	bot.muUpdate.Lock()
	defer bot.muUpdate.Unlock()

	// Search the entire channels list, and modify or append.
	var exists bool
	for n, c := range bot.Channels {
		if c.ID == channel.ID {
			exists = true
			bot.Channels[n] = channel
			break
		}
	}

	// Append it since it wasn't found.
	if !exists {
		bot.Channels = append(bot.Channels, channel)
	}

	exists = false
	// Search our guild link and modify it.
	for n, c := range bot.Links[channel.GuildID] {
		if c.ID == channel.ID {
			exists = true
			bot.Links[channel.GuildID][n] = channel
			break
		}
	}

	// Wasn't discovered in our links, append it.
	if !exists {
		bot.Links[channel.GuildID] = append(bot.Links[channel.GuildID], channel)
	}

	return
}

// GetGuildMembers returns EVERY user in a guild.
func (bot *Core) GetGuildMembers(guildID string, userAmount int) ([]*discordgo.Member, error) {
	s := bot.Session
	var usersAll []*discordgo.Member

	var pullAmount int
	var afterID string

	for userAmount > 0 {
		pullAmount = userAmount
		if userAmount > 1000 {
			pullAmount = 1000
		}

		users, err := s.GuildMembers(guildID, afterID, pullAmount)
		if err != nil {
			return nil, err
		}

		// Add it to our list.
		for _, u := range users {
			afterID = u.User.ID
			usersAll = append(usersAll, u)
		}

		userAmount -= pullAmount
	}

	return usersAll, nil
}
