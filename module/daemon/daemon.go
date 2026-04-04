package daemon

import (
	"gpm/module/database"
	"gpm/module/logger"
	"gpm/module/pm"
	"gpm/module/uds"
	"os"
	"os/exec"
)

const DAEMON_ENV = "GPM_DAEMON_PROCESS"

func Daemonize() {
	if os.Getenv(DAEMON_ENV) == "1" {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), DAEMON_ENV+"=1")

	setupDaemon(cmd)
	if err := cmd.Start(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func DaemonInit() {
	if os.Getenv(DAEMON_ENV) != "1" {
		return
	}

	// main logger
	log, err := logger.GetMainLogger()
	if err != nil {
		os.Exit(1)
	}

	// db
	defer database.DB.Close()

	// pid 체크
	_, running, err := PIDManager.CheckPID()
	if err != nil {
		log.Logln("Cannot check GPM daemon is running.")
		os.Exit(1)
	}
	if running {
		log.Logln("GPM is already running.")
		os.Exit(1)
	}

	// pid 저장
	err = PIDManager.RecordPid()
	if err != nil {
		log.Logln("Cannot record pid.")
		os.Exit(1)
	}
	defer PIDManager.DeletePid()

	// 서버 생성
	udsServer, err := uds.Listen()
	if err != nil {
		log.Logln("Cannot listen uds server.")
		os.Exit(1)
	}
	log.SetUDSServer(udsServer)

	// pm 생성
	pm := &pm.PM{}
	udsServer.SetPM(pm)

	select {}
}
