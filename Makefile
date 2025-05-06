include .env

all: goose-up sqlc


goose-up:
	cd auth/db/schema && goose postgres "postgres://$(AUTH_DB_USER):$(AUTH_DB_PASSWORD)@$(AUTH_DB_HOST):$(AUTH_DB_PORT)/$(AUTH_DB_NAME)?sslmode=disable" up

goose-down:
	cd auth/db/schema && goose postgres "postgres://$(AUTH_DB_USER):$(AUTH_DB_PASSWORD)@$(AUTH_DB_HOST):$(AUTH_DB_PORT)/$(AUTH_DB_NAME)?sslmode=disable" down

sqlc:
	cd auth && sqlc generate



clean: goose-down
	
	

