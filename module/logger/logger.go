package logger

import (
	"encoding/json"
	"fmt"
	"geep/module/types"
	"geep/module/util"
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
	//KB
	maxFileSize int
	//B
	logFileSize int
	//B
	errorFileSize int
}

func GetMainLogger() (*Logger, error) {
	geepDir, err := util.GetGeepDir()
	if err != nil {
		return nil, err
	}

	dirPath := filepath.Join(geepDir, "log")
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
		dirPath:     dirPath,
		logFile:     logFile,
		errorFile:   errorFile,
		name:        "",
		server:      nil,
		mutex:       &sync.Mutex{},
		main:        true,
		maxFileSize: 1024 * 100,
	}, nil
}

func CreateLogger(name string, timeRecording bool, server types.ServerInterface, maxFileSize int) (*Logger, error) {
	geepDir, err := util.GetGeepDir()
	if err != nil {
		return nil, err
	}

	dirPath := filepath.Join(geepDir, "log-process", filepath.Clean(name))
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
		dirPath:     dirPath,
		logFile:     logFile,
		errorFile:   errorFile,
		name:        name,
		server:      server,
		mutex:       &sync.Mutex{},
		main:        false,
		maxFileSize: maxFileSize,
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
		this.appendError(header + " " + message)
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

func (this *Logger) appendLog(message string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	message = message + "\n"
	size := len([]byte(message))
	if this.logFileSize+size > this.maxFileSize*1024 {
		err := this.newLogFile()
		if err != nil {
			return err
		}
	}

	_, err := this.logFile.WriteString(message)
	if err != nil {
		return err
	}
	this.logFileSize += size
	return nil
}
func (this *Logger) appendError(message string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	message = message + "\n"
	size := len([]byte(message))
	if this.errorFileSize+size > this.maxFileSize*1024 {
		err := this.newLogFile()
		if err != nil {
			return err
		}
	}

	_, err := this.errorFile.WriteString(message + "\n")
	if err != nil {
		return err
	}
	this.errorFileSize += size
	return nil
}

func (this *Logger) newLogFile() error {
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

	if this.logFile != nil {
		this.logFile.Close()
	}
	if this.errorFile != nil {
		this.errorFile.Close()
	}

	logFile, err := os.OpenFile(filepath.Join(this.dirPath, logFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	errorFile, err := os.OpenFile(filepath.Join(this.dirPath, errorFilename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	this.logFile = logFile
	this.errorFile = errorFile
	this.logFileSize = 0
	this.errorFileSize = 0
	return nil
}

func (this *Logger) TailLogs(lineCount int) ([]string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	filename := ""
	if this.logFile != nil {
		filename = this.logFile.Name()
	}

	return tailLines(filename, lineCount)
}

func (this *Logger) TailErrors(lineCount int) ([]string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	filename := ""
	if this.errorFile != nil {
		filename = this.errorFile.Name()
	}

	return tailLines(filename, lineCount)
}

func tailLines(filename string, lineCount int) ([]string, error) {
	if lineCount <= 0 {
		return []string{}, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	if size == 0 {
		return []string{}, nil
	}

	const bufferSize = 4096
	var (
		buf        = make([]byte, bufferSize)
		result     []byte
		foundLines = 0
		offset     = size
	)

	for offset > 0 && foundLines <= lineCount {
		readSize := int64(bufferSize)
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		n, err := f.ReadAt(buf[:readSize], offset)
		if err != nil && n == 0 {
			break
		}

		// 현재 버퍼에서 개행 문자 카운트 (역순 탐색)
		for i := n - 1; i >= 0; i-- {
			if buf[i] == '\n' {
				foundLines++
				if foundLines > lineCount {
					// 요청한 줄 수만큼 찾았으므로, 현재 개행 문자 다음부터의 데이터를 결과에 추가
					chunk := make([]byte, n-(i+1))
					copy(chunk, buf[i+1:n])
					result = append(chunk, result...)
					goto done
				}
			}
		}
		// 아직 줄 수가 부족하면 현재 읽은 블록 전체를 결과 앞에 붙임
		chunk := make([]byte, n)
		copy(chunk, buf[:n])
		result = append(chunk, result...)
	}

done:
	lines := strings.Split(string(result), "\n")

	// 파일이 개행 문자로 끝나는 경우 발생하는 마지막 빈 요소 제거
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// 실제 요청한 개수만큼 슬라이싱 (안전한 경계값 처리)
	if len(lines) > lineCount {
		lines = lines[len(lines)-lineCount:]
	}

	return lines, nil
}
