package internal

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/finfreezer/blogAggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	newRSSFeed := RSSFeed{}
	newReq, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	newReq.Header.Set("User-Agent", "gator")
	if err != nil {
		return nil, fmt.Errorf("Something went wrong, error: %s", err.Error())
	}

	newResponse, err := http.DefaultClient.Do(newReq)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong, error: %s", err.Error())
	}
	defer newResponse.Body.Close()

	data, err := io.ReadAll(newResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong, error: %s", err.Error())
	}
	xml.Unmarshal(data, &newRSSFeed)

	for i, item := range newRSSFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		newRSSFeed.Channel.Item[i] = item
	}
	//fmt.Println(newRSSFeed)

	return &newRSSFeed, nil
}

func scrapeFeeds(s *State) error {
	nextFeed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}
	s.Db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID},
	)
	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}

	fmt.Printf("Reading feed from %s\n", nextFeed.Url)
	for _, item := range feed.Channel.Item {
		fmt.Println(item.Title)
	}
	fmt.Printf("\n\n")
	return nil
}
