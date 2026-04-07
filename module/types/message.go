package types

type ConnectRequestMessage struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type CommandMessage struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// start

type StartMessage struct {
	// "start"
	Type            string            `json:"type"`
	Name            string            `json:"name"`
	Run             string            `json:"run"`
	Args            []string          `json:"args"`
	Cwd             string            `json:"cwd"`
	Env             map[string]string `json:"env"`
	MaxRestartCount int               `json:"maxRestartCount"`
}

type StartResultMessage struct {
	// "startResult"
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// stop

type StopMessage struct {
	// "stop"
	Type string `json:"type"`
	Name string `json:"name"`
}

type StopResultMessage struct {
	// "stopResult"
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Delete
type DeleteMessage struct {
	// "delete"
	Type string `json:"type"`
	Name string `json:"name"`
}

type DeleteResultMessage struct {
	// "deleteResult"
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// List
type ListMessage struct {
	// "list"
	Type string `json:"type"`
}

type ListResultMessage struct {
	// "listResult"
	Type string        `json:"type"`
	List []ListElement `json:"list"`
}

type ListElement struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	CPUPercent float64 `json:"cpuPercent"`
	Mem        float64 `json:"mem"`
	Restart    int     `json:"restart"`
}

// log
type LogMessage struct {
	// "log" | "error"
	Type    string `json:"type"`
	Message string `json:"message"`
}
