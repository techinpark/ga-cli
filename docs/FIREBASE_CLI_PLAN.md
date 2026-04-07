# fb-cli: Firebase Admin CLI Tool 구현 계획

## 개요

Firebase Admin SDK(Go)를 래핑하는 CLI 도구.
여러 Firebase 프로젝트의 사용자 관리, Firestore 조회, Remote Config, Cloud Messaging, Crashlytics 현황 등을 터미널에서 빠르게 조회/관리할 수 있도록 한다.

ga-cli와 동일한 아키텍처(cobra + viper + 멀티 계정)를 기반으로 구축.

## 사용 예시

```bash
# 프로젝트 목록
fb projects

# 사용자 관리
fb users list --project my-app --limit 20        # 사용자 목록
fb users get --email user@example.com             # 이메일로 조회
fb users get --uid abc123                         # UID로 조회
fb users count --project my-app                   # 총 사용자 수
fb users disable --uid abc123                     # 사용자 비활성화
fb users export --project my-app --format csv     # 사용자 내보내기

# Firestore
fb firestore list --collection users --limit 10   # 문서 목록
fb firestore get --path users/abc123              # 문서 조회
fb firestore query --collection orders \
  --where "status=pending" --limit 5              # 쿼리
fb firestore count --collection users             # 문서 수

# Remote Config
fb remoteconfig list                              # 전체 파라미터 목록
fb remoteconfig get my_feature_flag               # 특정 파라미터 값
fb remoteconfig set my_feature_flag true          # 파라미터 수정
fb remoteconfig history --limit 10                # 변경 이력

# Cloud Messaging
fb messaging send --topic news \
  --title "Breaking" --body "Hello World"         # 토픽 메시지 전송
fb messaging send --token <device-token> \
  --title "Hi" --body "Test"                      # 단일 기기 전송

# Analytics (Crashlytics 현황)
fb crashlytics summary --project my-app           # 최근 크래시 요약
fb crashlytics issues --project my-app --top 10   # 상위 크래시 이슈

# 인증
fb auth login                                     # Google 계정 로그인
fb auth login --account work                      # 멀티 계정
fb auth list                                      # 계정 목록
fb auth switch work                               # 계정 전환

# 출력 포맷
fb users list --format json                       # JSON
fb firestore list --collection users --format csv # CSV
```

## 기술 스택

| 항목 | 선택 | 이유 |
|------|------|------|
| 언어 | Go | 단일 바이너리, 빠른 실행, ga-cli와 동일 |
| CLI 프레임워크 | `cobra` | Go CLI 표준 |
| 설정 관리 | `viper` | YAML 설정, cobra 통합 |
| 테이블 출력 | 커스텀 renderer | ga-cli에서 검증됨 |
| Firebase Admin SDK | `firebase.google.com/go/v4` | 공식 Go SDK |
| Firestore | `cloud.google.com/go/firestore` | 공식 Go SDK |
| 인증 | OAuth2 / Service Account | ga-cli와 동일 패턴 |

## 프로젝트 구조

```
fb-cli/
├── main.go
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
├── .goreleaser.yml
├── .github/workflows/release.yml
├── cmd/
│   ├── root.go                 # 루트 커맨드, 글로벌 플래그
│   ├── deps.go                 # Dependencies 구조체
│   ├── auth.go                 # fb auth login/logout/list/switch
│   ├── projects.go             # fb projects
│   ├── users.go                # fb users list/get/count/disable/export
│   ├── firestore.go            # fb firestore list/get/query/count
│   ├── remoteconfig.go         # fb remoteconfig list/get/set/history
│   ├── messaging.go            # fb messaging send
│   └── crashlytics.go          # fb crashlytics summary/issues
├── internal/
│   ├── auth/
│   │   └── auth.go             # OAuth2 + 멀티 계정 (ga-cli 재사용)
│   ├── client/
│   │   ├── client.go           # 인터페이스 정의
│   │   ├── firebase.go         # Firebase App 초기화
│   │   ├── users.go            # Auth 사용자 관리
│   │   ├── firestore.go        # Firestore 클라이언트
│   │   ├── remoteconfig.go     # Remote Config 클라이언트
│   │   ├── messaging.go        # FCM 클라이언트
│   │   └── crashlytics.go      # Crashlytics (REST API)
│   ├── config/
│   │   └── config.go           # 설정 파일 관리
│   ├── formatter/
│   │   ├── formatter.go        # 포맷터 인터페이스 (ga-cli 재사용 가능)
│   │   ├── table.go
│   │   ├── json.go
│   │   └── csv.go
│   └── model/
│       └── model.go            # 도메인 모델
└── docs/
    └── ROADMAP.md
```

## 설정 파일

`~/.fb-cli/config.yaml`:

```yaml
active_account: default

# 프로젝트 별칭
aliases:
  my-app: "my-app-12345"
  my-blog: "my-blog-67890"

# 기본값
defaults:
  project: my-app              # 기본 프로젝트
  limit: 20                    # 기본 조회 개수
  output: table                # 기본 출력 포맷
```

