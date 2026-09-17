# docker

PHONY: up
up: 
	@docker-compose up

PHONY: down
down:
	@docker-compose down

PHONY: downv
downv:
	@docker-compose down -v

PHONY: build
build:
	@go build -o ./bin/recipes ./cmd/recipes 

PHONY: run
run: build
	@./bin/recipes -env-file=.env.local

PHONY: docs
docs:
	@swagger generate spec –o ./swagger.json