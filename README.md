# Reddit Poker Bot

A Go bot that monitors the r/poker subreddit for new posts in real-time using RSS feeds. No API keys required.

## Features

- Real-time monitoring of r/poker (or any subreddit)
- No authentication required (uses public RSS feeds)
- Configurable poll interval and post limits
- Terminal bell notification for new posts
- Shows post title, author, and link

## Prerequisites

- Go 1.21 or later

## Quick Start

```bash
# Clone and build
go build -o reddit-poker-bot .

# Run with defaults (monitors r/poker)
./reddit-poker-bot

# Or run directly
go run .
```

That's it! No API keys, no configuration required.

## Configuration (Optional)

You can customize behavior with environment variables:

```bash
# Monitor a different subreddit
SUBREDDIT=wallstreetbets ./reddit-poker-bot

# Faster polling (every 15 seconds)
POLL_INTERVAL_SECONDS=15 ./reddit-poker-bot

# Combine options
SUBREDDIT=poker POLL_INTERVAL_SECONDS=20 ./reddit-poker-bot
```

Or use an env file:

```bash
cp .env.example .env
# Edit .env as needed
source .env && ./reddit-poker-bot
```

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `SUBREDDIT` | `poker` | Subreddit to monitor |
| `POLL_INTERVAL_SECONDS` | `30` | Seconds between RSS polls |
| `POST_LIMIT` | `25` | Number of posts to fetch per poll |
| `USER_AGENT` | `RedditPokerBot/1.0` | User agent for HTTP requests |

## Output Example

```
═══════════════════════════════════════════════════════════════
NEW POST in r/poker
═══════════════════════════════════════════════════════════════
Title:  Just hit my first Royal Flush!
Author: u/poker_player123
Posted: 12s ago
Tags:   poker
Link:   https://www.reddit.com/r/poker/comments/abc123/...
═══════════════════════════════════════════════════════════════
```

## How It Works

The bot uses Reddit's public RSS feeds (`https://www.reddit.com/r/{subreddit}/new.rss`) which don't require authentication. It polls the feed at regular intervals and notifies you when new posts appear.

## Limitations

- RSS feeds may have a slight delay compared to the API (usually under a minute)
- No access to post scores, comment counts, or other metadata not in RSS
- Rate limiting may apply for very aggressive polling

## Extending the Bot

The `handleNewPost` function in `main.go` can be extended to:
- Send desktop notifications
- Post to Discord/Slack webhooks
- Log to a database
- Filter by keywords in the title
- Open the post URL in your browser automatically

## License

MIT
