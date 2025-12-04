package reddit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// Client fetches posts from Reddit via RSS
type Client struct {
	parser    *gofeed.Parser
	userAgent string
}

// Post represents a Reddit post
type Post struct {
	ID          string
	Title       string
	Author      string
	Subreddit   string
	URL         string
	Permalink   string
	Content     string
	CreatedAt   time.Time
	Categories  []string
}

// NewClient creates a new RSS-based Reddit client
func NewClient(userAgent string) *Client {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		parser:    parser,
		userAgent: userAgent,
	}
}

// GetNewPosts fetches the newest posts from a subreddit via RSS
func (c *Client) GetNewPosts(ctx context.Context, subreddit string, limit int) ([]Post, error) {
	feedURL := fmt.Sprintf("https://www.reddit.com/r/%s/new.rss?limit=%d", subreddit, limit)

	feed, err := c.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("parsing RSS feed: %w", err)
	}

	posts := make([]Post, 0, len(feed.Items))
	for _, item := range feed.Items {
		post := Post{
			ID:         item.GUID,
			Title:      item.Title,
			Author:     item.Author.Name,
			URL:        item.Link,
			Permalink:  item.Link,
			Content:    item.Content,
			Categories: item.Categories,
		}

		// Extract subreddit from categories
		for _, cat := range item.Categories {
			if cat != "" {
				post.Subreddit = cat
				break
			}
		}
		if post.Subreddit == "" {
			post.Subreddit = subreddit
		}

		// Parse published time
		if item.PublishedParsed != nil {
			post.CreatedAt = *item.PublishedParsed
		} else {
			post.CreatedAt = time.Now()
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// GetFullPermalink returns the full Reddit URL for a post
func (p *Post) GetFullPermalink() string {
	return p.Permalink
}
