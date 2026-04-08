# ga-cli

A command-line tool to query Google Analytics 4 data from the terminal.

View DAU, events, countries, platforms, realtime stats, and comparison reports across multiple GA4 properties — all without leaving your terminal.

## Install

### Homebrew

```bash
brew tap techinpark/tap
brew install ga-cli
```

### Go Install

```bash
go install github.com/techinpark/ga-cli@latest
```

### Build from Source

```bash
git clone https://github.com/techinpark/ga-cli.git
cd ga-cli
make install
```

## Quick Start

```bash
# 1. Log in with your Google account
ga auth login

# 2. List your GA4 properties
ga properties

# 3. Check DAU
ga dau my-app --days 7
```

## Authentication

```bash
ga auth login      # Browser-based Google login (recommended)
ga auth status     # Check auth status
ga auth logout     # Log out
```

Multiple accounts are supported:

```bash
ga auth login --account work       # Add work account
ga auth login --account personal   # Add personal account
ga auth list                       # List registered accounts
ga auth switch work                # Switch active account
```

### Auth Priority

| Priority | Method | Description |
|----------|--------|-------------|
| 1 | Service Account | `--credentials path/to/key.json` or config.yaml |
| 2 | OAuth2 (recommended) | `ga auth login` → browser login |
| 3 | ADC | `gcloud auth application-default login` |

> **Building from source**: OAuth2 credentials are not embedded by default.
> Provide `~/.ga-cli/credentials.json` or use ADC.

## Commands

### `properties` — List GA4 Properties

```bash
ga properties
```

```
ALIAS               PROPERTY ID    DISPLAY NAME
my-app              123456789      my-app
my-blog             987654321      my-blog-a1b2c
my-shop             555666777      my-shop-d3e4f
```

### `dau` — Daily Active Users

```bash
ga dau my-app --days 7
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

All properties at a glance:

```bash
ga dau --all                  # sorted by DAU descending (default)
ga dau --all --sort name      # sorted by name
ga dau --all --sort dau-asc   # sorted by DAU ascending
```

```
PROPERTY              DAU (TODAY)
my-app                6,010
my-blog                 249
my-shop                  97
```

### `events` — Event Analysis

```bash
ga events my-app --top 5 --days 30
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

### `countries` — Users by Country

```bash
ga countries my-app
```

```
MY-APP - Users by Country (Last 30 days)

#   COUNTRY         USERS    SESSIONS   VIEWS/USER
1   South Korea     4,500    45,000     25.3
2   United States   300      1,200      8.5
3   Japan           150      600        12.1
```

### `platforms` — Platform Breakdown

```bash
ga platforms my-app
```

```
MY-APP - Platform Breakdown (Last 30 days)

PLATFORM    USERS    SESSIONS    ENGAGED    RATE     SESSIONS/USER
iOS         5,200    50,000      44,000     88.0%    9.6
Android     200      800         650        81.3%    4.0
```

### `realtime` — Realtime Data

```bash
ga realtime my-app
```

```
MY-APP - Realtime

Active Users: 140

#   EVENT                    COUNT
1   page_view                1,299
2   screen_view              1,030
3   user_engagement          958
```

### `report` — Aggregated Reports

```bash
ga report daily my-app       # DAU 7d + events + platforms + realtime
ga report weekly my-app      # DAU 14d + events + countries + platforms
ga report compare my-app     # Day-over-day + week-over-week comparison
ga report daily --all        # All properties
```

**Compare report output:**

```
MY-APP - Comparison Report

Day over Day (Today vs Yesterday)
METRIC    TODAY      YESTERDAY  CHANGE
DAU       6,010      6,495      -7.5%
Events    4,200,000  4,500,000  -6.7%
Sessions  50,000     52,000     -3.8%

Week over Week (This Week vs Last Week)
METRIC    THIS WEEK   LAST WEEK   CHANGE
DAU       5,950       6,200       -4.0%
Events    29,400,000  31,500,000  -6.7%
Sessions  350,000     364,000     -3.8%
```

### `config` — Configuration Management

```bash
ga config list                        # Show all settings
ga config get defaults.days           # Get a value
ga config set defaults.days 14        # Set a value
ga config alias my-app 123456789      # Register alias
ga config alias --delete my-app       # Remove alias
ga config path                        # Show config file path
```

## Output Formats

```bash
ga dau my-app --format table   # default
ga dau my-app --format json    # JSON (pipe-friendly)
ga dau my-app --format csv     # CSV
```

Combine JSON output with `jq`:

```bash
ga dau my-app --format json | jq '.[0].active_users'
ga report compare my-app --format json | jq '.day_over_day'
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

Use aliases instead of property IDs:

```bash
ga dau my-app      # alias
ga dau 123456789   # property ID
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--credentials` | `-c` | config.yaml | Service account key path |
| `--format` | `-f` | table | Output format (table/json/csv) |
| `--config` | | ~/.ga-cli/config.yaml | Config file path |
| `--account` | | active | Account to use |

## Development

```bash
make build           # Build binary
make test            # Run tests
make test-coverage   # Coverage report
make check           # build + vet + test + lint
make run ARGS="dau my-app --days 7"
```

## Tech Stack

- Go + [cobra](https://github.com/spf13/cobra) + [viper](https://github.com/spf13/viper)
- Google Analytics Admin API v1beta
- Google Analytics Data API v1beta

## License

MIT
