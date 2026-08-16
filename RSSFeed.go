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

func scrapeFeeds(ctx context.Context, s *state) error {
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("flop getting next feed to fetch: %w", err)
	}
	_, err = s.db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("flop marking feed as fetched: %w", err)
	}

	rssFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return err
	}

	fmt.Printf("%s @ %s\n%s\n", rssFeed.Channel.Title, rssFeed.Channel.Link, rssFeed.Channel.Description)
	if len(rssFeed.Channel.Item) > 0 {
		for i := range rssFeed.Channel.Item {
			fmt.Printf(" - %s - %s\n\t%s\n\tLink: %s\n", rssFeed.Channel.Item[i].Title, rssFeed.Channel.Item[i].PubDate, rssFeed.Channel.Item[i].Description, rssFeed.Channel.Item[i].Link)
		}
	} else {
		fmt.Println("No feed items at the moment.")
	}

	return nil
}
