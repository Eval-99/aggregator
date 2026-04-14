package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")
	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	defer res.Body.Close()

	if res.StatusCode > 299 {
		fmt.Println("Page error, code was not 200 -> 299")
		os.Exit(1)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	var Feed RSSFeed
	if err := xml.Unmarshal(data, &Feed); err != nil {
		return &RSSFeed{}, err
	}

	Feed.Channel.Title = html.UnescapeString(Feed.Channel.Title)
	Feed.Channel.Description = html.UnescapeString(Feed.Channel.Description)

	for i := range Feed.Channel.Item {
		Feed.Channel.Item[i].Title = html.UnescapeString(Feed.Channel.Item[i].Title)
		Feed.Channel.Item[i].Description = html.UnescapeString(Feed.Channel.Item[i].Description)
	}

	return &Feed, nil
}
