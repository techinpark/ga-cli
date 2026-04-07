# CLAUDE.md

이 파일은 Claude Code가 ga-cli 프로젝트를 이해하는 데 도움이 되는 정보를 담고 있습니다.

## 프로젝트 개요

ga-cli - Google Analytics 4 데이터를 CLI에서 조회하는 Go 도구입니다.
GA4 Admin API와 Data API를 래핑하여 속성 목록, DAU, 이벤트, 국가, 플랫폼, 실시간 데이터를 조회합니다.

## 빌드 명령어

```bash
# 빌드
make build

# 테스트
make test

# 테스트 + 커버리지
make test-coverage

# 린트
make lint

# 전체 검증 (build + test + lint)
make check

# 실행
make run ARGS="properties --format json"

# 설치 (GOBIN에 설치)
make install

# 정리
make clean
```

## 기술 스택

- **언어**: Go 1.22+
- **CLI 프레임워크**: cobra v1.8+
- **설정 관리**: viper
- **테이블 출력**: tablewriter
- **GA API**: google.golang.org/api/analyticsadmin/v1beta, analyticsdata/v1beta
- **인증**: Service Account JSON / Application Default Credentials (ADC)

## 아키텍처

```
ga-cli/
├── cmd/                    # Cobra 커맨드 (얇은 래퍼)
├── internal/
│   ├── client/            # GA API 클라이언트 인터페이스 + 구현
│   ├── formatter/         # 출력 포맷터 (table, json, csv)
│   ├── config/            # Viper 설정 관리
│   └── model/             # 도메인 모델
├── main.go
├── go.mod
├── Makefile
└── CLAUDE.md
```

### 핵심 설계 원칙

1. **인터페이스 우선**: client 패키지에 인터페이스 정의, 구현은 별도 파일
2. **의존성 주입**: 커맨드는 Dependencies 구조체를 통해 의존성 수신
3. **불변 패턴**: 입력 파라미터 변경 금지, 새 값 반환
4. **에러 래핑**: 모든 에러는 `fmt.Errorf("context: %w", err)` 형식으로 래핑
5. **Context 전파**: API 호출 함수는 항상 context.Context를 첫 번째 인자로 수신

## 코딩 컨벤션

### Go 스타일
- 패키지명: 소문자, 단일 단어
- 공개 타입: PascalCase
- 비공개: camelCase
- 약어: 전체 대문자 (GetDAU, propertyID)
- 인터페이스: -er 접미사 (Formatter, Reporter)

### 에러 처리
```go
// 올바른 에러 처리
result, err := client.GetDAU(ctx, propertyID, dateRange)
if err != nil {
    return fmt.Errorf("failed to get DAU for property %s: %w", propertyID, err)
}

// 에러 메시지: 소문자 시작, 마침표 없음
return fmt.Errorf("invalid property ID: %s", id)
```

### 커맨드 플래그 규칙
- `--days`, `-d`: 조회 기간 (일, 기본값: 30)
- `--top`, `-t`: 상위 N개 (기본값: 20)
- `--format`, `-f`: 출력 형식 (table/json/csv, 기본값: table)
- `--credentials`, `-c`: 서비스 계정 키 경로
- `--all`: 전체 속성 대상

### 테스트 규칙
- TDD 필수 (RED -> GREEN -> IMPROVE)
- 테이블 기반 테스트 사용
- 외부 API는 반드시 mock
- 커버리지 80% 이상
- `go test -race` 통과 필수

### 커밋 메시지
- Conventional Commits 형식
- `feat:` 새 기능, `fix:` 버그 수정, `test:` 테스트
- `refactor:` 리팩토링, `docs:` 문서, `chore:` 빌드/설정

## 설정 파일

### ~/.ga-cli/config.yaml
```yaml
credentials: /path/to/serviceAccountKey.json

aliases:
  pastekeyboard: "151869894"
  simplespend: "384192617"
  oneshot-note: "508319520"
  moments: "302237145"

defaults:
  days: 30
  top: 20
  output: table
```

### 환경 변수
- `GA_SERVICE_ACCOUNT_KEY`: 서비스 계정 키 파일 경로
- `GA_CLI_DAYS`: 기본 조회 기간
- `GA_CLI_OUTPUT`: 기본 출력 형식

### 인증 우선순위
1. `--credentials` 플래그
2. `GA_SERVICE_ACCOUNT_KEY` 환경변수
3. `~/.ga-cli/config.yaml`의 `credentials` 필드
4. Application Default Credentials (ADC)

## 체크리스트

### 새 커맨드 추가 시

| 단계 | 파일 | 작업 |
|------|------|------|
| 1 | `internal/model/model.go` | 응답 모델 타입 정의 |
| 2 | `internal/client/client.go` | 인터페이스에 메서드 추가 |
| 3 | `internal/client/*_test.go` | 클라이언트 테스트 작성 (RED) |
| 4 | `internal/client/*.go` | 클라이언트 구현 (GREEN) |
| 5 | `internal/formatter/formatter.go` | Formatter 인터페이스에 메서드 추가 |
| 6 | `internal/formatter/*_test.go` | 포맷터 테스트 작성 |
| 7 | `internal/formatter/*.go` | 포맷터 구현 |
| 8 | `cmd/*_test.go` | 커맨드 테스트 작성 |
| 9 | `cmd/*.go` | Cobra 커맨드 구현 |
| 10 | `cmd/root.go` | root 커맨드에 서브커맨드 등록 |

### 코드 리뷰 전 확인

```bash
make check  # build + test + lint 전체 실행
```

- [ ] go build 통과
- [ ] go test 통과 (커버리지 80%+)
- [ ] go vet 통과
- [ ] golangci-lint 통과
- [ ] go test -race 통과

## 에이전트 팀

| 에이전트 | 역할 | 사용 시점 |
|----------|------|-----------|
| `ga-orchestrator` | 전체 워크플로우 조율 | 새 커맨드/기능 구현 |
| `ga-architect` | 아키텍처 설계 | 인터페이스/패키지 설계 |
| `ga-coder` | Go 코드 작성 | 구현 코드 작성 |
| `ga-tester` | TDD 테스트 작성 | 테스트 작성/커버리지 관리 |
| `ga-reviewer` | 코드 리뷰 | 코드 변경 후 리뷰 |

## 아키텍처 결정 기록 (ADR)

(프로젝트 진행 중 결정사항 여기에 추가)
