# ga-cli Roadmap

## 현재 구현 완료 (v0.1.0)

| 기능 | 상태 |
|------|------|
| `ga-cli properties` | ✅ |
| `ga-cli dau <property> [--days] [--all]` | ✅ |
| `ga-cli events <property> [--days] [--top]` | ✅ |
| `ga-cli countries <property> [--days]` | ✅ |
| `ga-cli platforms <property> [--days]` | ✅ |
| `ga-cli realtime <property>` | ✅ |
| `ga-cli auth login/status/logout` | ✅ |
| 출력 포맷 (table/json/csv) | ✅ |
| Service Account / OAuth2 / ADC 인증 | ✅ |
| 속성 별칭 (config.yaml aliases) | ✅ |

---

## P0 — 반드시 필요 (v0.2.0)

사용성과 인증 관련. 이것 없으면 실사용 불가.

### P0-1. OAuth Client ID 내장

**현재 문제**: 사용자가 GCP Console에서 직접 OAuth Client ID를 생성하고 `credentials.json`을 다운로드해야 함. 개발자가 아니면 진입장벽이 높음.

**해결**: Client ID/Secret을 바이너리에 내장하여 `ga-cli auth login`만 실행하면 바로 브라우저 로그인.

**구현**:
- `internal/auth/embedded.go` — Client ID, Secret 상수 정의
- `credentials.json` 파일 불필요
- `loadOAuthConfig()` 수정: 내장 Client ID 우선, `credentials.json` fallback

**필요 작업**:
- GCP 프로젝트에서 OAuth 2.0 Client ID (Desktop App) 1개 생성
- Analytics Data API + Admin API 활성화
- OAuth 동의 화면 설정 (외부/테스트 → 추후 게시)

---

### P0-2. `ga-cli config` 커맨드

**현재 문제**: 설정을 수동으로 `~/.ga-cli/config.yaml` 편집해야 함. 특히 aliases 등록이 불편.

**해결**:
```bash
ga-cli config init                          # 초기 설정 (인터랙티브)
ga-cli config set defaults.days 14          # 기본값 변경
ga-cli config get defaults.days             # 설정 조회
ga-cli config alias pastekeyboard 151869894 # 별칭 등록
ga-cli config alias --delete pastekeyboard  # 별칭 삭제
ga-cli config list                          # 전체 설정 출력
ga-cli config path                          # 설정 파일 경로 출력
```

**구현**:
- `cmd/config.go` — config 서브커맨드 + 하위 커맨드
- `internal/config/config.go` — Save() 메서드 추가

---

### P0-3. `ga-cli config init` 자동 셋업 (First-Run Experience)

**현재 문제**: 첫 사용 시 어떤 설정이 필요한지 모름.

**해결**: 첫 실행 시 자동 안내:
```
$ ga-cli properties
⚠ 설정 파일이 없습니다. 초기 설정을 시작합니다.

1. 인증 방식을 선택하세요:
   [1] Google 계정 로그인 (OAuth2) — 권장
   [2] Service Account 키 파일

2. 속성 목록을 조회합니다...
   발견된 속성:
   1. pastekeyboard (151869894)
   2. simplespend (384192617)
   ...

3. 별칭을 자동 등록했습니다.
   설정 파일: ~/.ga-cli/config.yaml
```

---

### P0-4. `dau --all` 정렬 + 정렬 플래그

**현재 문제**: `dau --all` 결과가 map 순회 순서라 매번 랜덤.

**해결**:
```bash
ga-cli dau --all                  # DAU 내림차순 (기본)
ga-cli dau --all --sort name      # 이름 오름차순
ga-cli dau --all --sort dau-asc   # DAU 오름차순
```

**구현**: `cmd/dau.go` — summaries 슬라이스 정렬 후 출력

---

### P0-5. 에러 메시지 개선

**현재 문제**: API 에러가 Google SDK 원문 그대로 출력. 사용자 친화적이지 않음.

**해결**:
| API 에러 | 개선된 메시지 |
|----------|---------------|
| 403 Permission Denied | `❌ 권한이 없습니다. GA4 속성에 대한 읽기 권한을 확인하세요.` |
| 401 Unauthenticated | `❌ 인증이 필요합니다. 'ga-cli auth login'을 실행하세요.` |
| 429 Quota Exceeded | `❌ API 호출 한도 초과. 잠시 후 다시 시도하세요.` |
| 404 Not Found | `❌ 속성을 찾을 수 없습니다: {propertyID}` |

**구현**: `internal/client/errors.go` — 에러 래퍼

---

## P1 — 있으면 좋은 기능 (v0.3.0)

UX 개선과 자동화 연동.

### P1-1. `ga-cli report` 종합 리포트

**설명**: 하나의 커맨드로 주요 지표 한눈에 확인.

