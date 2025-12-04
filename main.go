package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
		handleNewPost(ctx, post, discordClient)
	}

	// Create and start monitor
	monitor := reddit.NewMonitor(
		client,
		cfg.Subreddit,
		cfg.PollInterval,
		cfg.PostLimit,
		postHandler,
	)

	log.Printf("Monitoring r/%s for new posts (via RSS)...", cfg.Subreddit)
	log.Println("Press Ctrl+C to stop")
	log.Println("")

	if err := monitor.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Monitor error: %v", err)
	}

	log.Println("Bot stopped")
}

// handleNewPost is called when a new post is detected
func handleNewPost(ctx context.Context, post reddit.Post, discordClient *discord.WebhookClient) {
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

	// Send to Discord
	if err := discordClient.SendPost(ctx, post); err != nil {
		log.Printf("Failed to send to Discord: %v", err)
	} else {
		log.Println("✓ Posted to Discord")
	}
}
