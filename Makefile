.PHONY: build build-mysql dev clean frontend backend backend-mysql

# Build version, injected into the binary. Override with `make VERSION=v1.2.3`.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# Build everything (frontend + backend) — SQLite only (default)
build: frontend backend

# Build everything with MySQL support
build-mysql: frontend backend-mysql

# Build frontend and copy to embed directory
frontend:
	cd web && npm run build
	rm -rf internal/dashboard/dist
	cp -r web/dist internal/dashboard/dist

# Build Go binary — SQLite only (requires frontend to be built first)
backend:
	go build -ldflags="$(LDFLAGS)" -o gatecha ./cmd/gatecha

# Build Go binary with MySQL support (requires frontend to be built first)
backend-mysql:
	go build -tags mysql -ldflags="$(LDFLAGS)" -o gatecha ./cmd/gatecha

# Development: run backend with SQLite (frontend via vite dev proxy)
dev:
	GATECHA_DB_DSN=./data/gatecha.db go run ./cmd/gatecha

# Clean build artifacts
clean:
	rm -f gatecha
	rm -rf internal/dashboard/dist
	rm -rf web/dist
