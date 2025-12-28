# AAMI Development Guide

AAMI 프로젝트 개발 환경 설정 및 개발 가이드입니다.

## 사전 요구사항

### 필수 도구

- **Go 1.25+**: Config Server 백엔드 개발
- **Node.js 20+**: Config Server UI 개발 (선택)
- **Docker 20.10+**: 컨테이너 빌드 및 실행
- **Docker Compose v2.0+**: 로컬 개발 환경
- **PostgreSQL 16+**: 데이터베이스 (또는 Docker로 실행)

### 선택 도구

- **golangci-lint**: Go 코드 린팅
- **pnpm**: Node.js 패키지 관리 (UI 개발 시)
- **kubectl**: Kubernetes 배포 (선택)
- **terraform**: 인프라 프로비저닝 (선택)

## 환경 설정

### 1. Go 설치 및 확인

```bash
# Go 버전 확인
go version
# go version go1.21.x

# GOPATH 확인
echo $GOPATH
# /Users/yourname/go

# Go 모듈 활성화 확인 (기본값)
go env GO111MODULE
# on
```

### 2. Docker 설치 및 확인

```bash
# Docker 버전 확인
docker --version
# Docker version 24.0.x

# Docker Compose 버전 확인
docker-compose --version
# Docker Compose version v2.x.x

# Docker 실행 확인
docker ps
```

### 3. Node.js 설치 (UI 개발 시)

```bash
# Node.js 버전 확인
node --version
# v20.x.x

# pnpm 설치
npm install -g pnpm

# pnpm 버전 확인
pnpm --version
# 8.x.x
```

### 4. golangci-lint 설치

```bash
# macOS (Homebrew)
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 버전 확인
golangci-lint --version
```

## 프로젝트 설정

### 1. 저장소 클론

```bash
git clone https://github.com/fregataa/aami.git
cd aami
```

### 2. 로컬 개발 환경 시작

전체 모니터링 스택 (Prometheus, Grafana, PostgreSQL, Redis) 시작:

```bash
cd deploy/docker-compose

# 환경 변수 설정
cp .env.example .env
# .env 파일 수정 (DB 비밀번호 등)

# 스택 시작
docker-compose up -d

# 로그 확인
docker-compose logs -f

# 상태 확인
docker-compose ps
```

**접속 URL:**
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- Alertmanager: http://localhost:9093
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### 3. Config Server 개발

#### 프로젝트 초기화

```bash
cd services/config-server

# Go 모듈 초기화
go mod init github.com/fregataa/aami/config-server

# 의존성 설치
go mod tidy
```

#### 데이터베이스 마이그레이션

```bash
# 마이그레이션 실행
go run cmd/migrate/main.go up

# 마이그레이션 롤백
go run cmd/migrate/main.go down
```

#### Config Server 실행

```bash
# 환경 변수 설정
export DATABASE_URL="postgres://admin:changeme@localhost:5432/config_server?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export PORT="8080"

# 서버 실행
go run cmd/server/main.go

# 또는 빌드 후 실행
go build -o bin/config-server cmd/server/main.go
./bin/config-server
```

#### API 테스트

```bash
# Health check
curl http://localhost:8080/api/v1/health

# 타겟 목록 조회
curl http://localhost:8080/api/v1/targets

# Prometheus HTTP SD
curl http://localhost:8080/api/v1/sd/prometheus
```

### 4. Config Server UI 개발 (선택)

```bash
cd services/config-server-ui

# pnpm 설치 (전역)
npm install -g pnpm

# 의존성 설치
pnpm install

# 개발 서버 실행
pnpm dev
# http://localhost:3000

# 빌드
pnpm build

# 정적 빌드 확인
ls -la out/
```

## 코드 품질

### Linting

```bash
cd services/config-server

# golangci-lint 실행
golangci-lint run

# 자동 수정 가능한 항목 수정
golangci-lint run --fix

# 특정 디렉토리만 린트
golangci-lint run ./internal/api/...
```

### 포맷팅

```bash
# Go 포맷팅
go fmt ./...

# goimports (import 자동 정리)
goimports -w .

# 모든 파일 포맷팅 확인
gofmt -l .
```

### 테스트

