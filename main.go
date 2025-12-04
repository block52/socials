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

	// Create and start monitor
	monitor := reddit.NewMonitor(
		client,
		cfg.Subreddit,
		cfg.PollInterval,
		cfg.PostLimit,
		handleNewPost,
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
func handleNewPost(post reddit.Post) {
	age := time.Since(post.CreatedAt).Round(time.Second)

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
}
