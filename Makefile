.PHONY: run up build down re logs ps

# Run the forum locally without Docker
run:
	go run -tags fts5 ./cmd

# Start the forum using Docker
up:
	docker compose up -d

# Rebuild the Docker image and start the forum
build:
	docker compose up --build -d

# Stop Docker containers
down:
	docker compose down

# Rebuild and restart Docker
re: down build

# Follow application logs
logs:
	docker compose logs -f

# Check Docker container status
ps:
	docker compose ps
