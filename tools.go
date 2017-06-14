package godbot

import "github.com/bwmarrin/discordgo"

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
