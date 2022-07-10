include .env

PROJECTNAME := lasvistas_api
BIN := bin

MAIN = ./cmd/$(PROJECTNAME)/main.go
EXEC = $(BIN)/$(PROJECTNAME)

PGCONNECTION := postgresql://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DBNAME)?sslmode=$(PG_SSLMODE)

all: build

build: $(MAIN)
	go build -v -o $(EXEC) $< || exit

test_storage:
	go test -v ./storage/
.PHONY: test_storage    

run: build
	$(EXEC)
.PHONY: run

clean:
	rm -rf bin/
.PHONY: clean

# Migrations
migrate_up:
	migrate -path migrations -database $(PGCONNECTION) -verbose up
.PHONY: migrate_up

migrate_up_step:
	migrate -path migrations -database $(PGCONNECTION) -verbose up 1
.PHONY: migrate_up_step

migrate_down:
	migrate -path migrations -database $(PGCONNECTION) -verbose down
.PHONY: migrate_down

migrate_down_step:
	migrate -path migrations -database $(PGCONNECTION) -verbose down 1
.PHONY: migrate_down_step

migrate_force_version:
	migrate -path migrations -database $(PGCONNECTION) -verbose force $(version)
.PHONY: migrate_force_version

migrate_create:
	migrate create -ext sql -dir migrations -seq $(name)
.PHONY: migrate_create

migrate_version:
	migrate -path migrations -database $(PGCONNECTION) version
.PHONY: migrate_version

migrate_prod_up_step:
	migrate -path .prodmigrations -database $(PGCONNECTION) -verbose up 1
.PHONY: migrate_prod_up_step

migrate_prod_down_step:
	migrate -path .prodmigrations -database $(PGCONNECTION) -verbose down 1
.PHONY: migrate_prod_down_step

# test data
gen_test_csvs:
	python3 scripts/gen/test_data.py csv
.PHONY: gen_test_csvs

gen_test_migrations:
	python3 scripts/gen/test_data.py migration
.PHONY: gen_test_migrations
