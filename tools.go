package godbot

import "github.com/bwmarrin/discordgo"

// getMainChannel sets the main channel for the bot.
func (bot *Core) getMainChannel(gID string) *discordgo.Channel {
	c := bot.links[gID]
	for _, p := range c {
		if p.ID == gID {
			return p
		}
	}
	return nil
}

// SetMainChannel sets the channel to primarily sit in.
func (bot *Core) SetMainChannel(gID, cID string) error {
	for _, p := range bot.links[gID] {
		if p.ID == cID {
			bot.channelMain = p
			return nil
		}
	}
	return ErrNotFound
}

// GetBotUser returns the bot user data.
func (bot *Core) GetBotUser() User {
	var u User
	if bot.User != nil {
		u.User = bot.User
	}
	return u
}
