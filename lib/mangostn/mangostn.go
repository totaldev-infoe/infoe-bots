package mango

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/totaldev-infoe/infoe-bots/lib/discord"
	"log"
	"strings"
)

var (
	UserInQuestion           = "mangostn"
	RoleToAssign             = "1248052023128231937"
	BotUserID                = "1246178480471937054" // Bot user ID
	AllowedChannelsSubString = []string{
		"general", "movebot",
	}
)

func Mango(DiscordToken string) {
	discord.Call(DiscordToken, replyBackToMangoReplyWithoutRole, handleReactionAdd)
}

// Check if the channel name contains any of the allowed substrings
func channelNameContainsAllowedSubstring(channelName string) bool {
	channelName = strings.ToLower(channelName)
	for _, substr := range AllowedChannelsSubString {
		if strings.Contains(channelName, substr) {
			return true
		}
	}
	return false
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func replyBackToMangoReplyWithoutRole(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Fetch the channel information
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Println("Error getting the channel:", err)
		return
	}

	// Check if the channel name contains any of the allowed substrings
	if !channelNameContainsAllowedSubstring(channel.Name) {
		return
	}

	// Check if the user has the role 1248052023128231937
	member, err := s.GuildMember(m.GuildID, m.Author.ID)
	if err != nil {
		log.Println("Error getting the member:", err)
		return
	}

	hasRole := false
	for _, roleID := range member.Roles {
		if roleID == RoleToAssign {
			hasRole = true
			break
		}
	}

	// If the user has the role, do nothing
	if hasRole {
		return
	}

	// User shouldn't have role now

	// Check if the message is a reply
	if m.Message.ReferencedMessage != nil {

		// Get the original message
		originalMessage, err := s.ChannelMessage(m.ChannelID, m.Message.ReferencedMessage.ID)
		if err != nil {
			log.Println("Error getting the original message:", err)
			return
		}

		// Check if the original message author username is "totaldev"
		if originalMessage.Author.Username == UserInQuestion {
			// Add your custom action here
			customMessage := fmt.Sprintf(`
			Hey, %s! ,
			We've detected that you're attempting to contact our neighborhood new grad, %s. 
			
			There's a possibility you've been blocked by %s. Please react to one of his messages and if your discord client doesn't allow you to react, that means that %s blocked you. Please utilize this information however you see fit!
			
			To unsubscribe from further notifications, please react to this message!
`, m.Author.Username, originalMessage.Author.Username, originalMessage.Author.Username, originalMessage.Author.Username)

			reply := &discordgo.MessageSend{
				Content:   customMessage,
				Reference: &discordgo.MessageReference{MessageID: m.ID},
			}
			_, err := s.ChannelMessageSendComplex(m.ChannelID, reply)
			if err != nil {
				log.Println("Error sending the message:", err)
				return
			}
			return
		}
	}
}

// This function will be called when a reaction is added to a message
func handleReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Get the message to check if it's authored by the bot
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Println("Error getting the message:", err)
		return
	}

	// Check if the message is authored by the bot
	if message.Author.ID == BotUserID {
		// Add the role to the user who reacted
		err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, RoleToAssign)
		if err != nil {
			log.Println("Error adding role:", err)
		}
	}
}
