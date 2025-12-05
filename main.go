package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/block52/reddit-poker-bot/internal/claude"
	"github.com/block52/reddit-poker-bot/internal/config"
	"github.com/block52/reddit-poker-bot/internal/discord"
	"github.com/block52/reddit-poker-bot/internal/reddit"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Reddit Poker Bot starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create Reddit RSS client
	client := reddit.NewClient(cfg.UserAgent)

	// Create Discord webhook client
	discordClient := discord.NewWebhookClient(cfg.DiscordWebhookURL)
	if cfg.DiscordWebhookURL != "" {
		log.Println("Discord webhook integration enabled")
	}

	// Create Claude API client
	claudeClient := claude.NewClient(cfg.AnthropicAPIKey)
	if cfg.AnthropicAPIKey != "" {
		log.Println("Claude AI integration enabled for response generation")
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Create post handler that sends to Discord
	postHandler := func(post reddit.Post) {
		handleNewPost(ctx, post, discordClient, claudeClient)
	}

	// Start a monitor for each subreddit
	var wg sync.WaitGroup
	for _, subreddit := range cfg.Subreddits {
		wg.Add(1)
		go func(sub string) {
			defer wg.Done()

			monitor := reddit.NewMonitor(
				client,
				sub,
				cfg.PollInterval,
				cfg.PostLimit,
				postHandler,
			)

			log.Printf("Starting monitor for r/%s", sub)

			if err := monitor.Start(ctx); err != nil && err != context.Canceled {
				log.Printf("Monitor error for r/%s: %v", sub, err)
			}
		}(subreddit)
	}

	log.Printf("Monitoring %d subreddit(s): %s", len(cfg.Subreddits), strings.Join(cfg.Subreddits, ", "))
	log.Println("Press Ctrl+C to stop")
	log.Println("")

	wg.Wait()

	log.Println("Bot stopped")
}

// handleNewPost is called when a new post is detected
func handleNewPost(ctx context.Context, post reddit.Post, discordClient *discord.WebhookClient, claudeClient *claude.Client) {
	age := time.Since(post.CreatedAt).Round(time.Second)

	// Print to terminal
	fmt.Println("")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("NEW POST in r/%s\n", post.Subreddit)
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("Title:  %s\n", post.Title)
	fmt.Printf("Author: u/%s\n", post.Author)
	fmt.Printf("Posted: %s ago\n", age)
	if len(post.Categories) > 0 {
		fmt.Printf("Tags:   %s\n", strings.Join(post.Categories, ", "))
	}
	fmt.Printf("Link:   %s\n", post.GetFullPermalink())
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("")

	// Terminal bell for notification
	fmt.Print("\a")

	// Generate AI responses if Claude is enabled
	if claudeClient != nil {
		log.Println("Generating response suggestions with Claude AI...")
		responses, err := claudeClient.GenerateResponses(ctx, post)
		if err != nil {
			log.Printf("Failed to generate responses: %v", err)
			// Fall back to posting without responses
			if err := discordClient.SendPost(ctx, post); err != nil {
				log.Printf("Failed to send to Discord: %v", err)
			} else {
				log.Println("✓ Posted to Discord")
			}
			return
		}

		log.Println("✓ Generated response suggestions")

		// Send to Discord with responses in a thread
		if err := discordClient.SendPostWithResponses(ctx, post, responses.Witty, responses.Formal, responses.Block52); err != nil {
			log.Printf("Failed to send to Discord: %v", err)
		} else {
			log.Println("✓ Posted to Discord with response suggestions")
		}
	} else {
		// No Claude integration, just send the post
		if err := discordClient.SendPost(ctx, post); err != nil {
			log.Printf("Failed to send to Discord: %v", err)
		} else {
			log.Println("✓ Posted to Discord")
		}
	}
}
