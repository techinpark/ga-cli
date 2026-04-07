BINARY_NAME=ga-cli
VERSION?=0.1.0
BUILD_DIR=./build
OAUTH_CLIENT_ID?=
OAUTH_CLIENT_SECRET?=
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X github.com/techinpark/ga-cli/internal/auth.oauthClientID=$(OAUTH_CLIENT_ID) -X github.com/techinpark/ga-cli/internal/auth.oauthClientSecret=$(OAUTH_CLIENT_SECRET)"

.PHONY: build test lint run install clean check test-coverage test-race fmt vet help

## build: 바이너리 빌드
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

## test: 테스트 실행
test:
	go test ./... -v -count=1

## test-coverage: 커버리지 리포트 생성
test-coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic -count=1
	go tool cover -func=coverage.out
	@echo "---"
	@echo "HTML report: go tool cover -html=coverage.out -o coverage.html"

## test-race: 레이스 컨디션 검사
test-race:
	go test ./... -race -count=1

## lint: golangci-lint 실행
lint:
	golangci-lint run ./...

## vet: go vet 실행
vet:
	go vet ./...

## fmt: 코드 포맷팅
fmt:
	gofmt -s -w .
	goimports -w .

## run: 빌드 후 실행
run: build
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## install: GOBIN에 설치
install:
	go install $(LDFLAGS) .

## clean: 빌드 아티팩트 정리
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## check: 전체 검증 (build + vet + test-race + lint)
check: build vet test-race lint
	@echo "All checks passed."

## help: 도움말 표시
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
