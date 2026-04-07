package types

// PM
type PMInterface interface {
	Start(StartMessage) error
	Stop(StopMessage) error
	Delete(DeleteMessage) error
	Restart(RestartMessage) error
	Input(name string, command string) error
	List() []ListElement
	Tail(name string, lineCount int) ([]string, []string, error)
}

// Logger
type LoggerInterface interface {
	Logln(v ...any)
	Errorln(v ...any)
}

// UDS
type ServerInterface interface {
	Broadcast(name string, JSON []byte)
}
