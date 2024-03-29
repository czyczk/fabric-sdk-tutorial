export COMPOSE_PROJECT_NAME=lab805
export DOCKER_FILES_PARAM=-f docker-compose.yaml -f docker-compose-couch.yaml -f docker-compose-ipfs.yaml

.PHONY: all dev clean build env-up env-down run run-init run-serve

all: clean build env-up run

##### BUILD
build:
	@echo "Building..."
	@go build
	@echo "Done building."

##### ENV
env-up:
	@echo "Starting environment..."
	@cd fixtures && rm -rf ./ipfs && cp -r ./ipfs-template ./ipfs
	@cd fixtures && docker-compose ${DOCKER_FILES_PARAM} up --force-recreate -d
	@sleep 3
	@echo "Environment is up."

env-down:
	@echo "Stopping environment..."
	@cd fixtures && docker-compose ${DOCKER_FILES_PARAM} down
	@echo "Environment is down."

##### RUN
run:
	@echo "Starting app..."
	$(MAKE) run-init && $(MAKE) run-serve

run-init:
	@cd chaincode/src/universal_cc && go mod vendor && cd ../../.. && ./fabric-sdk-tutorial init

run-serve:
	@./fabric-sdk-tutorial serve

run-serve-ado1:
	@./fabric-sdk-tutorial serve -c "server-ado1.yaml"

run-serve-u1o2:
	@./fabric-sdk-tutorial serve -c "server-u1o2.yaml"

##### CLEAN
clean: env-down
	@echo "Cleaning up..."
	@rm -rf ./fixtures/ipfs
	@docker rm -f -v `docker ps -a --no-trunc | grep "lab805" | cut -d ' ' -f 1` 2>/dev/null || true
	@docker rmi `docker images --no-trunc | grep "lab805" | cut -d ' ' -f 1` 2>/dev/null || true
	@echo "Done cleaning up."