```bash
cd services/config-server

# 모든 테스트 실행
go test ./...

# 상세 출력
go test -v ./...

# 커버리지 측정
go test -cover ./...

# 커버리지 리포트 생성
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race condition 검사
go test -race ./...

# 특정 패키지만 테스트
go test ./internal/api/...

# 특정 테스트만 실행
go test -run TestCreateTarget ./internal/api/...
```

## 디버깅

### VS Code 디버깅

`.vscode/launch.json` 파일 생성:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Config Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/services/config-server/cmd/server",
      "env": {
        "DATABASE_URL": "postgres://admin:changeme@localhost:5432/config_server?sslmode=disable",
        "REDIS_URL": "redis://localhost:6379",
        "PORT": "8080"
      },
      "args": []
    }
  ]
}
```

### Delve (dlv) 사용

```bash
# Delve 설치
go install github.com/go-delve/delve/cmd/dlv@latest

# 디버깅 모드로 실행
cd services/config-server
dlv debug cmd/server/main.go

# 브레이크포인트 설정
(dlv) break main.main
(dlv) continue
```

## 브랜치 전략

```
main
  ├── develop
  │   ├── feature/bootstrap-script
  │   ├── feature/ssh-agent
  │   └── feature/fleet-management
  ├── bugfix/fix-login-validation
  └── hotfix/critical-security-fix
```

### 브랜치 명명 규칙

- `feature/*`: 새로운 기능
- `bugfix/*`: 버그 수정
- `hotfix/*`: 긴급 수정
- `refactor/*`: 리팩토링
- `docs/*`: 문서 업데이트
- `test/*`: 테스트 추가

## 커밋 메시지 규칙

```bash
# 형식
<type>: <subject>

<body>

# 예시
feat: Add bootstrap script auto registration

- Implement bootstrap token management API
- Add hardware auto-detection logic
- Create bootstrap.sh script

# 타입
- feat: 새로운 기능
- fix: 버그 수정
- docs: 문서 변경
- style: 코드 포맷팅
- refactor: 리팩토링
- test: 테스트 추가
- chore: 빌드/도구 변경
```

## 빌드 및 배포

### 로컬 빌드

```bash
cd services/config-server

# 바이너리 빌드
go build -o bin/config-server cmd/server/main.go

# 정적 바이너리 빌드 (CGO 비활성화)
CGO_ENABLED=0 go build -o bin/config-server cmd/server/main.go

# 릴리스 빌드 (최적화)
go build -ldflags="-s -w" -o bin/config-server cmd/server/main.go

# 크로스 컴파일 (Linux)
GOOS=linux GOARCH=amd64 go build -o bin/config-server-linux cmd/server/main.go
```

### Docker 빌드

```bash
cd services/config-server

# Docker 이미지 빌드
docker build -t aami/config-server:latest .

# 특정 플랫폼 빌드
docker buildx build --platform linux/amd64,linux/arm64 -t aami/config-server:latest .

# 이미지 실행
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://admin:changeme@host.docker.internal:5432/config_server?sslmode=disable" \
  -e REDIS_URL="redis://host.docker.internal:6379" \
  aami/config-server:latest
```

## 문제 해결

### Go 모듈 캐시 초기화

```bash
go clean -modcache
go mod download
```

### Docker 컨테이너 재시작

```bash
cd deploy/docker-compose

# 모든 컨테이너 중지 및 삭제
docker-compose down

# 볼륨까지 삭제 (데이터 손실 주의!)
docker-compose down -v

# 재시작
docker-compose up -d
```

### PostgreSQL 연결 오류

```bash
# PostgreSQL 컨테이너 로그 확인
docker-compose logs postgres

# PostgreSQL 접속 테스트
psql -h localhost -U admin -d config_server

# 마이그레이션 재실행
cd services/config-server
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up
```

## 추가 자료

- [PLAN.md](../PLAN.md) - 전체 아키텍처 및 요구사항
- [sprint-plan.md](../sprint-plan.md) - 상세 스프린트 계획
- [Go 공식 문서](https://go.dev/doc/)
- [Prometheus 문서](https://prometheus.io/docs/)
- [Docker Compose 문서](https://docs.docker.com/compose/)

## 지원

문제가 발생하면 다음을 확인하세요:

1. [GitHub Issues](https://github.com/fregataa/aami/issues)
2. [Troubleshooting Guide](./TROUBLESHOOTING.md)
3. Slack 채널: #aami-dev

---

**Happy Coding! 🚀**
