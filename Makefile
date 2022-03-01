include .env

pg-connection := postgresql://$(PG_USER):$(PG_PASSWORD)@$(SERVER_HOST):$(PG_PORT)/$(PG_DBNAME)?sslmode=$(PG_SSLMODE)

migrateup:
	migrate -path migrations -database $(pg-connection) -verbose up

migratedown:
	migrate -path migrations -database $(pg-connection) -verbose down

.PHONY: migrateup migratedown
