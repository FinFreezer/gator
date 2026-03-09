package internal

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
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
