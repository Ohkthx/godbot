package godbot

import "github.com/bwmarrin/discordgo"

func (bot *Core) readyHandler(s *discordgo.Session, event *discordgo.Ready) {
	bot.Lock()
	defer bot.Unlock()

	var err error
	bot.User, err = s.User("@me")
	if err != nil {
		bot.errorlog(err)
		bot.ready <- err.Error()
		return
	}

	err = bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		bot.ready <- err.Error()
		return
	}

	bot.GuildMain = bot.Guilds[0]
	bot.ChannelMain = bot.GetMainChannel(bot.GuildMain.ID)

	if bot.Game != "" {
		err = s.UpdateStatus(0, bot.Game)
		if err != nil {
			bot.errorlog(err)
			bot.ready <- err.Error()
			return
		}
	}

	bot.ready <- "ok"
}

// MessageCreateHandler assigns a function to handle messages.
func (bot *Core) MessageCreateHandler(msgHandler func(*discordgo.Session, *discordgo.MessageCreate)) {
	bot.mch = msgHandler
}

// MessageUpdateHandler assigns a function to handle messages.
func (bot *Core) MessageUpdateHandler(msgHandler func(*discordgo.Session, *discordgo.MessageUpdate)) {
	bot.muh = msgHandler
}

// GuildMemberAddHandler assigns a function to deal with newly joining users.
func (bot *Core) GuildMemberAddHandler(userHandler func(*discordgo.Session, *discordgo.GuildMemberAdd)) {
	bot.gmah = userHandler
}

// GuildMemberUpdateHandler assigns a function to deal with updating users.
func (bot *Core) GuildMemberUpdateHandler(userHandler func(*discordgo.Session, *discordgo.GuildMemberUpdate)) {
	bot.gmuh = userHandler
}

// GuildMemberRemoveHandler assigns a function to deal with leaving users.
func (bot *Core) GuildMemberRemoveHandler(userHandler func(*discordgo.Session, *discordgo.GuildMemberRemove)) {
	bot.gmrh = userHandler
}

// GuildCreateHandler assigns a function to deal with newly create guilds.
func (bot *Core) GuildCreateHandler(createHandler func(*discordgo.Session, *discordgo.GuildCreate)) {
	bot.gah = createHandler
}

// GuildRoleUpdateHandler will process new updates for guild roles.
func (bot *Core) GuildRoleUpdateHandler(updateHandler func(*discordgo.Session, *discordgo.GuildRoleUpdate)) {
	bot.gruh = updateHandler
}

// GuildRoleDeleteHandler checks if guild roles are removed.
func (bot *Core) GuildRoleDeleteHandler(deleteHandler func(*discordgo.Session, *discordgo.GuildRoleDelete)) {
	bot.grdh = deleteHandler
}

func (bot *Core) channelCreated(s *discordgo.Session, cc *discordgo.ChannelCreate) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}

func (bot *Core) channelDeleted(s *discordgo.Session, cd *discordgo.ChannelDelete) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}

func (bot *Core) channelUpdated(s *discordgo.Session, cu *discordgo.ChannelUpdate) {
	err := bot.UpdateConnections()
	if err != nil {
		bot.errorlog(err)
		return
	}
}
