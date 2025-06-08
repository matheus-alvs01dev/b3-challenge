server:
	@go run cmd/server/main.go

db-populate:
	@go run cmd/dbpopulate/main.go

migration-create:
	@goose -dir internal/adapter/db/migrations/ create $(name) sql

test:
	@go test -v ./... --cover