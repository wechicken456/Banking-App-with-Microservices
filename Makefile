include .env

all: goose-up sqlc


goose-up:
	cd auth/db/schema && goose postgres "postgres://$(AUTH_DB_USER):$(AUTH_DB_PASSWORD)@$(AUTH_DB_HOST):$(AUTH_DB_PORT)/$(AUTH_DB_NAME)?sslmode=disable" up
	cd account/db/schema && goose postgres "postgres://$(ACCOUNT_DB_USER):$(ACCOUNT_DB_PASSWORD)@$(ACCOUNT_DB_HOST):$(ACCOUNT_DB_PORT)/$(ACCOUNT_DB_NAME)?sslmode=disable" up
	cd transfer/db/schema && goose postgres "postgres://$(TRANSFER_DB_USER):$(TRANSFER_DB_PASSWORD)@$(TRANSFER_DB_HOST):$(TRANSFER_DB_PORT)/$(TRANSFER_DB_NAME)?sslmode=disable" up

goose-down:
	cd auth/db/schema && goose postgres "postgres://$(AUTH_DB_USER):$(AUTH_DB_PASSWORD)@$(AUTH_DB_HOST):$(AUTH_DB_PORT)/$(AUTH_DB_NAME)?sslmode=disable" down
	cd account/db/schema && goose postgres "postgres://$(ACCOUNT_DB_USER):$(ACCOUNT_DB_PASSWORD)@$(ACCOUNT_DB_HOST):$(ACCOUNT_DB_PORT)/$(ACCOUNT_DB_NAME)?sslmode=disable" down
	cd transfer/db/schema && goose postgres "postgres://$(TRANSFER_DB_USER):$(TRANSFER_DB_PASSWORD)@$(TRANSFER_DB_HOST):$(TRANSFER_DB_PORT)/$(TRANSFER_DB_NAME)?sslmode=disable" down

sqlc:
	cd auth && sqlc generate && cd ..
	cd account && sqlc generate && cd ..
	cd transfer && sqlc generate && cd ..

protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative account/proto/account_service.proto

clean: goose-down
	
	

