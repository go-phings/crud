.DEFAULT_GOAL := help

.PHONY: help test

test: ## Runs tests
	go test

start-db:
	@echo "* Creating docker container with PostgreSQL"
	docker run --name crud-sample-user-db -d -e POSTGRES_PASSWORD=crudpass -e POSTGRES_USER=cruduser -e POSTGRES_DB=crud -p 54321:5432 postgres:13
	@echo "* Sleeping for 10 seconds to give database time to initialize..."
	@sleep 10

run-sample-user: clean start-db ## Runs sample-user app
	@echo "* Building and starting application..."
	@echo "* Please run 'make clean' after terminating the application!"
	cd cmd/sample-user && go build .
	cd cmd/sample-user && ./sample-user
	
clean: ## Removes all created dockers
	@echo "* Removing previously created docker container..."
	docker rm -f crud-sample-user-db

help: ## Displays this help
	@awk 'BEGIN {FS = ":.*##"; printf "$(MAKEFILE_NAME)\n\nUsage:\n  make \033[1;36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[1;36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
