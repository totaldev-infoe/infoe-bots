package llmpromoter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/totaldev-infoe/infoe-bots/lib/sqlcache"
)

const (
	// Feature name for opt-out tracking
	featureName = "llm_promoter"
)

var (
	// Regex to match complaints about hosted LLMs
	llmComplaintRegex = regexp.MustCompile(`(?i)(claude|chatgpt|gpt-4|gpt4|windsurf|cursor|anthropic|openai|bard|gemini).*?(down|slow|unavailable|not working|broken|tokens|quota|limit)`)
)

// Initialize the SQLite cache
func init() {
	sqlcache.Initialize()
}

// HandleReactionAdd handles the reaction add event to opt out users
func HandleReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from the bot itself
	if r.UserID == s.State.User.ID {
		return
	}

	// Check if the reaction is to one of our bot's messages
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		return
	}

	// Only process if it's our bot's message and contains the opt-out text
	if message.Author.ID == s.State.User.ID && strings.Contains(strings.ToLower(message.Content), "react to this message to opt out") {
		
		// Add user to opt-out list
		err := sqlcache.AddOptOut(r.UserID, featureName)
		if err != nil {
			return
		}
		
		// Acknowledge the opt-out with a reaction and a direct message
		s.MessageReactionAdd(r.ChannelID, r.MessageID, "👍")
		
		// Send a DM to confirm opt-out
		channel, err := s.UserChannelCreate(r.UserID)
		if err == nil {
			s.ChannelMessageSend(channel.ID, "You've been opted out of LLM promotion messages. You won't receive these suggestions anymore.")
		}
		

	}
}

// HandleMessage processes messages to detect complaints about hosted LLMs
func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if user has opted out
	optedOut, err := sqlcache.HasOptedOut(m.Author.ID, featureName)
	if err != nil {
		return
	}
	if optedOut {
		return
	}

	content := m.Content

	// Check if message contains complaints about hosted LLMs
	if llmComplaintRegex.MatchString(content) {
		// Construct the response message
		message := fmt.Sprintf("Hey <@%s>, I noticed you might be having issues with a hosted LLM service. "+
			"Have you considered trying local LLMs? Here are some great options:\n\n"+
			"• **Qwen** - Fast and efficient, works well on consumer hardware\n"+
			"• **Llama 3** - Open source and high quality for local deployment\n"+
			"• **Mistral** - Excellent performance with lower resource requirements\n"+
			"• **Phi-3** - Microsoft's compact but powerful model\n\n"+
			"Local LLMs give you more control, privacy, and aren't subject to service outages or token limits.\n\n"+
			"_(React to this message to opt out of these suggestions)_", m.Author.ID)

		// Send the message as a reply
		_, err := s.ChannelMessageSendReply(m.ChannelID, message, &discordgo.MessageReference{
			MessageID: m.ID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		})
		if err != nil {
			// Fallback to regular message if reply fails
			_, err = s.ChannelMessageSend(m.ChannelID, message)
			if err != nil {
				return
			}
		}
	}
}
