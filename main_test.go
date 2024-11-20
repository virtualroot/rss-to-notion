package main

import (
	"testing"
)

func TestMain(t *testing.T) {

}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Feeds:        []string{"http://example.com/blog.rss", "http://example.org/feed.atom", "https://blog.example.net/feed.xml"},
				NotionDBID:   "8a994146-f3da-4f92-8a9d-2f993c16d95f",
				NotionAPIKey: "b4d4f338-b1c4-4c22-a29c-c5fec9d63c0a",
			},
			wantErr: false,
		},
		{
			name: "missing feeds",
			config: Config{
				Feeds:        []string{},
				NotionDBID:   "8a994146-f3da-4f92-8a9d-2f993c16d95f",
				NotionAPIKey: "b4d4f338-b1c4-4c22-a29c-c5fec9d63c0a",
			},
			wantErr: true,
		},
		{
			name: "missing notion db id",
			config: Config{
				Feeds:        []string{"http://example.com/blog.rss"},
				NotionDBID:   "",
				NotionAPIKey: "b4d4f338-b1c4-4c22-a29c-c5fec9d63c0a",
			},
			wantErr: true,
		},
		{
			name: "missing notion api key",
			config: Config{
				Feeds:        []string{"http://example.com/blog.rss"},
				NotionDBID:   "8a994146-f3da-4f92-8a9d-2f993c16d95f",
				NotionAPIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
