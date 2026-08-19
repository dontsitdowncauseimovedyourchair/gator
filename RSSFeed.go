package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/gator/internal/database"
	"github.com/google/uuid"
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

	fmt.Printf("Fetching %s @ %s\n%s\n", rssFeed.Channel.Title, rssFeed.Channel.Link, rssFeed.Channel.Description)

	if len(rssFeed.Channel.Item) > 0 {
		for i := range rssFeed.Channel.Item {
			_, err = s.db.CreatePost(ctx, database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       rssFeed.Channel.Item[i].Title,
				Url:         rssFeed.Channel.Item[i].Link,
				Description: parseDescription(rssFeed.Channel.Item[i].Description),
				PublishedAt: parseTime(rssFeed.Channel.Item[i].PubDate),
				FeedID:      feed.ID,
			})
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					continue
				}
				fmt.Printf("Flop saving post: %s\n", err.Error())
			}
		}
	} else {
		fmt.Println("No feed items at the moment.")
	}

	return nil
}

func parseTime(timeStr string) sql.NullTime {
	layouts := []string{
		time.RFC822,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822Z,
		time.RFC850,
		time.RFC3339,
		time.RFC3339Nano,
		time.ANSIC,
		time.RubyDate,
		time.DateTime,
	}
	for i := range layouts {
		if timeObj, err := time.Parse(layouts[i], timeStr); err == nil {
			return sql.NullTime{
				Time:  timeObj,
				Valid: true,
			}
		}
	}
	return sql.NullTime{
		Time:  time.Time{},
		Valid: false,
	}
}

func parseDescription(descStr string) sql.NullString {
	if len(strings.TrimSpace(descStr)) == 0 {
		return sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	return sql.NullString{
		String: html.UnescapeString(strings.TrimSpace(descStr)),
		Valid:  true,
	}
}
