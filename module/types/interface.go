package types

// PM
type PMInterface interface {
	Start(StartMessage) error
	Input(name string, command string)
}

// Logger
type LoggerInterface interface {
}

// UDS
type ServerInterface interface {
	Broadcast(name string, JSON []byte)
}
