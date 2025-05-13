package config

// Config holds notifier configuration
type Config struct {
	SubscribedEvents []SubscribedEvent `toml:"subscribed_events"`
	ClientConfig     ClientConfig      `toml:"client_config"`
}

// SubscribedEvent holds subscribed events config
type SubscribedEvent struct {
	Identifier string   `toml:"identifier"`
	Addresses  []string `toml:"addresses"`
}

// ClientConfig holds client web sockets config
type ClientConfig struct {
	Url string `toml:"url"`
}
