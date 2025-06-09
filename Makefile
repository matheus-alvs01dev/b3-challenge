server:
	@go run cmd/server/main.go

db-populate:
	@go run cmd/dbpopulate/main.go

migration-create:
	@goose -dir internal/adapter/db/migrations/ create $(name) sql

test:
	@go test -v ./... --cover

lint:
	@docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.1.5 golangci-lint run -v -c dev/golangci.yaml ./...