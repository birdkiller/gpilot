.PHONY: dev-deps backend frontend build test lint clean

# Start development dependencies (PostgreSQL + Redis)
dev-deps:
	docker-compose up -d

dev-deps-down:
	docker-compose down

# Backend
backend:
	cd backend && go run cmd/gpilot/main.go

backend-build:
	cd backend && go build -o bin/gpilot cmd/gpilot/main.go

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && go vet ./...

backend-tidy:
	cd backend && go mod tidy

# Frontend
frontend-install:
	cd frontend && npm install

frontend:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# All
build: backend-build frontend-build

clean:
	rm -rf backend/bin
	rm -rf frontend/dist
