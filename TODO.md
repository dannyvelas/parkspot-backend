## High Priority
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
- [ ] storage/permit_repo.Create: figure out better way to convert an inserted permit `CreatePermit` type to `Permit` type. maybe with helper func, maybe with psql
## Mid priority
- [x] check if it makes sense to use `%w` for errors in `storage/*_repo` files
- [x] probably fix the way that car and permit repo are tied together.
- [x] moved typesafe package to models repo
- [x] make errMissingFields a typesafe.error instead of having duplicate storage.ErrMissingFields and api.ErrMissingFields
- [x] split up permits getall test
- [x] find out how to make dateFormat global
- [x] change storage.erremtpyidarg to generic errinput
- [x] make squirrel errors a new error type
- [✗] switch int64 timestamp types to uint64 (won't do)
- [x] replace "No error" assert messages in tests with "error"
- [x] remove models/errors.go
- [ ] add common sentinel errors to api package like errQuery errDecoding and return them in response
- [ ] create models.NewPermit() func. rename CreatePermit/CreateCar to PermitArgs/CarArgs
- [ ] change psql int type to uint64
- [ ] add much more validation to permit type
- [ ] add `Create` tests to permitRepo:
    * creating a permit with a missing field doesn't work
    * creating a permit with a non-existent car works
    * creating a permit with an existent car works
    * creating a permit that already exists doesn't work
- [ ] make routing its own thing in `api/`
- [ ] make routing handlers receivers off of an injected struct (like in storage) to avoid func name conflicts
## Low priority
- [x] change error format to be filename.func so that only errors are separated by :
- [x] prepare limit and offset with squirrel, or at least make sure that its okay to not prepare them
- [x] change the string phrasing in storage.ErrMissingFields
- [x] add test to check that in car.CreateIfNotExists, creating a car that doesn't exist works
- [ ] add a list of colors to use as a dropdown
- [ ] update getoneadmin with sqlx semantics (use get instead of query.scan)
- [ ] probably remove the return from `StartServer` function
## Maybe going to do
- [ ] whether i should make empty-field checking a decorator in repo functions
- [ ] whether i should put all routing funcs in one file. or maybe put the admin/ routing funcs in api/admin
- [ ] add `Validated<model-name>` type to prevent redundant calls to `<model-name>.Validate`. hard because everything coming out of the db won't be able to be of this type
- [ ] Validate repo func decorator that could be defined in `models`
- [ ] permit_router: put list of permits that are active during the create permit start/end date range when len(activePermitsDuring) != 0 in error message
## Probably not gonna do
- [ ] change all `id` fields in database to be actually `<model-name>_id`
- [ ] change car to not be embedded in storage.permit for consistency with models schema
- [ ] add warning when a non-null empty string is read from db (aka when NullString.Valid is true but NullString.string == '')
- [ ] add a test to check that any combination of missing fields doesn't work when creating a car
- [ ] maybe make carRepo, permitRepo, adminRepo, ... fields on `storage.Database` and make all the receiving funcs of those repos receivers of storage.Database. that way repo funcs can easily call repo funcs of a different model
- [ ] make insert repo functions actually query the inserted values from the database instead of just returning their arguments. also test that the values are the same
## Conventions
- [ ] add CONVENTIONS doc and mention in it that the storage models use <model-name>Id for id fields
- [ ] mention in conventions that the error msg is `file_name.func_name: error: wrapped-error`. func name and wrapped-error are optional wrapped-error will be %v if it's a 3rd party error and %w if its an error defined within this code
- [ ] move comment about // check that they're equal not using suite.Equal because... to CONVENTIONS.md
- [ ] mention that we use Id and not ID
