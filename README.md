# ga-cli

Google Analytics 4 데이터를 터미널에서 빠르게 조회하는 CLI 도구.

여러 GA4 속성의 DAU, 이벤트, 국가별 사용자, 플랫폼 통계, 실시간 데이터를 한 곳에서 확인할 수 있습니다.

## Install

### Homebrew (권장)

```bash
brew tap techinpark/tap
brew install ga-cli
```

### Go Install

```bash
go install github.com/techinpark/ga-cli@latest
```

### 소스에서 빌드

```bash
git clone https://github.com/techinpark/ga-cli.git
cd ga-cli
make install
```

## Quick Start

```bash
# 1. Google 계정 로그인
ga-cli auth login

# 2. 속성 목록 확인
ga-cli properties

# 3. DAU 조회
ga-cli dau my-app --days 7
```

## Authentication

```bash
ga-cli auth login     # 브라우저에서 Google 계정 로그인 (권장)
ga-cli auth status    # 인증 상태 확인
ga-cli auth logout    # 로그아웃
```

여러 Google 계정을 등록하고 전환할 수 있습니다:

```bash
ga-cli auth login --account work       # 회사 계정 추가
ga-cli auth login --account personal   # 개인 계정 추가
ga-cli auth list                       # 등록된 계정 목록
ga-cli auth switch work                # 활성 계정 전환
```

### 인증 우선순위

| 순위 | 방식 | 설명 |
|------|------|------|
| 1 | Service Account | `--credentials path/to/key.json` 또는 config.yaml |
| 2 | OAuth2 (권장) | `ga-cli auth login` → 브라우저 로그인 |
| 3 | ADC | `gcloud auth application-default login` |

> **소스에서 직접 빌드한 경우**, OAuth2 credentials가 내장되어 있지 않습니다.
> `~/.ga-cli/credentials.json`을 직접 제공하거나 ADC를 사용하세요.

## Commands

### `properties` — 속성 목록

```bash
ga-cli properties
```

```
ALIAS               PROPERTY ID    DISPLAY NAME
my-app              123456789      my-app
my-blog             987654321      my-blog-a1b2c
my-shop             555666777      my-shop-d3e4f
```

### `dau` — 일일 활성 사용자

```bash
ga-cli dau my-app --days 7
```

```
MY-APP - Daily Active Users (Last 7 days)

DATE         DAU      CHANGE
2026-04-01   6,495
2026-04-02   6,430    -1.0%
2026-04-03   6,335    -1.5%
2026-04-04   6,309    -0.4%
2026-04-05   5,906    -6.4%
2026-04-06   6,010    +1.8%
2026-04-07   4,162    (today)

Avg: 5,950
```

전체 속성 한눈에:

```bash
ga-cli dau --all
```

```
PROPERTY              DAU (TODAY)
my-app                6,010
my-blog                 249
my-shop                  97
```

### `events` — 이벤트 분석

```bash
ga-cli events my-app --top 5 --days 30
```

```
MY-APP - Top Events (Last 30 days)

#   EVENT                    COUNT       USERS
1   page_view                4,200,000   70,000
2   screen_view              4,100,000   75,000
3   user_engagement          3,900,000   74,000
4   click                    500,000     50,000
5   session_start            490,000     75,000
```

### `countries` — 국가별 사용자

```bash
ga-cli countries my-app
```

```
MY-APP - Users by Country (Last 30 days)

#   COUNTRY         USERS    SESSIONS   VIEWS/USER
1   South Korea     4,500    45,000     25.3
2   United States   300      1,200      8.5
3   Japan           150      600        12.1
```

### `platforms` — 플랫폼 분석

```bash
ga-cli platforms my-app
```

```
MY-APP - Platform Breakdown (Last 30 days)

PLATFORM    USERS    SESSIONS    ENGAGED    RATE     SESSIONS/USER
iOS         5,200    50,000      44,000     88.0%    9.6
Android     200      800         650        81.3%    4.0
```

### `realtime` — 실시간 데이터

```bash
ga-cli realtime my-app
```

```
MY-APP - Realtime

Active Users: 140

#   EVENT                    COUNT
1   page_view                1,299
2   screen_view              1,030
3   user_engagement          958
```

## Output Formats

```bash
ga-cli dau my-app --format table   # 기본값
ga-cli dau my-app --format json    # JSON (파이프용)
ga-cli dau my-app --format csv     # CSV
```

JSON 출력을 `jq`와 조합:

```bash
ga-cli dau my-app --format json | jq '.[0].active_users'
```

## Configuration

`~/.ga-cli/config.yaml`:

```yaml
credentials: /path/to/serviceAccountKey.json

aliases:
  my-app: "123456789"
  my-blog: "987654321"
  my-shop: "555666777"

defaults:
  days: 30
  top: 20
  output: table
```

속성 별칭을 등록하면 property ID 대신 이름으로 조회할 수 있습니다:

```bash
ga-cli dau my-app      # alias 사용
ga-cli dau 123456789   # property ID 직접 사용
```

## Global Flags

| 플래그 | 단축 | 기본값 | 설명 |
|--------|------|--------|------|
| `--credentials` | `-c` | config.yaml | 서비스 계정 키 경로 |
| `--format` | `-f` | table | 출력 형식 (table/json/csv) |
| `--config` | | ~/.ga-cli/config.yaml | 설정 파일 경로 |

## Tech Stack

- Go + [cobra](https://github.com/spf13/cobra) + [viper](https://github.com/spf13/viper)
- Google Analytics Admin API v1beta
- Google Analytics Data API v1beta

## License

MIT
