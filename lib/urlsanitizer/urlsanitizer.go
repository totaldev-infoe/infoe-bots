package urlsanitizer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var instagramRegex = regexp.MustCompile(`https://www\.instagram\.com/[^?]+\?igsh=[^&\s]+`)
var tiktokRegex = regexp.MustCompile(`https://(?:www\.)?tiktok\.com/[^/]+/?$`)

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := m.Content

	// Handle Instagram links
	if instagramRegex.MatchString(content) {
		cleanURL := strings.Split(content, "?")[0]
		// Delete the original message
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to remove message with tracking parameters")
			return
		}

		// Construct the message with warning and attribution
		warning := fmt.Sprintf("🔒 This message is reposted, original message from <@%s>.\nℹ️ Using igsh can potentially expose your Instagram profile when others click on it as Instagram can recommend your profile to them in the future.\n\n", m.Author.ID)
		
		// Add the clean URL and any additional content
		message := warning + cleanURL
		if len(content) > len(cleanURL) {
			// Preserve any additional message content
			extraContent := strings.TrimSpace(strings.Replace(content, instagramRegex.FindString(content), "", 1))
			if extraContent != "" {
				message = warning + extraContent + "\n" + cleanURL
			}
		}
		
		// Send the reposted message as a reply first
		_, err = s.ChannelMessageSendReply(m.ChannelID, message, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			// Fallback to regular message if reply fails
			s.ChannelMessageSend(m.ChannelID, message)
		}
		return
	}

	// Warn about potentially unsafe TikTok links
	if tiktokRegex.MatchString(content) && !strings.Contains(content, "/video/") {
		warning := "⚠️ Warning: This TikTok link format may expose your profile. Consider using the full video URL format (e.g., https://www.tiktok.com/@username/video/123456) instead."
		// Send warning as a reply to the original message
		_, err := s.ChannelMessageSendReply(m.ChannelID, warning, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			// Fallback to regular message if reply fails
			s.ChannelMessageSend(m.ChannelID, warning)
		}
		return
	}
}
