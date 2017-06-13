package godbot

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// Standard error messages
var (
	ErrNilGuilds   = errors.New("bot.guilds is nil")
	ErrNilChannels = errors.New("bot.channels is nil")
	ErrNilLinks    = errors.New("bot.links is nil")
	ErrNotFound    = errors.New("not found")
)

// Codes for types of Connections.
const (
	bwChannel = 1 << iota
	bwGuild
	bwLinks
	bwPrivate
)

// GetConnections returns the current connection structure.
func (bot *Core) GetConnections() (*Connections, error) {
	err := bot.UpdateConnections()
	if err != nil {
		return nil, err
	}

	return &Connections{Links: bot.links, Guilds: bot.guilds, Channels: bot.channels}, nil

}

// UpdateConnections is a public wrapper queries discord for all information needed by bot.
func (bot *Core) UpdateConnections() error {
	toUpdate := bwChannel | bwGuild | bwLinks | bwPrivate
	return bot.updateConnections(toUpdate)
}

// updateConnections queries discord for specified information.
func (bot *Core) updateConnections(toUpdate int) error {
	var err error
	bot.muUpdate.Lock()
	defer bot.muUpdate.Unlock()

	for toUpdate > 0 {
		switch {
		case toUpdate&bwGuild == bwGuild:
			toUpdate = toUpdate ^ bwGuild
			err = bot.queryGuilds()
		case toUpdate&bwChannel == bwChannel:
			toUpdate = toUpdate ^ bwChannel
			err = bot.queryChannels()
		case toUpdate&bwLinks == bwLinks:
			toUpdate = toUpdate ^ bwLinks
			err = bot.queryLinks()
		case toUpdate&bwPrivate == bwPrivate:
			toUpdate = toUpdate ^ bwPrivate
			err = bot.queryPrivate()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// queryLinks creates the map with [guild] -> [channels]
func (bot *Core) queryLinks() error {
	s := bot.session

	if bot.guilds == nil {
		return ErrNilGuilds
	} else if bot.links == nil {
		bot.links = make(map[string][]*discordgo.Channel)
	}

	for _, g := range bot.guilds {
		if _, ok := bot.links[g.ID]; ok == false {
			c, err := s.GuildChannels(g.ID)
			if err != nil {
				return err
			}

			bot.links[g.ID] = c
		}
	}

	return nil
}

// queryGuilds pulls all guilds associated with the bot.
func (bot *Core) queryGuilds() error {
	var in bool
	var err error
	s := bot.session

	guilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		return err
	}

	for _, g := range guilds {
		guild, err := s.Guild(g.ID)
		if err != nil {
			return err
		}
		in = false
		if bot.guilds != nil {
			for _, t := range bot.guilds {
				if t.ID == guild.ID {
					in = true
					break
				}
			}
		}
		if in == false {
			bot.guilds = append(bot.guilds, guild)
		}
	}

	return nil
}

// queryChannels just updates the core.channels slices with current guilds.
func (bot *Core) queryChannels() error {
	var in bool
	s := bot.session

	if bot.guilds == nil {
		return ErrNilGuilds
	}

	for _, g := range bot.guilds {
		channels, err := s.GuildChannels(g.ID)
		if err != nil {
			return err
		}

		for _, c := range channels {
			in = false
			if bot.channels != nil {
				for _, t := range bot.channels {
					if c.ID == t.ID {
						in = true
						break
					}
				}
			}
			if in == false {
				bot.channels = append(bot.channels, c)
			}
		}
	}
	return nil
}

func (bot *Core) queryPrivate() error {
	var in bool
	s := bot.session

	private, err := s.UserChannels()
	if err != nil {
		return err
	}

	if len(private) > 0 {
		for _, p := range private {
			in = false
			if bot.private != nil {
				for _, t := range bot.private {
					if t.ID == p.ID {
						in = true
						break
					}
				}
			}
			if in == false {
				bot.private = append(bot.private, p)
			}
		}
	}
	return nil
}

// ConnectionsReset sets the defaults for the bots connections.
func (bot *Core) ConnectionsReset() error {
	err := bot.UpdateConnections()
	if err != nil {
		return err
	}

	bot.guildMain = bot.guilds[0]
	bot.channelMain = bot.getMainChannel(bot.guildMain.ID)

	return nil
}
