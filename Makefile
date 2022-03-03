include .env

pg-connection := postgresql://$(PG_USER):$(PG_PASSWORD)@$(SERVER_HOST):$(PG_PORT)/$(PG_DBNAME)?sslmode=$(PG_SSLMODE)

migrate_up:
	migrate -path migrations -database $(pg-connection) -verbose up

migrate_up_step:
	migrate -path migrations -database $(pg-connection) -verbose up 1

migrate_down:
	migrate -path migrations -database $(pg-connection) -verbose down

migrate_down_step:
	migrate -path migrations -database $(pg-connection) -verbose down 1
