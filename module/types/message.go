package types

type ConnectRequestMessage struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type CommandMessage struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type StartMessage struct {
	Type string            `json:"type"`
	Name string            `json:"name"`
	Run  string            `json:"run"`
	Args []string          `json:"args"`
	Cwd  string            `json:"cwd"`
	Env  map[string]string `json:"Env"`
}

type LogMessage struct {
	// "log" | "error"
	Type    string `json:"type"`
	Message string `json:"message"`
}

type StartResultMessage struct {
	// "startResult"
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
