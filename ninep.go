package ninep

// Option sets Server or Client options such as logging, max message
// size etc.
type Option func(interface{}) error

// version defines the protocol version string.
const version = "9P2000.L"
