include .env

PROJECTNAME := lasvistas_api
BIN := bin

MAIN = ./cmd/$(PROJECTNAME)/main.go
EXEC = $(BIN)/$(PROJECTNAME)

all: build

.PHONY: build
build: $(MAIN)
	go build -v -o $(EXEC) $< || exit

.PHONY: test_storage
test_storage:
	go test -v ./storage/

.PHONY: run
run: build
	$(EXEC)

.PHONY: run_memory
run_memory: build
	$(EXEC) -memory

.PHONY: clean
clean:
	rm -rf bin/

# Migrations
.PHONY: migrate_up
migrate_up:
	migrate -path migrations -database $(DATABASE_URL) -verbose up

.PHONY: migrate_up_step
migrate_up_step:
	migrate -path migrations -database $(DATABASE_URL) -verbose up 1

.PHONY: migrate_down
migrate_down:
	migrate -path migrations -database $(DATABASE_URL) -verbose down

.PHONY: migrate_down_step
migrate_down_step:
	migrate -path migrations -database $(DATABASE_URL) -verbose down 1

.PHONY: migrate_force_version
migrate_force_version:
	migrate -path migrations -database $(DATABASE_URL) -verbose force $(version)

.PHONY: migrate_create
migrate_create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: migrate_version
migrate_version:
	migrate -path migrations -database $(DATABASE_URL) version

.PHONY: migrate_prod_up_step
migrate_prod_up_step:
	migrate -path .prodmigrations -database $(DATABASE_URL) -verbose up 1

.PHONY: migrate_prod_down_step
migrate_prod_down_step:
	migrate -path .prodmigrations -database $(DATABASE_URL) -verbose down 1

# test data
.PHONY: gen_test_csvs
gen_test_csvs:
	python3 scripts/gen/test_data.py csv

.PHONY: gen_test_migrations
gen_test_migrations:
	python3 scripts/gen/test_data.py migration
