# services

PHONY: build_recipes
build_recipes: 
	@go build -o ./bin/recipes ./cmd/recipes 

PHONY: recipes
recipes: build_recipes
	@./bin/recipes -env-file=.env.local

# workers

# migrators

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