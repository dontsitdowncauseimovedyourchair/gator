package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var err error
	if req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil); err == nil {
		client := &http.Client{}
		req.Header.Set("User-Agent", "gator")
		if res, err := client.Do(req); err == nil {
			defer res.Body.Close()
			if res.StatusCode > 299 {
				return nil, fmt.Errorf("response status code indicates flop: %d", res.StatusCode)
			}
			if data, err := io.ReadAll(res.Body); err == nil {
				var feed RSSFeed
				if err = xml.Unmarshal(data, &feed); err == nil {
					feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
					feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
					for i := range feed.Channel.Item {
						feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
						feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
					}
					return &feed, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("fetch flop: %w", err)
}
