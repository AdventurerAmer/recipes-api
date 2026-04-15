build:
	@go build -o ./bin/recipes ./cmd/recipes 

run: build
	@./bin/recipes

gen_docs:
	@swagger generate spec –o ./swagger.json