# Reddit Poker Bot

A Go bot that monitors the r/poker subreddit for new posts in real-time, allowing you to respond quickly to get karma and visibility.

## Features

- Real-time monitoring of r/poker (or any subreddit)
- OAuth2 authentication with Reddit API
- Configurable poll interval and post limits
- Terminal bell notification for new posts
- Shows post title, author, flair, and content preview

## Prerequisites

- Go 1.21 or later
- A Reddit account
- Reddit API credentials (see setup below)

## Setup

### 1. Create a Reddit App

1. Go to https://www.reddit.com/prefs/apps
2. Click "create another app..."
3. Fill in the details:
   - **name**: Your bot name (e.g., "PokerBot")
   - **type**: Select "script"
   - **description**: Optional
   - **redirect uri**: `http://localhost:8080` (not used but required)
4. Click "create app"
5. Note your **client ID** (under the app name) and **client secret**

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```bash
REDDIT_CLIENT_ID=your_client_id
REDDIT_CLIENT_SECRET=your_client_secret
REDDIT_USERNAME=your_reddit_username
REDDIT_PASSWORD=your_reddit_password
REDDIT_USER_AGENT=RedditPokerBot/1.0 by u/your_username
```

### 3. Build and Run

```bash
# Build the binary
go build -o reddit-poker-bot .

# Run with environment variables
source .env && ./reddit-poker-bot
```

Or run directly with Go:

```bash
source .env && go run .
```

## Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `REDDIT_CLIENT_ID` | (required) | Your Reddit app client ID |
| `REDDIT_CLIENT_SECRET` | (required) | Your Reddit app client secret |
| `REDDIT_USERNAME` | (required) | Your Reddit username |
| `REDDIT_PASSWORD` | (required) | Your Reddit password |
| `REDDIT_USER_AGENT` | `RedditPokerBot/1.0` | User agent string |
| `REDDIT_SUBREDDIT` | `poker` | Subreddit to monitor |
| `POLL_INTERVAL_SECONDS` | `30` | Seconds between API polls |
| `POST_LIMIT` | `25` | Number of posts to fetch per poll |

## Rate Limits

Reddit's API has rate limits:
- OAuth apps: ~60 requests per minute
- The default 30-second poll interval is well within limits
- Avoid setting `POLL_INTERVAL_SECONDS` below 10

## Output Example

```
═══════════════════════════════════════════════════════════════
🆕 NEW POST in r/poker
═══════════════════════════════════════════════════════════════
📝 Title: Just hit my first Royal Flush!
👤 Author: u/poker_player123
⏰ Posted: 12s ago
🏷️  Flair: Hand Analysis
🔗 Link: https://reddit.com/r/poker/comments/abc123/...
📄 Text: Was playing 1/2 at the local casino when...
═══════════════════════════════════════════════════════════════
```

## Extending the Bot

The `handleNewPost` function in `main.go` can be extended to:
- Send desktop notifications
- Post to Discord/Slack webhooks
- Log to a database
- Auto-generate response drafts
- Filter by keywords or flair

## License

MIT
