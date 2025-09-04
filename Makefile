run:
	@go run ./cmd

env:
	@test -f .env || cp .env.example .env