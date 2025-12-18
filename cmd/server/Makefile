.PHONY: help install install-clean build-admin build-portal build-all-ui run-local run-dev run-dev-hot run-staging-hot test test-verbose clean reset-db logs status stop-proxy deploy

# Default target - show help
help:
	@echo "Cronos Platform - Available Make Commands"
	@echo "=========================================="
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make run-local      - Quick start (SQLite, auto-builds UIs)"
	@echo "  make run-dev-hot    - Development mode with hot reload"
	@echo ""
	@echo "🛠  Build Commands:"
	@echo "  make install        - Install all dependencies"
	@echo "  make install-clean  - Clean & reinstall (fixes corrupted deps)"
	@echo "  make build-admin    - Build Vue admin panel"
	@echo "  make build-portal   - Build Vue portal panel"
	@echo "  make build-tailwind - Build Tailwind CSS for public templates"
	@echo "  make build-all-ui   - Build admin, portal, and Tailwind CSS"
	@echo ""
	@echo "🏃 Run Commands:"
	@echo "  make run-local      - Run locally with SQLite (fresh DB each time)"
	@echo "  make run-dev        - Run with pre-built UIs"
	@echo "  make run-dev-hot    - Run with hot reload (Vue dev servers + Air)"
	@echo "  make run-staging-hot - Run in development mode (Cloud SQL)"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  make reset-db       - Delete local SQLite database"
	@echo ""
	@echo "📊 Monitoring:"
	@echo "  make logs           - View logs (requires running server)"
	@echo "  make status         - Check if server is running"
	@echo "  make stop-proxy     - Stop Cloud SQL Proxy if stuck"
	@echo ""
	@echo "🧹 Cleanup:"
	@echo "  make clean          - Remove build artifacts and temp files"
	@echo ""
	@echo "⚠️  If you get module/build errors, try:"
	@echo "  make install-clean  - Fixes corrupted node_modules"
	@echo ""
	@echo "🚢 Deployment:"
	@echo "  make deploy         - Deploy to Google App Engine"
	@echo ""
	@echo "📝 Test Credentials (local mode):"
	@echo "  Staff:  dev@example.com / devpassword"
	@echo "  Client: client@example.com / password"

# Install all dependencies
install:
	@echo "📦 Installing Go dependencies..."
	@go mod download
	@echo "📦 Installing Admin UI dependencies..."
	@cd admin && npm install
	@echo "📦 Installing Portal UI dependencies..."
	@cd portal && npm install
	@echo "✅ All dependencies installed!"

# Clean and reinstall all dependencies (fixes corrupted node_modules)
install-clean:
	@echo "🧹 Cleaning node_modules and package-lock files..."
	@rm -rf admin/node_modules admin/package-lock.json
	@rm -rf portal/node_modules portal/package-lock.json
	@echo "📦 Reinstalling Admin UI dependencies..."
	@cd admin && npm install
	@echo "📦 Reinstalling Portal UI dependencies..."
	@cd portal && npm install
	@echo "✅ Clean install complete!"

# Build the Vue Admin app
build-admin:
	@echo "🔨 Building Admin UI (Output: static/admin)..."
	@cd admin && npm install && npm run build
	@echo "✅ Admin UI build complete."

# Build the Vue Portal app
build-portal:
	@echo "🔨 Building Portal UI (Output: static/portal)..."
	@cd portal && npm install && npm run build
	@echo "✅ Portal UI build complete."

# Build Tailwind CSS for public templates
build-tailwind:
	@echo "🎨 Building Tailwind CSS (Output: assets/css/outputs.css)..."
	@npm install
	@npm run tailwindcss:build
	@echo "✅ Tailwind CSS build complete."

# Build both Admin and Portal UIs + Tailwind CSS
build-all-ui: build-admin build-portal build-tailwind

# Quick local run - builds everything and starts server
run-local: build-all-ui
	@echo "🚀 Starting Cronos in LOCAL mode..."
	@echo ""
	@echo "Environment: SQLite database (fresh on each run)"
	@echo "Server: http://localhost:8080"
	@echo ""
	@echo "Test Credentials:"
	@echo "  Staff:  dev@example.com / devpassword"
	@echo "  Client: client@example.com / password"
	@echo ""
	@ENVIRONMENT=local go run .

# Run the Go server in development mode with pre-built UIs
run-dev: build-all-ui
	@echo "🚀 Starting Cronos in DEV mode (pre-built UIs)..."
	@ENVIRONMENT=local go run .

# Run the Go server with hot reloading enabled
run-dev-hot:
	@echo "🔥 Starting Cronos in HOT RELOAD mode..."
	@echo ""
	@echo "Installing air for hot reloading if not already installed..."
	@command -v air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@echo "Ensuring UI dependencies are installed for admin..."
	@(cd admin && npm install)
	@echo "Ensuring UI dependencies are installed for portal..."
	@(cd portal && npm install)
	@echo ""
	@echo "Starting services:"
	@echo "  - Vue Admin Dev Server (port 5173)"
	@echo "  - Vue Portal Dev Server (port 5174)"
	@echo "  - Go Server with Air (port 8080)"
	@echo ""
	@echo "Server: http://localhost:8080"
	@echo ""
	@(cd admin && npm run dev --watch) & \
	(cd portal && npm run dev --watch) & \
	ENVIRONMENT=local air

