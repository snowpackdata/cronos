.PHONY: build-admin build-portal build-all-ui run-dev run-dev-hot run-staging-hot deploy

# Build the Vue Admin app
build-admin:
	@echo "Building Admin UI (Output: static/admin)..."
	cd admin && npm install && npm run build
	@echo "Admin UI build complete."

# Build the Vue Portal app
build-portal:
	@echo "Building Portal UI (Output: static/portal)..."
	cd portal && npm install && npm run build
	@echo "Portal UI build complete."

# Build both Admin and Portal UIs
build-all-ui: build-admin build-portal

# Run the Go server in development mode with pre-built UIs
run-dev: build-all-ui
	ENVIRONMENT=local go run .

# Run the Go server with hot reloading enabled
run-dev-hot:
	@echo "Installing air for hot reloading if not already installed..."
	@command -v air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@echo "Ensuring UI dependencies are installed for admin..."
	(cd admin && npm install)
	@echo "Ensuring UI dependencies are installed for portal..."
	(cd portal && npm install)
	@echo "Starting Vue dev servers (admin, portal) and Go server with hot reloading in parallel..."
	@(cd admin && npm run dev --watch) & \
	(cd portal && npm run dev --watch) & \
	ENVIRONMENT=local air

# Run with staging/development configuration
run-staging-hot:
	@echo "Installing air for hot reloading if not already installed..."
	@command -v air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	@echo "Ensuring UI dependencies are installed for admin..."
	(cd admin && npm install)
	@echo "Ensuring UI dependencies are installed for portal..."
	(cd portal && npm install)
	@echo "Starting Vue dev servers (admin, portal) and Go server with hot reloading in parallel..."
	@(cd admin && npm run dev) & \
	(cd portal && npm run dev) & \
	ENVIRONMENT=development air

# Deploy to Google App Engine
deploy: build-all-ui
	@echo "Building UIs and deploying Cronos to Google App Engine..."
	gcloud app deploy app.yaml --project=snowpack-368423

