package config

import (
	"github.com/ossf/package-feeds/events"
	"github.com/ossf/package-feeds/feeds"
)

var Hi = &feeds.Package{}

type ScheduledFeedConfig struct {
	// Configures the publisher for pushing packages after polling
	PubConfig PublisherConfig `yaml:"publisher"`

	// Configures the feeds to be used for polling from package repositories
	EnabledFeeds []string `yaml:"enabled_feeds"`

	HTTPPort int    `yaml:"http_port,omitempty"`
	PollRate string `yaml:"poll_rate"`
	Timer    bool   `yaml:"timer"`

	// Configures the EventHandler instance to be used throughout the package-feeds application
	EventsConfig *EventsConfig `yaml:"events"`

	eventHandler *events.Handler
	// 1 comment
	// 2 comment
	// 3 comment
	// 4 comment
	// 5 comment
	// 6 comment
	// 7 comment
	// 8 comment
	// 9 comment
	// 10 comment
}

type PublisherConfig struct {
	Type   string      `mapstructure:"type"`
	Config interface{} `mapstructure:"config"`
}

type EventsConfig struct {
	Sink        string        `yaml:"sink"`
	EventFilter events.Filter `yaml:"filter"`
}