# Run with staging/development configuration (Cloud SQL)
run-staging-hot:
	@echo "🏗️  Starting Cronos in DEVELOPMENT mode (Cloud SQL)..."
	@echo ""
	@echo "Checking for required environment variables..."
	@test -n "$$CLOUD_SQL_CONNECTION_NAME" || (echo "❌ CLOUD_SQL_CONNECTION_NAME not set. Source .envrc first." && exit 1)
	@test -n "$$CLOUD_SQL_USERNAME" || (echo "❌ CLOUD_SQL_USERNAME not set. Source .envrc first." && exit 1)
	@test -n "$$CLOUD_SQL_PASSWORD" || (echo "❌ CLOUD_SQL_PASSWORD not set. Source .envrc first." && exit 1)
	@test -n "$$CLOUD_SQL_DATABASE_NAME" || (echo "❌ CLOUD_SQL_DATABASE_NAME not set. Source .envrc first." && exit 1)
	@echo "✅ Environment variables configured"
	@echo ""
	@echo "Checking for Cloud SQL Proxy..."
	@test -f ~/cloud-sql-proxy || (echo "❌ Cloud SQL Proxy not found at ~/cloud-sql-proxy" && exit 1)
	@echo "✅ Cloud SQL Proxy found"
	@echo ""
	@echo "Starting Cloud SQL Proxy on port $${DB_PORT:-5434}..."
	@~/cloud-sql-proxy --port $${DB_PORT:-5434} $$CLOUD_SQL_CONNECTION_NAME > /tmp/cloud-sql-proxy.log 2>&1 & echo $$! > /tmp/cloud-sql-proxy.pid
	@echo "Waiting for proxy to be ready..."
	@sleep 5
	@tail -3 /tmp/cloud-sql-proxy.log
	@echo "✅ Cloud SQL Proxy started (PID: $$(cat /tmp/cloud-sql-proxy.pid))"
	@echo ""
	@echo "Starting services:"
	@echo "  - Cloud SQL Proxy (port $${DB_PORT:-5434})"
	@echo "  - Vue Admin Dev Server (port 5173)"
	@echo "  - Vue Portal Dev Server (port 5174)"
	@echo "  - Go Server with Air (port 8080)"
	@echo ""
	@echo "Server: http://localhost:8080"
	@echo "Proxy logs: tail -f /tmp/cloud-sql-proxy.log"
	@echo ""
	@echo "⚠️  Press Ctrl+C to stop all services"
	@echo ""
	@command -v air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@trap 'echo ""; echo "🛑 Stopping services..."; kill $$(cat /tmp/cloud-sql-proxy.pid 2>/dev/null) 2>/dev/null; rm -f /tmp/cloud-sql-proxy.pid; exit' INT TERM; \
	(cd admin && npm install && npm run dev) & \
	(cd portal && npm install && npm run dev) & \
	ENVIRONMENT=development air; \
	kill $$(cat /tmp/cloud-sql-proxy.pid 2>/dev/null) 2>/dev/null; \
	rm -f /tmp/cloud-sql-proxy.pid

# Run all tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./... -coverprofile=coverage.txt -covermode=atomic
	@echo "✅ Tests complete!"

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests with verbose output..."
	@go test -v ./... -coverprofile=coverage.txt -covermode=atomic
	@echo ""
	@echo "📊 Coverage report saved to coverage.txt"

# Delete local SQLite database
reset-db:
	@echo "🗑️  Deleting local SQLite database..."
	@rm -f cronos.db
	@echo "✅ Database deleted. It will be recreated on next run."

# View server logs (requires server to be running)
logs:
	@echo "📋 Server logs (press Ctrl+C to exit)..."
	@tail -f /var/log/golang/golang-server.log 2>/dev/null || echo "⚠️  Log file not found. Is the server running?"

# Check if server is running
status:
	@echo "📊 Checking server status..."
	@curl -s http://localhost:8080 > /dev/null 2>&1 && echo "✅ Server is running on http://localhost:8080" || echo "❌ Server is not running"

# Stop Cloud SQL Proxy if running
stop-proxy:
	@echo "🛑 Stopping Cloud SQL Proxy..."
	@if [ -f /tmp/cloud-sql-proxy.pid ]; then \
		kill $$(cat /tmp/cloud-sql-proxy.pid) 2>/dev/null && echo "✅ Proxy stopped" || echo "⚠️  Proxy process not found"; \
		rm -f /tmp/cloud-sql-proxy.pid; \
	else \
		pkill -f cloud-sql-proxy && echo "✅ Proxy stopped" || echo "ℹ️  No proxy running"; \
	fi

# Clean build artifacts and temporary files
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf admin/dist
	@rm -rf admin/node_modules
	@rm -rf portal/dist
	@rm -rf portal/node_modules
	@rm -rf static/admin
	@rm -rf static/portal
	@rm -f cronos.db
	@rm -rf tmp
	@rm -f coverage.txt
	@rm -f /tmp/cloud-sql-proxy.pid
	@rm -f /tmp/cloud-sql-proxy.log
	@echo "✅ Clean complete!"

# Deploy to Google App Engine
deploy: build-all-ui
	@echo "🚢 Building UIs and deploying Cronos to Google App Engine..."
	@gcloud app deploy app.yaml --project=snowpack-368423

