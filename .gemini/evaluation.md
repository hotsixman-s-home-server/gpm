# GEEP Codebase Evaluation Report

본 보고서는 GEEP(Go Process Manager) 프로젝트의 소스 코드를 아키텍처, 성능, 안정성, 유지보수성 측면에서 비판적으로 분석한 결과입니다.

## 1. 아키텍처 및 디자인 패턴 (Architecture & Design)

### 긍정적 측면
- **CLI/Daemon 분리**: 현대적인 프로세스 매니저의 표준적인 구조를 잘 따르고 있으며, UDS(Unix Domain Socket)를 통한 IPC 구현이 적절합니다.
- **인터페이스 활용**: `types.PMInterface`, `types.ServerInterface` 등을 정의하여 모듈 간 의존성을 추상화하려 노력한 흔적이 보입니다.

### 개선 필요 사항
- **상태 관리의 파편화**: 프로세스 정보가 `PM` 구조체의 `map`과 `slice`에 중복 저장(`process`, `processArr`)되어 있어, 데이터 일관성 유지가 어려울 수 있습니다.
- **데이터베이스 영속성 미흡**: `module/database/db.go`의 주요 로직(로그 파일 업데이트 등)이 주석 처리되어 있어, 데몬 재시작 시 프로세스 상태 복구가 불가능한 상태입니다.

## 2. 자원 관리 및 안정성 (Resource Management & Stability)

### 심각한 문제: 파일 디스크립터 누수 (Critical: FD Leak)
- **`module/logger/logger.go`**: `newLogFile()` 함수에서 새로운 로그 파일을 열 때, 기존에 열려있던 `this.logFile`과 `this.errorFile`을 `Close()`하지 않고 덮어씁니다. 로그 로테이션이 발생할 때마다 파일 디스크립터가 누수되어 장기 실행 시 데몬이 크래시될 위험이 매우 높습니다.

### 프로세스 정리 로직
- **`PMProcess.clean()`**: 파이프를 닫는 로직은 있으나, `cmd.Wait()` 이후에도 `os.Process` 자원이 완전히 정리되었는지 확인하는 절차가 더 견고해야 합니다.

## 3. 동시성 및 동기화 (Concurrency & Synchronization)

### 레이스 컨디션 위험
- **`module/server/server.go`**: `Broadcast` 함수에서 각 클라이언트에 대해 `go func() { client.conn.Write(...) }()`를 실행합니다. `net.Conn.Write`는 기본적으로 스레드 안전(Thread-safe)하지 않으므로, 동일 클라이언트에 여러 메시지가 동시에 전송될 경우 데이터가 섞이거나 에러가 발생할 수 있습니다. 클라이언트별 전송 채널(queue)을 두는 방식이 권장됩니다.
- **`pm.processMutex`**: `initProcess` 내부에서 뮤텍스를 잡은 채로 외부 입출력이나 시간이 걸리는 작업을 수행하는 구간이 있습니다. 잠금 범위를 최소화해야 합니다.

## 4. 성능 최적화 (Performance)

### 로그 Tail 로직의 비효율성
- **`module/logger/logger.go` (tailLines)**: 파일을 끝에서부터 한 바이트씩 읽으며 메모리에 버퍼를 쌓는 방식입니다. 로그 파일이 커질 경우 CPU와 메모리에 과도한 부하를 줄 수 있습니다. `io.Seek`을 사용하여 큰 블록 단위로 읽는 방식이 더 효율적입니다.

## 5. 코드 품질 및 관행 (Code Quality & Conventions)

- **에러 처리**: 많은 곳에서 에러를 단순히 `mainLogger.Errorln(err)`으로 출력하고 흐름을 계속 진행합니다. 치명적인 에러와 무시 가능한 에러를 구분하여 적절한 복구 로직이나 조기 리턴(Early Return)이 필요합니다.
- **매직 스트링**: 프로세스 상태(`running`, `stop`, `error`)가 문자열 리터럴로 산재해 있습니다. `const`를 사용한 열거형으로 관리해야 합니다.

## 종합 의견
GEEP은 기능적으로 동작하는 프로토타입 수준을 넘어섰으나, **데몬으로서의 장기적인 안정성(특히 자원 누수 및 동시성)** 측면에서 보완이 시급합니다. 특히 로그 로테이션 시의 파일 클로즈 누락은 반드시 수정되어야 할 우선순위 1순위 과제입니다.

---

## 6. 구체적 개선 가이드 (Actionable Improvements)

### 6.1 파일 디스크립터 누수 해결 (`module/logger/logger.go`)
```go
func (this *Logger) newLogFile() error {
    // 개선: 기존 파일을 명확히 닫아주어야 함
    if this.logFile != nil {
        this.logFile.Close()
    }
    if this.errorFile != nil {
        this.errorFile.Close()
    }

    // ... 새 파일 오픈 로직 ...
}
```

### 6.2 안전한 네트워크 전송 (`module/server/server.go`)
```go
// ServerSideClient 구조체 개선
type ServerSideClient struct {
    conn     net.Conn
    name     string
    reader   *bufio.Reader
    sendChan chan []byte // 전송 전용 채널 추가
}

// 전용 쓰기 루프 도입
func (c *ServerSideClient) writeLoop() {
    for msg := range c.sendChan {
        c.conn.Write(msg)
    }
}

// Broadcast는 채널에 넣기만 함
func (server *Server) Broadcast(name string, JSON []byte) {
    server.mutex.Lock()
    defer server.mutex.Unlock()
    for _, client := range server.client {
        if client.name == name {
            select {
            case client.sendChan <- append(JSON, '\n'):
            default: // 채널이 꽉 찼을 경우에 대한 처리
            }
        }
    }
}
```

### 6.3 로그 Tail 성능 최적화
- `os.Open` 후 `f.Seek(0, io.SeekEnd)`를 사용하여 파일 끝에서부터 역순으로 큰 버퍼(예: 4KB) 단위로 읽으며 `\n` 개수를 세는 방식으로 변경합니다.

### 6.4 상태 값 상수화 (`module/pm/pm.go`)
```go
type PMProcessStatus string

const (
    StatusRunning PMProcessStatus = "running"
    StatusStop    PMProcessStatus = "stop"
    StatusError   PMProcessStatus = "error"
)
```

### 6.5 에러 처리 및 조기 리턴 (Early Return)
- `if err != nil { logger.Error(err); return err }` 패턴을 일관되게 적용하여 예외 상황에서 불필요한 로직이 실행되지 않도록 강제합니다.

