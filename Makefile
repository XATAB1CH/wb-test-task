run:
	@go run ./cmd

env:
	@test -f .env || cp .env.example .env

imports:
	@go install golang.org/x/tools/cmd/goimports@latest
	@goimports -w .