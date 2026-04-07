# Codebase Analysis: GEEP (Go + Keep) (GEEP)

GEEP은 Go 기반의 경량 프로세스 매니저입니다.

## 1. 아키텍처 개요 (Architecture Overview)

GEEP은 **CLI (Client)**와 **Daemon (Server)** 구조로 설계되었습니다. 사용자는 CLI를 통해 명령을 내리고, 백그라운드에서 실행되는 Daemon이 실제 프로세스 관리와 상태 유지를 담당합니다.

### 1.1 프로그램 진입점 (`main.go`)
- 환경 변수 `GEEP_DAEMON_PROCESS`의 존재 여부에 따라 실행 모드를 결정합니다.
  - `1`인 경우: `daemon.DaemonInit()` 실행 (Daemon 모드)
  - 그 외: `cli.Execute()` 실행 (CLI 모드)

## 2. 주요 모듈별 역할

### 2.1 `module/cli` (Cobra & Bubble Tea 기반 CLI)
- **Root**: `rootCmd` 정의 및 하위 명령어 관리.
- **Init**: `geep init` 명령 시 `daemon.Daemonize()`를 호출하여 백그라운드 데몬을 시작합니다.
- **Commands**: `start`, `stop`, `restart`, `delete`, `list`, `connect`, `terminate` 등.
- **TUI (Connect)**: `bubbletea`와 `lipgloss`를 사용하여 관리 중인 프로세스의 실시간 로그를 보여주고 표준 입력을 전달하는 인터페이스를 제공합니다.

### 2.2 `module/daemon` (백그라운드 관리)
- **Self-Daemonization**: 현재 실행 중인 바이너리를 환경 변수와 함께 재실행(`exec.Command`)하여 터미널 세션과 분리합니다.
- **Platform Support**: `unix.go`, `windows.go`를 통해 OS별 데몬화 설정(Setsid 등)을 지원합니다.
- **DaemonInit**: 데몬 시작 시 로거, DB 초기화, PID 기록, 서버 시작 등을 수행합니다.
- **Termination**: `KillDaemon()` 함수를 통해 실행 중인 데몬 PID를 찾아 안전하게 종료합니다.

### 2.3 `module/server` & `module/client` (Inter-Process Communication)
- **IPC**: Unix Domain Socket을 통해 CLI와 Daemon 간의 JSON 기반 통신을 처리합니다.
- **Server**: 데몬에서 실행되며 클라이언트의 요청을 대기하고 처리합니다.
- **Client**: CLI 명령 실행 시 서버에 연결하여 요청을 전송합니다. `connect` 명령 시에는 양방향 스트리밍 통신을 수행합니다.

### 2.4 `module/pm` (Process Management)
- 실제 외부 프로세스를 실행(`start.go`), 중지(`stop.go`), 재시작(`restart.go`) 시키는 핵심 비즈니스 로직을 담당합니다.
- **Recovery**: 프로세스 종료 시 `MaxRecoverCount`만큼 자동 재시작 로직을 포함합니다.

### 2.5 `module/logger` (Log Management)
- 데몬 자체 로그 및 관리 대상 프로세스의 로그(`stdout`, `stderr`)를 캡처하여 파일로 저장합니다.
- **Log Rotation**: 설정된 `MaxLogfileSize`를 초과하지 않도록 관리합니다.

## 3. 핵심 워크플로우

1. **데몬 시작**: `geep init` -> `Daemonize()` (재실행) -> `DaemonInit()` (상주 시작).
2. **명령 전달**: `geep start` -> CLI(`client`)가 UDS를 통해 Daemon(`server`)에 요청 -> Daemon(`pm`)이 프로세스 실행.
3. **상태 모니터링**: Daemon이 실행 중인 앱의 상태 및 로그를 실시간으로 추적하며, `list` 요청 시 정보를 반환합니다.
4. **실시간 연결**: `geep connect` -> UDS 스트리밍 연결 -> Daemon이 해당 프로세스의 실시간 로그 전송 및 CLI로부터 입력 수신.

## 4. 설정 및 파일 경로
- **Base Dir**: `~/.geep/`
- **Socket Path**: `~/.geep/geep.sock`
- **Daemon Log**: `~/.geep/daemon.log`
- **Process Logs**: `~/.geep/log-process/[name].log`
