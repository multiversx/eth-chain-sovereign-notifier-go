package factory

// ETHClient defines what a websocket client should do
type ETHClient interface {
	Close()
}
