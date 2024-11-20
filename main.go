package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/jomei/notionapi"
	"github.com/mmcdole/gofeed"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

// Config holds the application configuration settings.
// It includes RSS feed URLs and Notion API credentials.
type Config struct {
	Feeds        []string `yaml:"feeds"`
	NotionDBID   string   `yaml:"notion_db_id"`
	NotionAPIKey string   `yaml:"notion_api_key"`
}

// Validate checks if the Config struct has all required fields populated.
// It returns an error if any required field is missing or empty.
func (c *Config) Validate() error {
	// Check if any feeds are specified
	if len(c.Feeds) == 0 {
		return fmt.Errorf("no feeds specified")
	}
	// Verify Notion database ID is provided
	if c.NotionDBID == "" {
		return fmt.Errorf("notion_db_id is required")
	}
	// Verify Notion API key is provided
	if c.NotionAPIKey == "" {
		return fmt.Errorf("notion_api_key is required")
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		printUsage()
	case "run":
		runSync()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`RSS to Notion Sync Tool

Usage:
  %s <command>

Commands:
  run   Execute the RSS to Notion synchronization
  help  Show this help message

Configuration:
  Create a feeds.yaml file with the following structure:
    feeds:
      - https://example.com/feed.xml
    notion_db_id: your_notion_database_id
    notion_api_key: your_notion_api_key
`, os.Args[0])
}

func runSync() {
	config, err := loadConfig("feeds.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	notionClient := notionapi.NewClient(notionapi.Token(config.NotionAPIKey))

	// 3 requests per second
	// https://developers.notion.com/reference/request-limits#rate-limits
	limiter := rate.NewLimiter(rate.Every(time.Second/3), 1)

	var wg sync.WaitGroup
	for _, feedURL := range config.Feeds {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			processFeed(url, notionClient, config.NotionDBID, limiter)
		}(feedURL)
	}
	wg.Wait()
}

// processFeed fetches RSS feed items from the given URL and adds them to a Notion database.
// It uses a rate limiter to control API requests and checks for existing entries to avoid duplicates.
// Parameters:
//   - feedURL: URL of the RSS feed to process
//   - client: Notion API client instance
//   - dbID: ID of the target Notion database
//   - limiter: Rate limiter to control API request frequency
func processFeed(feedURL string, client *notionapi.Client, dbID string, limiter *rate.Limiter) {
	log.Printf("Fetching: %s\n", feedURL)
	items, err := fetchFeed(feedURL)
	if err != nil {
		log.Printf("Error fetching feed %s: %v\n", feedURL, err)
		return
	}

	for _, item := range items {
		if err := limiter.Wait(context.Background()); err != nil {
			log.Printf("Rate limiter error: %v\n", err)
			continue
		}

		exists, err := checkIfExists(client, dbID, item.Link)
		if err != nil {
			log.Printf("Error checking if item exists: %v\n", err)
			continue
		}
		if exists {
			log.Printf("Skipping existing item: %s\n", item.Title)
			continue
		}

		if err := addToNotion(client, dbID, item); err != nil {
			log.Printf("Error adding item to Notion: %v\n", err)
			continue
		}
		log.Printf("Added: %s\n", item.Title)
	}
}

// loadConfig reads and parses the YAML configuration file.
// Parameters:
//   - filename: path to the YAML configuration file
//
// Returns:
//   - *Config: pointer to the parsed configuration struct
//   - error: any error that occurred during reading or parsing
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// fetchFeed retrieves RSS feed items from a given URL.
// Parameters:
//   - url: the URL of the RSS feed to fetch
//
// Returns:
//   - []*gofeed.Item: slice of feed items
//   - error: any error that occurred during fetching
func fetchFeed(url string) ([]*gofeed.Item, error) {
	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize feed parser with custom user agent
	fp := gofeed.NewParser()
	fp.UserAgent = "rss-to-notion/1.0 (https://github.com/virtualroot/rss-to-notion)"

	// Parse the feed URL with context
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	// Return the feed items
	return feed.Items, nil
}

// checkIfExists checks if a URL already exists in the Notion database.
// Parameters:
//   - client: Notion API client
//   - dbID: ID of the Notion database
//   - url: URL to check for existence
//
// Returns:
//   - bool: true if URL exists, false otherwise
//   - error: any error that occurred during the check
func checkIfExists(client *notionapi.Client, dbID string, url string) (bool, error) {
	// Create filter to search for URL in database
	filter := &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "URL",
			RichText: &notionapi.TextFilterCondition{
				Equals: url,
			},
		},
	}

	// Query the database with the filter
	result, err := client.Database.Query(context.Background(), notionapi.DatabaseID(dbID), filter)
	if err != nil {
		return false, err
	}

	// Return true if any results found, false otherwise
	return len(result.Results) > 0, nil
}

// convertToMarkdown converts HTML content to markdown format.
// Parameters:
//   - htmlContent: HTML string to convert
//
// Returns:
//   - string: converted markdown text
//   - error: any error that occurred during conversion
func convertToMarkdown(htmlContent string) (string, error) {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return "", fmt.Errorf("error converting HTML to markdown: %v", err)
	}
	return markdown, nil
}

func addToNotion(client *notionapi.Client, dbID string, item *gofeed.Item) error {
	properties := notionapi.Properties{
		"Name": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{Text: &notionapi.Text{Content: item.Title}},
			},
		},
		"URL": notionapi.URLProperty{
			URL: item.Link,
		},
	}

	// Add published date if available
	if item.PublishedParsed != nil {
		properties["Published"] = notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: (*notionapi.Date)(item.PublishedParsed),
			},
		}
	}

	markdown, err := convertToMarkdown(item.Content)

	if err != nil {
		return err
	}

	content := markdown
	var children []notionapi.Block
	for len(content) > 0 {
		chunk := content
		if len(chunk) > 2000 {
			chunk = content[:2000]
			content = content[2000:]
		} else {
			content = ""
		}

		children = append(children, &notionapi.ParagraphBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: "block",
				Type:   "paragraph",
			},
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					{Text: &notionapi.Text{Content: chunk}},
				},
			},
		})
	}

	_, err = client.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(dbID),
		},
		Properties: properties,
		Children:   children,
	})

	return err
}
