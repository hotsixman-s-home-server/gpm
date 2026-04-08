package daemon

import (
	"errors"
	"geep/module/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func checkPid() (int, bool, error) {
	geepDir, err := util.GetGeepDir()
	if err != nil {
		return 0, false, err
	}

	pidFilePath := filepath.Join(geepDir, "pid")

	// 1. 파일 읽기
	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		return 0, false, nil // 파일이 없으면 당연히 프로세스도 없는 것으로 간주
	}

	// 2. 숫자 추출 (공백/줄바꿈 제거)
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, false, nil // 숫자가 아니면 잘못된 파일
	}

	// 3. 프로세스 생존 확인
	alive := isAlive(pid)

	// err가 nil이면 존재함, syscall.EPERM이면 권한은 없지만 존재함
	return pid, alive, nil
}

func recordPid() error {
	geepDir, err := util.GetGeepDir()
	if err != nil {
		return err
	}

	pidFilePath := filepath.Join(geepDir, "pid")
	err = deletePid()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(pidFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(os.Getpid()))
	if err != nil {
		return err
	}
	return nil
}

func deletePid() error {
	geepDir, err := util.GetGeepDir()
	if err != nil {
		return err
	}

	pidFilePath := filepath.Join(geepDir, "pid")

	err = os.Remove(pidFilePath)

	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
