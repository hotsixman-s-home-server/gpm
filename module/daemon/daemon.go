package daemon

import (
	"fmt"
	"gpm/module/database"
	"gpm/module/logger"
	"gpm/module/pm"
	"gpm/module/uds"
	"gpm/module/util"
	"os"
	"os/exec"
)

const DAEMON_ENV = "GPM_DAEMON_PROCESS"

/*
-1: Not daemon
0: success
1: start error
2: already running
*/
func SpawnDaemon() (int, error) {
	if os.Getenv(DAEMON_ENV) == "1" {
		return -1, nil
	}

	_, alive, _ := checkPid()
	if alive {
		return 2, nil
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), DAEMON_ENV+"=1")
	//cmd.Stdin = nil
	//cmd.Stdout = nil
	//cmd.Stderr = nil

	setupDaemon(cmd)
	if err := cmd.Start(); err != nil {
		return 1, err
	}
	return 0, nil
}

func KillDaemon() (int, error) {
	pid, running, err := checkPid()
	if err != nil {
		return 1, err
	}

	if !running {
		return -1, nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return -1, err
	}

	err = process.Kill()
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func DaemonInit() {
	if os.Getenv(DAEMON_ENV) != "1" {
		return
	}

	// .gpm folder
	homeDir, err := util.GetHomeDirPath()
	if err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}
	err = os.MkdirAll(homeDir, 0644)
	if err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}

	// db
	err = database.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer database.DB.Close()

	// main logger
	mainLog, err := logger.GetMainLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// pid 체크
	_, running, err := checkPid()
	if err != nil {
		mainLog.Logln(err)
		os.Exit(1)
	}
	if running {
		mainLog.Logln("GPM is already running.")
		os.Exit(1)
	}

	// pid 저장
	err = recordPid()
	if err != nil {
		mainLog.Logln("Cannot record pid.")
		os.Exit(1)
	}
	defer deletePid()

	// 서버 생성
	udsServer, err := uds.Listen(mainLog)
	if err != nil {
		mainLog.Logln("Cannot listen uds server.")
		os.Exit(1)
	}
	mainLog.SetServer(udsServer)

	// pm 생성
	PM := pm.NewPM(mainLog)
	PM.SetServer(udsServer)
	udsServer.SetPM(PM)

	// log
	mainLog.Logln("GPM daemon started.")

	select {}
}
