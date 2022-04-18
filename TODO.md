- [x] make some request_ts's null in test migrations because it is nullable
- [x] add `Create` to permitRepo
- [x] add `Get` test to carRepo
    ✓ getting a car that doesn't exist doesn't work
    ✓ getting a car that does exist with NULL fields works
    ✓ getting a car that does exist with no NULL fields works
- [x] add `Create` test to carRepo
    ✓ creating a car with any missing field doesn't work
    ✓ creating a car that already exists doesn't work
- [x] add `CreateIfNotExists` test to carRepo
    ✓ creating a car with a missing field doesn't work
    ✓ creating a car that already exists just returns that car with no error
## Mid priority
- [ ] add `Create` tests to permitRepo:
    * creating a permit with a missing field doesn't work
    * creating a permit with a non-existent car works
    * creating a permit with an existent car works
    * creating a permit that already exists doesn't work
## Low priority
- [x] check if it makes sense to use `%w` for errors in `storage/*_repo` files
- [x] probably fix the way that car and permit repo are tied together.
- [ ] change error format to be filename.func so that only errors are separated by :
- [ ] moved typesafe package to models repo
- [ ] add a list of colors to use as a dropdown
- [ ] add common sentinel errors to api package like errQuery errDecoding
- [ ] make errMissingFields a typesafe.error instead of having duplicate storage.ErrMissingFields and api.ErrMissingFields
- [ ] change the string phrasing in storage.ErrMissingFields
- [ ] update getoneadmin with sqlx semantics (use get instead of query.scan)
- [ ] add CONVENTIONS doc and mention in it that the storage models use <model-name>Id for id fields
- [ ] change all `id` fields in database to be actually `<model-name>_id`
- [ ] and then change car to not be embedded in permit for consistency with models schema
- [ ] add warning when a non-null empty string is read from db (aka when NullString.Valid is true but NullString.string == '')
- [ ] add a test to check that any combination of missing fields doesn't work when creating a car
- [ ] add test to check that in car.CreateIfNotExists, creating a car that doesn't exist works
- [ ] make insert repo functions actually query the inserted values from the database instead of just returning their arguments. also test that the values are the same
- [ ] probably remove the return from `StartServer` function
## Blocked
- [ ] decide how much of car repo should be private
## Keep in mind
- [ ] whether i should make empty-field checking a decorator in repo functions
- [ ] whether i should put all routing funcs in one file. or maybe put the admin/ routing funcs in api/admin
