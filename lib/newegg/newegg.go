package newegg

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly/v2"
	"github.com/totaldev-infoe/infoe-bots/lib/discord"
)

var (
	// Browser versions
	chromeVersions  = []string{"120.0.0.0", "121.0.0.0", "122.0.0.0"}
	safariVersions  = []string{"17.2", "17.1", "16.6"}
	firefoxVersions = []string{"121.0", "122.0", "123.0"}

	// Operating systems and their versions
	osVersions = map[string][]string{
		"Windows":   {"10.0", "11.0"},
		"Macintosh": {"10_15_7", "11_0", "12_0", "13_0"},
		"X11":       {"Linux x86_64", "Ubuntu; Linux x86_64"},
	}

	// Common referrers
	referrers = []string{
		"https://www.google.com/",
		"https://www.bing.com/",
		"https://duckduckgo.com/",
		"https://www.newegg.com/",
		"https://www.reddit.com/",
	}

	// Accept languages with weights
	acceptLanguages = []string{
		"en-US,en;q=0.9",
		"en-GB,en;q=0.9",
		"en-CA,en;q=0.9,fr-CA;q=0.8",
		"en-AU,en;q=0.9",
	}
)

// Function to get a random item from a slice
func randomChoice[T any](items []T) T {
	return items[rand.Intn(len(items))]
}

// Function to get a random key-value pair from a map
func randomMapChoice[K comparable, V any](m map[K][]V) (K, V) {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	key := keys[rand.Intn(len(keys))]
	value := m[key][rand.Intn(len(m[key]))]
	return key, value
}

// Function to generate a random User-Agent
func randomUserAgent() string {
	rand.Seed(time.Now().UnixNano())

	// Randomly choose browser type and generate appropriate User-Agent
	browserType := rand.Intn(3)
	switch browserType {
	case 0: // Chrome
		os, osVer := randomMapChoice(osVersions)
		version := randomChoice(chromeVersions)
		return fmt.Sprintf("Mozilla/5.0 (%s; %s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", os, osVer, version)
	case 1: // Firefox
		os, osVer := randomMapChoice(osVersions)
		version := randomChoice(firefoxVersions)
		return fmt.Sprintf("Mozilla/5.0 (%s; %s; rv:%s) Gecko/20100101 Firefox/%s", os, osVer, version, version)
	default: // Safari
		version := randomChoice(safariVersions)
		return fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15", version)
	}
}

// RTX card models and their search URLs
var (
	rtxCards = map[string]string{
		"!5070_newegg": "RTX+5070",
		"!5080_newegg": "RTX+5080",
		"!5090_newegg": "RTX+5090",
	}

	// Regular expression for the generic search command
	genericSearchRegex = regexp.MustCompile(`^!newegg\s+"([^"]+)"$`)
)

// sanitizeSearchTerm cleans and validates the search term
func sanitizeSearchTerm(term string) (string, error) {
	// Remove any potentially dangerous characters
	// Only allow alphanumeric, spaces, and basic punctuation
	safeRegex := regexp.MustCompile(`[^a-zA-Z0-9\s\-_+.,]`)
	sanitized := safeRegex.ReplaceAllString(term, "")

	// Trim spaces and check length
	sanitized = strings.TrimSpace(sanitized)
	if len(sanitized) == 0 {
		return "", fmt.Errorf("search term is empty after sanitization")
	}
	if len(sanitized) > 100 {
		return "", fmt.Errorf("search term is too long (max 100 characters)")
	}

	// URL encode the term
	sanitized = url.QueryEscape(sanitized)
	return sanitized, nil
}

func LookupRTX(DiscordToken string) {
	discord.Call(DiscordToken, messageNeweggLookup)
}

