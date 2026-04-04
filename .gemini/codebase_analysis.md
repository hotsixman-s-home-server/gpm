# Codebase Analysis: Go Process Manager (GPM)

GPM은 PM2의 핵심 기능을 벤치마킹한 Go 기반의 경량 프로세스 매니저입니다.

## 1. 아키텍처 개요 (Architecture Overview)

GPM은 **CLI (Client)**와 **Daemon (Server)** 구조로 설계되었습니다. 사용자는 CLI를 통해 명령을 내리고, 백그라운드에서 실행되는 Daemon이 실제 프로세스 관리와 상태 유지를 담당합니다.

### 1.1 프로그램 진입점 (`main.go`)
- 환경 변수 `GPM_DAEMON_PROCESS`의 존재 여부에 따라 실행 모드를 결정합니다.
  - `1`인 경우: `daemon.DaemonInit()` 실행 (Daemon 모드)
  - 그 외: `cli.Execute()` 실행 (CLI 모드)

## 2. 주요 모듈별 역할

### 2.1 `module/cli` (Cobra 기반 CLI)
- **Root**: `rootCmd` 정의 및 하위 명령어 관리.
- **Init**: `gpm init` 명령 시 `daemon.Daemonize()`를 호출하여 백그라운드 데몬을 시작합니다.
- **Connect**: 데몬과의 통신을 위한 초기 연결 로직을 담당합니다.

### 2.2 `module/daemon` (백그라운드 관리)
- **Self-Daemonization**: 현재 실행 중인 바이너리를 환경 변수와 함께 재실행(`exec.Command`)하여 터미널 세션과 분리합니다.
- **Platform Support**: `unix.go`, `windows.go`를 통해 OS별 데몬화 설정(Setsid 등)을 지원합니다.
- **DaemonInit**: 데몬 시작 시 로거, DB 초기화, PID 기록, UDS 서버 시작 등을 수행합니다.

### 2.3 `module/uds` (Inter-Process Communication)
- **Socket Path**: `~/.gpm/gpm.sock`을 기본 경로로 사용합니다.
- **IPC**: Unix Domain Socket을 통해 CLI와 Daemon 간의 JSON 기반 통신을 처리합니다.

### 2.4 `module/database` (상태 유지)
- **SQLite**: `~/.gpm/main.db`에 프로세스 정보 및 PID를 저장합니다.
- **PID Table**: 데몬의 중복 실행 방지를 위해 현재 데몬의 PID를 기록합니다.

### 2.5 `module/logger` (로그 관리)
- 데몬 자체 로그 및 관리 대상 프로세스의 로그(`stdout`, `stderr`)를 캡처하여 파일로 저장합니다.

## 3. 핵심 워크플로우

1. **데몬 시작**: `gpm init` -> `Daemonize()` (재실행) -> `DaemonInit()` (상주 시작).
2. **명령 전달**: `gpm start <app>` -> CLI가 UDS를 통해 Daemon에 요청 -> Daemon이 `os/exec`으로 앱 실행.
3. **상태 모니터링**: Daemon이 실행 중인 앱의 상태(CPU, Memory) 및 로그를 실시간으로 추적.

## 4. 설정 및 파일 경로
- **Base Dir**: `~/.gpm/`
- **DB Path**: `~/.gpm/main.db`
- **Socket Path**: `~/.gpm/gpm.sock`
- **Daemon Log**: `~/.gpm/daemon.log`
