include .env

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

# Seed data (local dev only — plain SQL files, not golang-migrate-tracked,
# never run against prod)
.PHONY: seed_up
seed_up:
	for f in seeds/*.up.sql; do psql $(DATABASE_URL) -v ON_ERROR_STOP=1 -f $$f || exit 1; done

.PHONY: seed_down
seed_down:
	for f in $$(ls -r seeds/*.down.sql); do psql $(DATABASE_URL) -v ON_ERROR_STOP=1 -f $$f || exit 1; done

# test data
.PHONY: gen_test_csvs
gen_test_csvs:
	python3 scripts/db/gen/test_data.py csv

.PHONY: gen_test_seeds
gen_test_seeds:
	python3 scripts/db/gen/test_data.py seed
