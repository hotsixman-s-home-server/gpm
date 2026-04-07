# GEEP (Go + Keep) (GEEP) Specification

Go 기반의 간단한 프로세스 매니저 사양서입니다.

## 1. 핵심 기능 (Core Features)

### 1.1 프로세스 생명주기 관리
- **Start**: 새로운 프로세스를 실행하고 관리 목록에 추가합니다.
- **Stop**: 실행 중인 프로세스를 안전하게 종료합니다.
- **Restart**: 프로세스를 재시작합니다.
- **Delete**: 관리 목록에서 프로세스를 제거하고 종료합니다.
- **List**: 현재 관리 중인 모든 프로세스의 상태(PID, 상태, 메모리, CPU 등)를 요약해서 보여줍니다.

### 1.2 자동 복구 (Auto-restart)
- 프로세스가 예기치 않게 종료(Crash)될 경우 설정된 정책에 따라 자동으로 재시작합니다.
- 무한 재시작 방지를 위한 최대 재시도 횟수 및 간격 설정을 지원합니다.

### 1.3 자식 프로세스 통신 (Stdio Communication)
- **Stdio Pipes**: 자식 프로세스의 `stdout` 및 `stderr`를 파이프로 연결하여 실시간 로그를 캡처하고 파일로 저장합니다.
- **IPC over Stdio**: `stdin`을 통해 자식 프로세스에 특정 명령을 전달할 수 있는 구조를 지원합니다.
- 각 프로세스별로 별도의 로그 파일을 유지합니다.

## 2. 시스템 아키텍처

### 2.1 클라이언트-서버 구조 (Daemon & CLI)
- **Daemon (Server)**: 백그라운드에서 상주하며 실제 프로세스를 실행, 감시, 관리합니다. 
  - **Self-Daemonize**: `os/exec`을 통해 자기 자신을 재실행하여 터미널 세션과 분리합니다.
  - **Platform Support**: Unix(`Setsid`), Windows(`HideWindow`, `CREATE_NEW_PROCESS_GROUP`)를 모두 지원합니다.
- **CLI (Client)**: 사용자의 명령을 받아 Unix Domain Socket(UDS)을 통해 데몬과 통신합니다. `Cobra` 라이브러리를 사용하여 명령어를 처리합니다.

### 2.2 통신 방식 (IPC)
- **Unix Domain Socket (UDS)**: `module/uds`를 통해 공통 관리됩니다.
- **Socket Path**: `~/.geep/geep.sock` 경로를 사용합니다.
- **Protocol**: JSON 기반의 Request/Response 구조체를 사용합니다.

## 3. 프로젝트 구조 (Module Focused)

- `cmd/geep/`: CLI 엔트리 포인트 및 명령어 정의.
- `module/daemon/`: 프로세스 백그라운드 전환 및 플랫폼별 데몬화 로직.
- `module/uds/`: 소켓 생성, 연결, 데이터 송수신 공통 로직.
- `module/models/`: 공통 데이터 구조 (ProcessInfo, Request, Response).
- `logs/`: `~/.geep/daemon.log` 및 개별 프로세스 로그 보관.

## 4. 기술 스택
- **Language**: Go (Golang)
- **CLI Library**: Cobra (github.com/spf13/cobra)
- **Communication**: Unix Domain Socket (net.Listen/Dial "unix")
- **Process Info**: `os/exec` 패키지 및 시스템 시그널 활용
