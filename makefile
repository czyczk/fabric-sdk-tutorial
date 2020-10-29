.PHONY: all dev clean build env-up env-down run

all: clean env-up run

##### BUILD
build:
	@echo "Building..."
	@go build
	@echo "Done building."

##### ENV
env-up:
	@echo "Starting environment..."
	@cd fixtures && docker-compose up --force-recreate -d
	@sleep 2
	@echo "Environment is up."

env-down:
	@echo "Stopping environmnt..."
	@cd fixtures && docker-compose down
	@echo "Environment is down."

##### RUN
run:
	@echo "Starting app..."
	@./fabric-sdk-tutorial

##### CLEAN
clean: env-down
	@echo "Cleaning up..."
	@docker rm -f -v `docker ps -a --no-trunc | grep "lab805" | cut -d ' ' -f 1` 2>dev/null || true
	@docker rmi `docker images --no-trunc | grep "lab805" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Done cleaning up."