include .env

PROJECTNAME := lasvistas_api
BIN := bin

MAIN = ./cmd/$(PROJECTNAME)/main.go
EXEC = $(BIN)/$(PROJECTNAME)

PGCONNECTION := postgresql://$(PG_USER):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DBNAME)?sslmode=$(PG_SSLMODE)

all: build

build: $(MAIN)
	go build -v -o $(EXEC) $< || exit

run: build
	$(EXEC)

clean:
	rm -rf bin/

# Migrations
migrate_up:
	migrate -path migrations -database $(PGCONNECTION) -verbose up

migrate_up_step:
	migrate -path migrations -database $(PGCONNECTION) -verbose up 1

migrate_down:
	migrate -path migrations -database $(PGCONNECTION) -verbose down

migrate_down_step:
	migrate -path migrations -database $(PGCONNECTION) -verbose down 1


.PHONY: clean migrate_up migrate_up_step migrate_down migrate_down_step
