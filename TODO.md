- [x] make some request_ts's null in test migrations because it is nullable
- [x] add `Create` to permitRepo
- [ ] add `Get` test to carRepo
    * getting a car that doesn't exit doesn't work
    * getting a car that does exit works
- [ ] add `Create` test to carRepo
    * creating a car with a missing field doesn't work
    * creating a car that already exists doesn't work
- [ ] add `CreateIfNotExists` test to carRepo
    * creating a car with a missing field doesn't work
    * creating a car that already exists just returns that car with no error
    * creating a car that doesn't exist works
- [ ] add `Create` tests to permitRepo:
    * creating a permit with a missing field doesn't work
    * creating a permit with a non-existent car works
    * creating a permit with an existent car works
    * creating a permit that already exists doesn't work
- [ ] make insert repo functions actually query the inserted values from the database instead of just returning their arguments. also test that the values are the same
## Low priority
- [ ] check if it makes sense to use `%w` for errors in `storage/*_repo` files
- [ ] update getoneadmin with sqlx semantics (use get instead of query.scan)
- [ ] probably fix the way that car and permit repo are tied together.
- [ ] add CONVENTIONS doc and mention in it that the storage models use <model-name>Id for id fields
- [ ] change all `id` fields in database to be actually `<model-name>_id`
- [ ] and then change car to not be embedded in permit for consistency with models schema
- [ ] add warning when a non-null empty string is read from db (aka when NullString.Valid is true but NullString.string == '')
