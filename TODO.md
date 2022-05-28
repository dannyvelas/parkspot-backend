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
- [x] storage/permit_repo.Create: figure out better way to convert an inserted permit `CreatePermit` type to `Permit` type. maybe with helper func, maybe with psql
- [x] create exceptions table and allow requests to create exceptional permits
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
- [x] create models.NewPermit() func
- [✗] change respondJSON to respondData
- [✗] add api.error for malformed jwt
- [✗] add common sentinel errors to api package like errQuery errDecoding and return them in response
- [x] change psql int type to uint64
- [x] add much more validation to permit type
- [x] make amount of parking days a constant
- [x] figure out if to have CreateCar.ToCar(id string) func and CreatePermit.ToPermit(id int64) func
- [x] add `Create` tests to permitRepo:
    * creating a permit with a missing field doesn't work
    * creating a permit with a non-existent car works
    * creating a permit with an existent car works
- [x] rename `CreatePermit` and `CreateCar` structs to `NewPermitArgs` and `NewCarArgs`
- [x] add `started server at URL:PORT` to main message
- [x] change PORT env variable to string
- [ ] created a new `api.ErrMalformed` error meant for json unmarsheling errors
- [ ] make exceptionReason, make, and model, nullable pointer strings
- [ ] change error messages for residents when they're creating a permit
- [ ] add `AddToAmtParkingDaysUsed` and `GetAll` testing to resident repo
- [ ] add emptyID checking to getActiveDuring\* permit repo funcs as well as resident repo  func: `AddToAmtParkingDaysUsed`
- [ ] add check to make sure permit request start date is not in past
## Low priority
- [x] change error format to be filename.func so that only errors are separated by :
- [x] prepare limit and offset with squirrel, or at least make sure that its okay to not prepare them
- [x] change the string phrasing in storage.ErrMissingFields
- [x] add test to check that in car.CreateIfNotExists, creating a car that doesn't exist works
- [x] probably remove the return from `StartServer` function
- [x] start replacing time.Parse(str) with non-errorable time.Date(...) for brevity in permit_repo_test
- [x] add dateFormat to golang config
- [x] make routing its own thing in `api/`
- [x] update getoneadmin with sqlx semantics (use get instead of query.scan)
- [x] rename `limit` query parameter to `limit`
- [ ] remove "No error when" messages from repo_tests. unnecessary
- [ ] add expiration JWT time to constants
- [ ] make CORS / acceptCredentials=true options only for dev and not prod environment if they're not necessary in prod
- [ ] add warning when a non-null empty string is read from db (aka when NullString.Valid is true but NullString.string == '')
- [ ] make python script also generate down migrations
- [ ] change WHERE db stmts in car_repo to be like `WHERE license_plate = ..` and not `WHERE car.license_plate = ...` same thing for `car.id`
- [ ] add a list of colors to use as a dropdown
## Maybe going to do
- [✗] whether i should make empty-field checking a decorator in repo functions
- [✗] add `Validated<model-name>` type to prevent redundant calls to `<model-name>.Validate`. hard because everything coming out of the db won't be able to be of this type. (now, models types are validated by default)
- [x] change the argument that goes into permitRepo.Create func. rn it is a CreatePermit which has a CreateCar inside of it. but the CreateCar doesn't get used. so change it to a form of CreatePermit that doesn't have a CreateCar.
- [x] remove `admin/` and `resident/` prefix for adminonly and residentonly routes, respectively, since there are many routes that are shared between both
- [ ] maybe use validator
- [ ] make routing handlers receivers off of an injected struct (like in storage) to avoid func name conflicts
- [ ] implement double-submit tokens
    * implement double-submit tokens without REDIS
    * implement double-submit tokens with REDIS
    * for fun: replace r.context() with struct?
- [ ] move `migrations/` dir inside of `storage`
- [ ] share existingCreateCar variable between both permit_repo_test and car_repo_test
- [ ] figure out if to use type aliases for `models` datatype fields like LicensePlate Make, model, AddToAmtParkingDaysUsed, ..StartDate.., etc (this would prevent passing a licensePlate (string) as a `Make` (also string) argument accidentally
- [ ] remove `json` tags from models, since that is an api concern?
- [ ] delete Car.GetOne if it's not going to be used
- [ ] whether to change carID to UUID type
- [ ] difference between using byte[8] for residentID for just string
- [ ] whether i should put all routing funcs in one file. or maybe put the admin/ routing funcs in api/admin
- [ ] permit_router: put list of permits that are active during the create permit start/end date range when len(activePermitsDuring) != 0 in error message
## Probably not gonna do
- [ ] make models.Permit `make` and `model` fields nullable
    * it's probably fine: their existence as an empty string communicates their non-existence. there is no way that an empty string will ever be communicated as a valid existing value.
    * we could add the "omitEmpty" flag to the json tag if we ever wanted to distinguish. but that's not necessary now. this would only be nice for consistency reasons if we had future fields that could are nullable and did have valid empty values like `amtCars` (does 0 mean that there are no cars or that this field was never set?). but, that's not the case now.
- [ ] change all `id` fields in database to be actually `<model-name>_id`. not necessary. the way it is now is consistent (all database models have `id`, storage models have `<modelName>Id` and regular models have `Id`.) 
- [ ] Validate repo func decorator that could be defined in `models`
- [ ] make models.<model-name> struct fields private so that `models.<model-name>{}` initializations outside of `models` package can be prevented
- [ ] change car to not be embedded in storage.permit for consistency with models schema
- [ ] add a test to check that any combination of missing fields doesn't work when creating a car
- [ ] maybe make carRepo, permitRepo, adminRepo, ... fields on `storage.Database` and make all the receiving funcs of those repos receivers of storage.Database. that way repo funcs can easily call repo funcs of a different model
- [ ] make insert repo functions actually query the inserted values from the database instead of just returning their arguments. also test that the values are the same
## Conventions
- [ ] add CONVENTIONS doc and mention in it that the storage models use <model-name>Id for id fields
- [ ] mention in conventions that the error msg is `file_name.func_name: error: wrapped-error`. func name and wrapped-error are optional wrapped-error will be %v if it's a 3rd party error and %w if its an error defined within this code
- [ ] move comment about // check that they're equal not using suite.Equal because... to CONVENTIONS.md
- [ ] mention that we use Id and not ID
