.PHONY: build dev clean frontend backend

# Build everything (frontend + backend)
build: frontend backend

# Build frontend and copy to embed directory
frontend:
	cd web && npm run build
	rm -rf internal/dashboard/dist
	cp -r web/dist internal/dashboard/dist

# Build Go binary (requires frontend to be built first)
backend:
	go build -ldflags="-s -w" -o gatecha ./cmd/gatecha

# Development: run backend with hot reload (frontend via vite dev proxy)
dev:
	GATECHA_DB_PATH=./data/gatecha.db go run ./cmd/gatecha

# Clean build artifacts
clean:
	rm -f gatecha
	rm -rf internal/dashboard/dist
	rm -rf web/dist
