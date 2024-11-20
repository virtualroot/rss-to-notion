# RSS to Notion

A Go application that automatically syncs feeds to a Notion database. It fetches feeds in parallel, converts HTML content to Markdown, and creates pages in your Notion database while avoiding duplicates.

## Features

- Supports RSS, Atom and JSONFeed
- Parallel processing of multiple feeds
- HTML to Markdown conversion until we find a way to support Notion's RichText
- Duplicate entry detection
- Rate limiting to respect Notion API limits
- YAML configuration file support

## Prerequisites

- Go :hamster:
- A Notion account and integration API key
- A Notion database to store the feeds entries

## Setup

1. Create a Notion integration:
   - Go to https://www.notion.so/my-integrations
   - Click "New integration"
   - Give it a name, choose Type "Internal" and associate it with your workspace
   - Save the API key

2. Create a Notion database:
   - Create a new database in Notion
   - Add the following properties:
     - Name (title)
     - URL (url)
     - Published (date)
   - Share the database with your integration
   - Copy the database ID from the URL (the part after the workspace name and before the question mark)

3. Clone the repository:
   ```bash
   git clone https://github.com/virtualroot/rss-to-notion
   cd rss-to-notion
   ```

4. Create a `feeds.yaml` configuration file:
   ```yaml
   feeds:
      - https://example.com/blog.rss
      - https://example.org/feed.atom
      - https://blog.example.net/feed.xml
   notion_db_id: "your-database-id"
   notion_api_key: "your-integration-api-key"
   ```

5. Build and run:
   ```bash
   go build
   ./rss-to-notion
   ```

## Configuration

The `feeds.yaml` file supports the following options:

- `feeds`: List of RSS feed URLs to monitor
- `notion_db_id`: Your Notion database ID
- `notion_api_key`: Your Notion API integration token

## Rate Limiting

The application implements rate limiting of 3 requests per second to comply with Notion's API guidelines.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
