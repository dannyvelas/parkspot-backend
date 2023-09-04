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

.PHONY: clean
clean:
	rm -rf bin/

# Migrations
.PHONY: migrate_up
migrate_up:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) up

.PHONY: migrate_up_step
migrate_up_step:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) up 1

.PHONY: migrate_down
migrate_down:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) down

.PHONY: migrate_down_step
migrate_down_step:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) down 1

.PHONY: migrate_force_version
migrate_force_version:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) force $(version)

.PHONY: migrate_create
migrate_create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: migrate_version
migrate_version:
	go run cmd/migrate/main.go -path migrations -database $(DATABASE_URL) version

# test data
.PHONY: gen_test_csvs
gen_test_csvs:
	python3 scripts/gen/test_data.py csv

.PHONY: gen_test_migrations
gen_test_migrations:
	python3 scripts/gen/test_data.py migration