```bash
ga-cli report pastekeyboard
```

출력:
```
PASTEKEYBOARD - Report (Last 30 days)

📊 DAU Summary
  Today: 6,010 | Avg: 5,950 | Peak: 6,495

📈 Top Events
  1. wordsCount        4,200,000
  2. screen_view       4,100,000
  3. user_engagement   3,900,000

🌍 Top Countries
  1. South Korea    4,500
  2. United States    300
  3. Japan            150

📱 Platforms
  iOS       5,200 (88.0% engaged)
  Android     200 (81.3% engaged)

⚡ Realtime: 140 users now
```

**구현**: `cmd/report.go` — 기존 client 메서드 조합 호출

---

### P1-2. `--slack` 플래그 (슬랙 웹훅 연동)

**설명**: 결과를 슬랙 채널로 전송.

```bash
ga-cli dau --all --slack                    # 기본 웹훅으로 전송
ga-cli report pastekeyboard --slack         # 리포트를 슬랙으로
ga-cli config set slack.webhook "https://hooks.slack.com/services/xxx"
```

**구현**:
- `internal/slack/webhook.go` — 웹훅 POST
- 글로벌 플래그 `--slack` 추가
- config.yaml에 `slack.webhook` 필드

---

### P1-3. `ga-cli compare` 기간 비교

**설명**: 두 기간의 지표를 비교.

```bash
ga-cli compare pastekeyboard --period week    # 이번 주 vs 지난 주
ga-cli compare pastekeyboard --period month   # 이번 달 vs 지난 달
```

출력:
```
PASTEKEYBOARD - Week over Week Comparison

METRIC          THIS WEEK    LAST WEEK    CHANGE
DAU (avg)       5,950        6,200        -4.0%
Events          4.2M         4.5M         -6.7%
Sessions        50,000       52,000       -3.8%
```

**구현**: `cmd/compare.go` + `internal/client/data.go`에 비교 메서드 추가

---

### P1-4. `--watch` 실시간 갱신 모드

**설명**: realtime 데이터를 N초 간격으로 갱신.

```bash
ga-cli realtime pastekeyboard --watch        # 10초 간격 (기본)
ga-cli realtime pastekeyboard --watch --interval 5  # 5초 간격
```

**구현**: `cmd/realtime.go` — `--watch` 플래그 + 터미널 클리어 + 루프

---

### P1-5. Shell Completion (자동완성)

**설명**: cobra 내장 completion + 속성 이름 자동완성.

```bash
ga-cli completion zsh > ~/.zfunc/_ga-cli     # zsh 자동완성 설치
ga-cli dau paste<TAB>                         # → pastekeyboard
```

**구현**: cobra의 `ValidArgsFunction` + config aliases에서 속성명 자동완성

---

### P1-6. `ga-cli funnel` 퍼널 분석

**설명**: 이벤트 퍼널을 정의하고 전환율 추적.

```bash
ga-cli funnel pastekeyboard \
  --steps "session_start,screen_view,wordsCount,ad_impression"
```

출력:
```
PASTEKEYBOARD - Funnel Analysis (Last 30 days)

STEP              USERS    DROP-OFF    RATE
session_start     75,000
screen_view       70,000    -5,000     93.3%
wordsCount        60,000   -10,000     85.7%
ad_impression     50,000   -10,000     83.3%

Overall: 66.7%
```

**구현**: GA4 Funnel API 또는 커스텀 계산

---

### P1-7. Homebrew 배포

**설명**: `brew install` 로 간편 설치.

```bash
brew tap techinpark/tap
brew install ga-cli
```

**구현**:
- GoReleaser 설정 (`.goreleaser.yml`)
- GitHub Actions CI/CD
- Homebrew tap 리포지토리 (`homebrew-tap`)

---

## 우선순위 요약

| 순위 | 기능 | 핵심 가치 |
|------|------|-----------|
| **P0-1** | OAuth Client ID 내장 | 진입장벽 제거 |
| **P0-2** | `config` 커맨드 | 설정 편의성 |
| **P0-3** | First-Run 자동 셋업 | 온보딩 |
| **P0-4** | `dau --all` 정렬 | 기본 UX |
| **P0-5** | 에러 메시지 개선 | 사용자 경험 |
| **P1-1** | `report` 종합 리포트 | 핵심 유틸리티 |
| **P1-2** | Slack 연동 | 자동화 |
| **P1-3** | `compare` 기간 비교 | 분석 |
| **P1-4** | `--watch` 실시간 | 모니터링 |
| **P1-5** | Shell Completion | 개발자 경험 |
| **P1-6** | `funnel` 퍼널 분석 | 고급 분석 |
| **P1-7** | Homebrew 배포 | 배포 |
