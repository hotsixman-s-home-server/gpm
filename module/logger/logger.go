package logger

import (
	"encoding/json"
	"fmt"
	"gpm/module/database"
	"gpm/module/types"
	"gpm/module/util"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func Logln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[32m" + timeString + " [LOG]" + "\033[0m"
	fmt.Println(header, message)
}

func Errorln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[31m" + timeString + " [ERROR]" + "\033[0m"
	fmt.Println(header, message)
}

type Logger struct {
	dirPath       string
	file          *os.File
	name          string
	timeRecording bool
	udsServer     types.UDSServerInterface
	mutex         *sync.Mutex
}

func GetMainLogger() (*Logger, error) {
	homeDir, err := util.GetHomeDirPath()
	if err != nil {
		return nil, err
	}

	dirPath := filepath.Join(homeDir, "log")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	filename := util.Now() + ".log"
	err = database.DB.UpdateMainLogFile(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(dirPath, filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		dirPath,
		file,
		"",
		true,
		nil,
		&sync.Mutex{},
	}, nil
}

func CreateLogger(name string, timeRecording bool, udsServer types.UDSServerInterface) (*Logger, error) {
	homeDir, err := util.GetHomeDirPath()
	if err != nil {
		return nil, err
	}

	dirPath := filepath.Join(homeDir, "log-process", filepath.Clean(name))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return nil, err
		}
	}

	filename := name + "-" + util.Now() + ".log"
	err = database.DB.UpdateLogFile(name, filename)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(dirPath, filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		dirPath,
		file,
		name,
		timeRecording,
		udsServer,
		&sync.Mutex{},
	}, nil
}

func (this *Logger) SetUDSServer(udsServer types.UDSServerInterface) {
	this.udsServer = udsServer
}

func (this *Logger) Logln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	header := ""
	if this.timeRecording {
		timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
		header = "\033[32m" + timeString + " [LOG]" + "\033[0m"
	} else {
		header = "\033[32m" + " [LOG]" + "\033[0m"
	}

	if this.file != nil {
		this.appendLog(header + " " + message)
	}
	if this.udsServer != nil {
		messageJSON := map[string]string{
			"type":    "log",
			"message": message,
		}

		JSON, err := json.Marshal(messageJSON)
		if err == nil {
			this.udsServer.Broadcast(this.name, JSON)
		}
	}
}

func (this *Logger) Errorln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	header := ""
	if this.timeRecording {
		timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
		header = "\033[31m" + timeString + " [Error]" + "\033[0m"
	} else {
		header = "\033[31m" + " [Error]" + "\033[0m"
	}

	if this.file != nil {
		this.appendLog(header + " " + message)
	}
	if this.udsServer != nil {
		messageJSON := map[string]string{
			"type":    "error",
			"message": message,
		}

		JSON, err := json.Marshal(messageJSON)
		if err == nil {
			this.udsServer.Broadcast(this.name, JSON)
		}
	}
}

func (this *Logger) appendLog(message string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.file.WriteString(message + "\n")
}

func (this *Logger) ReadLastLines(n int) ([]string, error) {
	this.mutex.Lock()
	if this.file == nil {
		this.mutex.Unlock()
		return nil, fmt.Errorf("file is not opened")
	}
	filePath := this.file.Name()
	this.mutex.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	if fileSize == 0 {
		return []string{}, nil
	}

	return this.readLastLinesInternal(file, n, fileSize)
}

func (this *Logger) readLastLinesInternal(file *os.File, n int, fileSize int64) ([]string, error) {
	var (
		lines   []string
		cursor  int64 = 0
		bufSize int64 = 1024
		tail    string
	)

	for int64(len(lines)) <= int64(n) && cursor < fileSize {
		cursor += bufSize
		if cursor > fileSize {
			cursor = fileSize
		}

		_, err := file.Seek(-cursor, io.SeekEnd)
		if err != nil {
			return nil, err
		}

		data := make([]byte, bufSize)
		if cursor == fileSize && fileSize%bufSize != 0 {
			data = make([]byte, fileSize%bufSize)
		}

		nRead, _ := file.Read(data)
		content := string(data[:nRead]) + tail

		lines = strings.Split(content, "\n")
		// 파일 끝이 \n으로 끝나면 마지막 빈 요소 제거
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		if len(lines) > 1 {
			tail = lines[0]
			lines = lines[1:]
		}

		if cursor >= fileSize {
			break
		}
	}

	if len(lines) > n {
		return lines[len(lines)-n:], nil
	}
	return lines, nil
}

func (this *Logger) recreateFile() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	var filename string
	if this.name == "" {
		filename = util.Now() + ".log"
		err := database.DB.UpdateMainLogFile(filename)
		if err != nil {
			return err
		}
	} else {
		filename = this.name + "-" + util.Now() + ".log"
		err := database.DB.UpdateLogFile(this.name, filename)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(filepath.Join(this.dirPath, filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	this.file = file
	return nil
}
