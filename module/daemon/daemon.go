package daemon

import (
	"fmt"
	"geep/module/database"
	"geep/module/logger"
	"geep/module/pm"
	"geep/module/server"
	"geep/module/util"
	"os"
	"os/exec"
)

const DAEMON_ENV = "GEEP_DAEMON_PROCESS"

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
	env := append(os.Environ(), DAEMON_ENV+"=1")
	cmd.Env = env
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

	// .geep folder
	geepDir, err := util.GetGeepDir()
	if err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}
	err = os.MkdirAll(geepDir, 0755)
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
	mainLogger, err := logger.GetMainLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// pid 체크
	_, running, err := checkPid()
	if err != nil {
		mainLogger.Logln(err)
		os.Exit(1)
	}
	if running {
		mainLogger.Logln("GEEP is already running.")
		os.Exit(1)
	}

	// pid 저장
	err = recordPid()
	if err != nil {
		mainLogger.Logln("Cannot record pid.")
		os.Exit(1)
	}
	defer deletePid()

	// 서버 생성
	udsServer, err := server.NewUDSServer()
	if err != nil {
		mainLogger.Logln("Cannot listen uds server.")
		os.Exit(1)
	}
	udsServer.SetLogger(mainLogger)
	mainLogger.SetServer(udsServer)

	// pm 생성
	PM := pm.NewPM(mainLogger)
	PM.SetServer(udsServer)
	udsServer.SetPM(PM)

	// log
	mainLogger.Logln("GEEP daemon started.")

	select {}
}