## 인증

ga-cli와 동일한 멀티 계정 패턴:

```
~/.fb-cli/
├── config.yaml
├── credentials.json           # OAuth Client ID (내장 가능)
└── accounts/
    ├── default.json
    ├── work.json
    └── personal.json
```

### 필요 스코프
```go
var Scopes = []string{
    "https://www.googleapis.com/auth/firebase",
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/datastore",
}
```

## 커맨드별 상세 설계

### 1. `fb projects`

프로젝트 목록 조회 (Firebase Management API).

```
$ fb projects

ALIAS        PROJECT ID           DISPLAY NAME
my-app       my-app-12345         My Application
my-blog      my-blog-67890        My Blog
(none)       test-project-99999   Test Project
```

- API: Firebase Management REST API `GET /v1beta1/projects`
- config aliases와 매칭

---

### 2. `fb users` (Firebase Authentication)

```
$ fb users list --project my-app --limit 5

#   UID                     EMAIL                   CREATED         LAST SIGN-IN    STATUS
1   abc123def456            user1@gmail.com         2026-01-15      2026-04-07      active
2   xyz789ghi012            user2@example.com       2025-12-01      2026-04-05      active
3   mno345pqr678            user3@test.com          2025-06-20      2025-12-31      disabled
...

Total: 15,234 users
```

```
$ fb users get --email user1@gmail.com

UID:            abc123def456
Email:          user1@gmail.com
Display Name:   John Doe
Phone:          +82-10-1234-5678
Provider:       google.com, password
Created:        2026-01-15 09:30:00
Last Sign-In:   2026-04-07 14:22:00
Status:         active
Email Verified: true
```

- SDK: `firebase.google.com/go/v4/auth`
- `ListUsers()` — 페이지네이션 지원
- `GetUser(uid)`, `GetUserByEmail(email)`
- `UpdateUser(uid, params)` — disable/enable
- `Users.Iterate()` — export용 전체 순회

---

### 3. `fb firestore`

```
$ fb firestore list --collection users --limit 3

#   DOCUMENT ID       FIELDS
1   abc123            {"name": "John", "age": 30, "active": true}
2   def456            {"name": "Jane", "age": 25, "active": true}
3   ghi789            {"name": "Bob", "age": 35, "active": false}

Total: 3 / 15,234 documents
```

```
$ fb firestore get --path users/abc123

Document: users/abc123
Created:  2026-01-15T09:30:00Z
Updated:  2026-04-07T14:22:00Z

FIELD          TYPE      VALUE
name           string    John
age            number    30
email          string    john@example.com
active         boolean   true
tags           array     ["admin", "premium"]
address        map       {"city": "Seoul", "country": "KR"}
```

```
$ fb firestore query --collection orders --where "status=pending" --limit 5

#   DOCUMENT ID    status     amount    created_at
1   ord_001        pending    29,900    2026-04-07
2   ord_002        pending    15,000    2026-04-07
3   ord_003        pending    49,900    2026-04-06
```

- SDK: `cloud.google.com/go/firestore`
- `Collection().Documents()` — 목록
- `Doc().Get()` — 단일 문서
- `Collection().Where().Limit().Documents()` — 쿼리
- `Collection().Count()` — 개수 (Aggregation query)

---

### 4. `fb remoteconfig`

```
$ fb remoteconfig list

PARAMETER                TYPE       DEFAULT VALUE    CONDITIONS
enable_dark_mode         boolean    false            iOS: true
max_retry_count          number     3
welcome_message          string     "Hello!"         ko: "안녕하세요!"
feature_new_ui           boolean    false            beta_users: true
```

```
$ fb remoteconfig history --limit 5

#   VERSION    UPDATED BY              UPDATED AT           DESCRIPTION
1   42         user@example.com        2026-04-07 10:00     Update feature flags
2   41         admin@example.com       2026-04-05 15:30     Increase retry count
3   40         user@example.com        2026-04-01 09:00     Add dark mode
```

- SDK: `firebase.google.com/go/v4/remoteconfig` (또는 REST API)
- `Get()` — 전체 템플릿
- `Update()` — 파라미터 수정
- `ListVersions()` — 변경 이력

---

### 5. `fb messaging`

```
$ fb messaging send --topic news --title "Breaking News" --body "Important update"

Message sent successfully.
Message ID: projects/my-app/messages/msg_12345
Topic: news
```

```
$ fb messaging send --token "device-token-here" --title "Hello" --body "Test message"

Message sent successfully.
Message ID: projects/my-app/messages/msg_67890
Target: device token
```

- SDK: `firebase.google.com/go/v4/messaging`
- `Send(message)` — 단일 메시지 전송
- Topic, Token, Condition 지원
- `--dry-run` 플래그로 검증만

---

### 6. `fb crashlytics`

```
$ fb crashlytics summary --project my-app

MY-APP - Crashlytics Summary (Last 7 days)

METRIC              VALUE
Crash-free users    98.5%
Total crashes       1,234
Affected users      456
Top device          iPhone 15 Pro
Top OS              iOS 17.4
```