func messageNeweggLookup(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Variables for search parameters
	var searchQuery string
	var displayName string

	// Check if it's a generic search command
	if matches := genericSearchRegex.FindStringSubmatch(m.Content); len(matches) == 2 {
		// Sanitize the search term
		sanitized, err := sanitizeSearchTerm(matches[1])
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Invalid search term: %v", err))
			return
		}
		searchQuery = sanitized
		displayName = matches[1] // Use original term for display
	} else {
		// Check if it's an RTX card lookup
		cardQuery, ok := rtxCards[m.Content]
		if !ok {
			return
		}
		searchQuery = cardQuery
		displayName = m.Content[1:5] // e.g., "5090" from "!5090_newegg"
	}

	// Send initial response
	initialMsg, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔍 Searching Newegg for %s...", displayName))
	if err != nil {
		fmt.Printf("Error sending initial message: %v\n", err)
		return
	}

	// Initialize variables for retry logic
	maxRetries := 5
	baseDelay := 5 * time.Second
	var lastError error
	var successfulCollector *colly.Collector
	var listings strings.Builder

	// Retry loop
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay with jitter
			delay := baseDelay * time.Duration(attempt)
			jitter := time.Duration(rand.Float64() * float64(delay))
			delay = delay + jitter

			// Update message to show retry attempt
			s.ChannelMessageEdit(m.ChannelID, initialMsg.ID,
				fmt.Sprintf("🔍 Searching Newegg for %s... (Attempt %d/%d)", displayName, attempt+1, maxRetries))

			// Wait before retrying
			time.Sleep(delay)
		}

		// Reset listings for each attempt
		listings.Reset()
		listings.WriteString(fmt.Sprintf("📊 **%s Listings on Newegg:**\n\n", displayName))

		// Create a new collector for each attempt
		c := colly.NewCollector(
			colly.AllowedDomains("www.newegg.com"),
			colly.UserAgent(randomUserAgent()),
		)

		// Configure collector
		c.WithTransport(&http.Transport{
			DisableKeepAlives: true, // Don't reuse connections
		})

		// Randomize delay between requests
		c.SetRequestTimeout(20 * time.Second)
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*newegg.com*",
			RandomDelay: 2 * time.Second,
			Parallelism: 1,
		})

		// Set headers
		c.OnRequest(func(r *colly.Request) {
			// Randomize all headers
			r.Headers.Set("User-Agent", randomUserAgent())
			r.Headers.Set("Referer", randomChoice(referrers))
			r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
			r.Headers.Set("Accept-Language", randomChoice(acceptLanguages))
			r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
			r.Headers.Set("Connection", "keep-alive")
			r.Headers.Set("Upgrade-Insecure-Requests", "1")
			r.Headers.Set("Sec-Fetch-Dest", "document")
			r.Headers.Set("Sec-Fetch-Mode", "navigate")
			r.Headers.Set("Sec-Fetch-Site", "none")
			r.Headers.Set("Sec-Fetch-User", "?1")
			r.Headers.Set("DNT", "1")
			r.Headers.Set("Cache-Control", "max-age=0")
		})

		// Scrape product listings
		c.OnHTML(".item-cell", func(e *colly.HTMLElement) {
			title := e.ChildText(".item-title")
			price := e.ChildText(".price-current")
			link := e.ChildAttr(".item-title", "href")
			outOfStockText := e.ChildText(".item-promo")

			// Clean price text
			price = strings.TrimSpace(strings.ReplaceAll(price, "\n", ""))

			// Determine stock status
			inStock := !strings.Contains(strings.ToUpper(outOfStockText), "OUT OF STOCK")
			stockStatus := "✅ In Stock"
			if !inStock {
				stockStatus = "❌ Out of Stock"
			}

			// Add listing to message
			listings.WriteString(fmt.Sprintf("💻 **%s**\n", title))
			listings.WriteString(fmt.Sprintf("💰 Price: %s\n", price))
			listings.WriteString(fmt.Sprintf("📦 Status: %s\n", stockStatus))
			listings.WriteString(fmt.Sprintf("🔗 Link: %s\n", link))
			listings.WriteString("--------------------------------\n")
		})

		// Configure collector
		c.WithTransport(&http.Transport{
			DisableKeepAlives: true,
		})

		// Randomize delay between requests
		c.SetRequestTimeout(20 * time.Second)
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*newegg.com*",
			RandomDelay: 2 * time.Second,
			Parallelism: 1,
		})

		// Set headers
		c.OnRequest(func(r *colly.Request) {
			// Randomize all headers for each attempt
			r.Headers.Set("User-Agent", randomUserAgent())
			r.Headers.Set("Referer", randomChoice(referrers))
			r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
			r.Headers.Set("Accept-Language", randomChoice(acceptLanguages))
			r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
			r.Headers.Set("Connection", "keep-alive")
			r.Headers.Set("Upgrade-Insecure-Requests", "1")
			r.Headers.Set("Sec-Fetch-Dest", "document")
			r.Headers.Set("Sec-Fetch-Mode", "navigate")
			r.Headers.Set("Sec-Fetch-Site", "none")
			r.Headers.Set("Sec-Fetch-User", "?1")
			r.Headers.Set("DNT", "1")
			r.Headers.Set("Cache-Control", "max-age=0")
		})

		// Track if we found any listings
		foundListings := false

		// Error handling
		c.OnError(func(r *colly.Response, err error) {
			lastError = err
		})

		// Scrape product listings
		c.OnHTML(".item-cell", func(e *colly.HTMLElement) {
			foundListings = true
			title := e.ChildText(".item-title")
			price := e.ChildText(".price-current")
			link := e.ChildAttr(".item-title", "href")
			outOfStockText := e.ChildText(".item-promo")

			// Clean price text
			price = strings.TrimSpace(strings.ReplaceAll(price, "\n", ""))

			// Determine stock status
			inStock := !strings.Contains(strings.ToUpper(outOfStockText), "OUT OF STOCK")
			stockStatus := "✅ In Stock"
			if !inStock {
				stockStatus = "❌ Out of Stock"
			}

			// Add listing to message
			listings.WriteString(fmt.Sprintf("💻 **%s**\n", title))
			listings.WriteString(fmt.Sprintf("💰 Price: %s\n", price))
			listings.WriteString(fmt.Sprintf("📦 Status: %s\n", stockStatus))
			listings.WriteString(fmt.Sprintf("🔗 Link: %s\n", link))
			listings.WriteString("--------------------------------\n")
		})

		// Visit Newegg search page
		url := fmt.Sprintf("https://www.newegg.com/p/pl?d=%s", searchQuery)
		err = c.Visit(url)

		// If we got listings or no error, save the collector and break
		if err == nil && foundListings {
			successfulCollector = c
			break
		}

		// If this was the last attempt, show the error
		if attempt == maxRetries-1 {
			s.ChannelMessageEdit(m.ChannelID, initialMsg.ID,
				fmt.Sprintf("❌ Failed to access Newegg search for %s after %d attempts. Last error: %v", displayName, maxRetries, lastError))
			return
		}
	}

	// If we have a successful collector, create thread and post results
	if successfulCollector != nil {
		// Update initial message
		s.ChannelMessageEdit(m.ChannelID, initialMsg.ID, "Found the following! Details in :thread:")

		// Create a thread for the listings
		thread, err := s.MessageThreadStart(m.ChannelID, initialMsg.ID, fmt.Sprintf("%s Search Results", displayName), 60)
		if err != nil {
			s.ChannelMessageEdit(m.ChannelID, initialMsg.ID, "❌ Error creating thread for listings")
			fmt.Printf("Error creating thread: %v\n", err)
			return
		}

		// Send the listings in the thread
		message := listings.String()
		if message == "📊 **RTX 5090 Listings on Newegg:**\n\n" {
			s.ChannelMessageSend(thread.ID, "No RTX 5090 listings found on Newegg.")
		} else {
			// Split message if it's too long for Discord
			const maxLength = 2000
			for len(message) > 0 {
				if len(message) <= maxLength {
					s.ChannelMessageSend(thread.ID, message)
					break
				}
				// Find last newline before maxLength
				splitIndex := strings.LastIndex(message[:maxLength], "\n")
				if splitIndex == -1 {
					splitIndex = maxLength
				}
				s.ChannelMessageSend(thread.ID, message[:splitIndex])
				message = message[splitIndex:]
			}
		}
	}
}
