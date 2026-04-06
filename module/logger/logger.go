package logger

import (
	"encoding/json"
	"fmt"
	"gpm/module/types"
	"gpm/module/util"
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

func SLogln(v ...any) string {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[32m" + timeString + " [LOG]" + "\033[0m"
	return header + " " + message
}

func SErrorln(v ...any) string {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[31m" + timeString + " [ERROR]" + "\033[0m"
	return header + " " + message
}

type Logger struct {
	dirPath   string
	logFile   *os.File
	errorFile *os.File
	name      string
	server    types.ServerInterface
	mutex     *sync.Mutex
	main      bool
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

	logFilename := strings.ReplaceAll(util.Now(), ":", "_") + " log.log"
	errorFilename := strings.ReplaceAll(util.Now(), ":", "_") + " error.log"
	//err = database.DB.UpdateMainLogFile(filename)
	//if err != nil {
	//	return nil, err
	//}

	logFile, err := os.OpenFile(filepath.Join(dirPath, logFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	errorFile, err := os.OpenFile(filepath.Join(dirPath, errorFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		dirPath:   dirPath,
		logFile:   logFile,
		errorFile: errorFile,
		name:      "",
		server:    nil,
		mutex:     &sync.Mutex{},
		main:      true,
	}, nil
}

func CreateLogger(name string, timeRecording bool, server types.ServerInterface) (*Logger, error) {
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

	logFilename := name + "-" + strings.ReplaceAll(util.Now(), ":", "_") + " log.log"
	errorFilename := name + "-" + strings.ReplaceAll(util.Now(), ":", "_") + " error.log"
	//err = database.DB.UpdateLogFile(name, filename)
	//if err != nil {
	//	return nil, err
	//}

	logFile, err := os.OpenFile(filepath.Join(dirPath, logFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	errorFile, err := os.OpenFile(filepath.Join(dirPath, errorFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		dirPath:   dirPath,
		logFile:   logFile,
		errorFile: errorFile,
		name:      name,
		server:    server,
		mutex:     &sync.Mutex{},
		main:      false,
	}, nil
}

func (this *Logger) SetServer(server types.ServerInterface) {
	this.server = server
}

func (this *Logger) Logln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[32m" + timeString + " [LOG]" + "\033[0m"

	if this.logFile != nil {
		this.appendLog(header + " " + message)
	}
	if this.server != nil {
		messageJSON := map[string]string{
			"type":    "log",
			"message": message,
		}

		JSON, err := json.Marshal(messageJSON)
		if err == nil {
			this.server.Broadcast(this.name, JSON)
		}
	}
	if this.main {
		Logln(v...)
	}
}

func (this *Logger) Errorln(v ...any) {
	message := strings.TrimRight(fmt.Sprintln(v...), " \t\n\r")
	timeString := "[" + time.Now().Format("2006-01-02 15:04:05") + "]"
	header := "\033[31m" + timeString + " [Error]" + "\033[0m"

	if this.errorFile != nil {
		this.appendLog(header + " " + message)
	}
	if this.server != nil {
		messageJSON := map[string]string{
			"type":    "error",
			"message": message,
		}

		JSON, err := json.Marshal(messageJSON)
		if err == nil {
			this.server.Broadcast(this.name, JSON)
		}
	}
	if this.main {
		Errorln(v...)
	}
}

func (this *Logger) appendLog(message string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.logFile.WriteString(message + "\n")
}
func (this *Logger) appendError(message string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.errorFile.WriteString(message + "\n")
}

/*
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
*/

func (this *Logger) recreateFile() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	logFilename := ""
	errorFilename := ""
	if this.main {
		logFilename = strings.ReplaceAll(util.Now(), ":", "_") + " log.log"
		errorFilename = strings.ReplaceAll(util.Now(), ":", "_") + " error.log"
		//err := database.DB.UpdateMainLogFile(filename)
		//if err != nil {
		//	return err
		//}
	} else {
		logFilename = this.name + "-" + strings.ReplaceAll(util.Now(), ":", "_") + " log.log"
		errorFilename = this.name + "-" + strings.ReplaceAll(util.Now(), ":", "_") + " error.log"
		//err := database.DB.UpdateLogFile(this.name, filename)
		//if err != nil {
		//	return err
		//}
	}

	logFile, err := os.OpenFile(filepath.Join(this.dirPath, logFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	errorFile, err := os.OpenFile(filepath.Join(this.dirPath, errorFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	this.logFile = logFile
	this.errorFile = errorFile
	return nil
}
