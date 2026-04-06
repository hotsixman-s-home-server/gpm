package types

import "fmt"

// Server
type InvalidMessage struct {
	JSON string
}

func (m InvalidMessage) Error() string {
	return fmt.Sprintf("Invalid message: %s", m.JSON)
}

type UndefinedProcessNameError struct{}

func (_ UndefinedProcessNameError) Error() string {
	return "'name' field is not defined."
}

// PM
/*
특정 이름의 프로세스가 없습니다.
*/
type NoProcessError struct {
	Name string
}

func (this NoProcessError) Error() string {
	return fmt.Sprintf("There is no process named \"%s\"", this.Name)
}

/*
프로세스가 실행중입니다.
*/
type ProcessRunningError struct {
	Name string
}

func (this ProcessRunningError) Error() string {
	return fmt.Sprintf("Process \"%s\" is running.", this.Name)
}
