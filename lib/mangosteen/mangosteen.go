package badMango

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/totaldev-infoe/infoe-bots/lib/discord"
	"log"
	"strings"
	"time"
)

var (
	MangoID 				 = "533545308374827029"
	AllowedChannelsSubString = []string{
		"general", "movebot",
	}
	LastDeleteTimestamp time.Time // Store timestamp of last message delete from mango
)

const OneDay = 24 * time.Hour

func BadMango(DiscordToken string) {
	discord.Call(DiscordToken, mangoDeleteHandler)
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

// mangoDeleteHandler handles the message delete event by mango
func mangoDeleteHandler(s *discordgo.Session, m *discordgo.MessageDelete) {
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore all users that are not mango
	// TODO: Test if it is m.Message.Author.ID or if m.Author.ID works
	if m.Author.ID != MangoID {
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

	// Check if the user has triggered the bot in the last 24 hours
	if lastTrigger, exists := UserCache[m.Author.ID]; exists && time.Since(lastTrigger) < OneDay {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log.Printf("error getting channel: %v", err)
		return
	}

		// Update the timestamp of the last message deletion by Mango
		LastDeleteTimestamp = time.Now()


	message := fmt.Sprintf("Bad Mango! Stop deleting messages in channel: %s", channel.Name)

	_, err = s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		log.Printf("error sending message: %v", err)
	}
}

// mangoCreateHandler handles the message create event by mango
func mangoCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore all users that are not mango
	if m.Author.ID != MangoID {
		return
	}

	// Check if the current time is within an hour of the LastDeleteTimestamp
	if time.Since(LastDeleteTimestamp) <= time.Hour {
		// Duplicate the message in the channel
		_, err := s.ChannelMessageSend(m.ChannelID, m.Content)
		if err != nil {
			log.Printf("error duplicating message: %v", err)
		}
	}
}

