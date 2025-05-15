package config

// Config holds notifier configuration
type Config struct {
	MarshallerType string `toml:"marshaller_type"`
	HasherType     string `toml:"hasher_type"`

	MinBlocksConfirmation uint8             `toml:"min_blocks_confirmation"`
	SubscribedEvents      []SubscribedEvent `toml:"subscribed_events"`
	ClientConfig          ClientConfig      `toml:"client_config"`
}

// SubscribedEvent holds subscribed events config
type SubscribedEvent struct {
	Identifier string `toml:"identifier"`
	Address    string `toml:"addresses"`
}

// ClientConfig holds client web sockets config
type ClientConfig struct {
	Url string `toml:"url"`
}