```
$ fb crashlytics issues --project my-app --top 5

#   ISSUE ID    TITLE                                    EVENTS    USERS    STATUS
1   is_001      NullPointerException in MainActivity     500       200      open
2   is_002      IndexOutOfBounds in ListAdapter          300       150      open
3   is_003      NetworkError in ApiClient                200       100      open
4   is_004      ConcurrentModification in Cache          100       50       resolved
5   is_005      OOM in ImageLoader                       80        40       open
```

- API: Crashlytics REST API (Firebase 공식 Go SDK에 Crashlytics 미포함)
- `GET /v1beta1/projects/{id}/issues`
- Google API 클라이언트로 직접 호출

---

## 글로벌 플래그

| 플래그 | 단축 | 기본값 | 설명 |
|--------|------|--------|------|
| `--project` | `-p` | config default | Firebase 프로젝트 ID/별칭 |
| `--credentials` | `-c` | config.yaml | 서비스 계정 키 경로 |
| `--format` | `-f` | table | 출력 형식 (table/json/csv) |
| `--account` | | active | 사용할 계정 |
| `--config` | | ~/.fb-cli/config.yaml | 설정 파일 경로 |

## ga-cli와 공유 가능한 코드

| 패키지 | 재사용 | 방법 |
|--------|--------|------|
| `internal/auth/` | 100% | 거의 동일 (스코프만 변경) |
| `internal/formatter/` | 80% | 인터페이스 확장, table/json/csv 로직 동일 |
| `internal/config/` | 90% | 구조 동일 (필드만 다름) |
| `cmd/auth.go` | 95% | 멀티 계정 로직 동일 |
| `.goreleaser.yml` | 100% | 바이너리명만 변경 |
| `Makefile` | 100% | 바이너리명만 변경 |

→ 향후 공통 모듈을 별도 Go 패키지로 추출 가능 (`github.com/techinpark/cli-common`)

## 구현 순서

### Phase 1: 기본 구조 + 핵심 커맨드
1. Go 모듈 초기화, cobra 셋업
2. 인증 (ga-cli auth 패키지 포팅)
3. 설정 파일 로드
4. `fb projects` 구현
5. `fb users list/get/count` 구현

### Phase 2: Firestore + Remote Config
6. `fb firestore list/get/query/count` 구현
7. `fb remoteconfig list/get/set` 구현
8. `fb remoteconfig history` 구현

### Phase 3: Messaging + Crashlytics
9. `fb messaging send` 구현
10. `fb crashlytics summary/issues` 구현

### Phase 4: 편의 기능
11. `fb users export` (CSV/JSON 내보내기)
12. `fb users disable/enable` (상태 변경)
13. `fb firestore delete` (문서 삭제, --dry-run 필수)
14. JSON/CSV 출력 지원
15. Homebrew 배포

### Phase 5: 고급 기능 (선택)
16. `fb firestore watch --path users/abc123` (실시간 리스너)
17. `fb firestore import/export` (데이터 마이그레이션)
18. `fb storage list/upload/download` (Cloud Storage)
19. `fb functions list/logs` (Cloud Functions)
20. 인터랙티브 모드 (`fb shell`)

## Firebase Admin SDK 초기화 패턴

```go
import (
    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/auth"
    "google.golang.org/api/option"
)

func NewFirebaseApp(projectID string, opts []option.ClientOption) (*firebase.App, error) {
    config := &firebase.Config{ProjectID: projectID}
    return firebase.NewApp(context.Background(), config, opts...)
}

// 각 서비스 클라이언트
authClient, _ := app.Auth(ctx)          // Authentication
firestoreClient, _ := app.Firestore(ctx) // Firestore
messagingClient, _ := app.Messaging(ctx) // FCM
```

## 보안 고려사항

1. **쓰기 작업은 확인 필요**: `users disable`, `remoteconfig set`, `firestore delete` 등은 `--yes` 플래그 없으면 확인 프롬프트
2. **`--dry-run`**: messaging send, firestore delete에 지원
3. **감사 로그**: 쓰기 작업 시 로컬 로그 기록 (`~/.fb-cli/audit.log`)
4. **Rate Limiting**: Firebase Admin SDK 자체 제한 준수

## 차별점 (vs Firebase CLI)

| | Firebase CLI (`firebase`) | fb-cli |
|---|---|---|
| 초점 | 배포/호스팅/에뮬레이터 | 운영/조회/분석 |
| 사용자 관리 | 미지원 | 목록/조회/비활성화/내보내기 |
| Firestore 조회 | 에뮬레이터만 | 프로덕션 직접 조회 |
| 출력 포맷 | 고정 | table/json/csv |
| 멀티 계정 | 미지원 | 지원 |
| 파이프라인 | 제한적 | JSON/CSV로 자유롭게 |
| 설치 크기 | ~200MB (Node.js) | ~20MB (Go 바이너리) |
