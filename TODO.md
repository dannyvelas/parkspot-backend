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
- [x] make the `filter` argument to permitRepo.Search an enumerated string to disallow invalid values. or at least perform checking for invalid values
- [x] either A) remove `window` as an option for getting expired permits from repo or B) allow the api to pass in a `window` value when searching for expired permits
- [x] (DEPLOY) add DMARC records to mail server domain
- [x] (DEPLOY) make sure that traffic to parkspotapp.com or any of its subdomains doesn't redirect to the api.lasvistas.parkspotapp or lasvistas.parkspotapp on either port 80 (HTTP) or 443 (HTTPS)
- [x] (DEPLOY) remove NGINX welcome pages
- [x] (DEPLOY) set up renewal for app certificates
- [x] (DEPLOY) set up firewall on server again
- [x] (DEPLOY) change receiving email of password resets from your personal email to the email of the user
- [x] (DEPLOY) change parking days yearly limit to 20
- [x] (DEPLOY) change backend api url from api.lasvistas.parkspotapp.com to api.parkspotapp.com
- [x] (DEPLOY) add dev deploys to dev.api.parkspotapp.com and dev.parkspotapp.com
- [ ] (DEPLOY) (not important) use non-root user in vultr server
- [ ] (DEPLOY) (not important) remove api hello world at "/"
- [ ] (DEPLOY) (not important) remove /api/ prefix from routes
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
- [x] created a new `api.ErrMalformed` error meant for json unmarsheling errors
- [x] change login_router instances to auth_router
- [x] make Role type lowercase.
- [x] rename permit_router funcs so that they explain that they deal w permits for consistency
- [x] add a check for the license_plate of a new car being longer than 10 (avoid database truncation)
- [x] add a check for unique resident emails
- [x] when an permit request has a license plate of an existing car, if the car object in the payload has non null make and model fields, and the existing car has null make and model fields in the database, have the make/model fields in the car object in the payload overwrite the null make/model fields in the database
- [x] remove inline executeTest funcs in permit_router test, they subtly ignore deletepermit errors and are unnecessary in the subtract funcs
- [x] add resident edit/delete functionality
- [x] make listWithMetadata not be generic. unnecessary. just use any. SOL: can't. i would have to convert each slice of structs to a slice of any to initialize listWithMetadata
- [x] maybe use embedding for the LicensePlate, Color, Make, Model fields that Car and Permit share. WONT DO. ugly and not necessary since Permit won't need to re-use any functions that will be defined for Car (like checking for empty/invalid lp/color/make/model)
- [x] remove `<file_name>_<func>` error message convention. not sustainable
- [x] probably make resident creation/deletion endpoint paths consistent (these say `account`, others say `resident`)
- [x] probably remove redundant checks for errnorows in routers that delete residents (you first check whether resident exists by using residentRepo.GetOne, and then by using residentRepo.Delete)
- [✗] add emptyID checking to getActiveDuring\* permit repo funcs as well as resident repo func: `AddToAmtParkingDaysUsed` (wont do bc this should happen at service level not repo level)
- [ ] explore mocking repos so that we don't need to worry about creating/cleaning up rows in db in and in-between tests
- [ ] explore moving some tests from `api/` to `app/`, if test does not focus on any HTTP-related logic
- [ ] when deleting permits, make sure a resident is never set less than 0 days
- [ ] make sure that residents can't make an API request to see someone elses permit
- [ ] add check to make sure permit request start date is not in past
- [ ] make sure residents can't create visitors with a start date in the past
- [ ] add check that contractors can't stay until forever and can stay only for (x) days
- [ ] increment token version on password reset
- [ ] use a `stmtBuilder` variable that will be global in repo files to build sql stmts (already exists in permit_repo)
- [ ] fix respond.go. it doesn't actually set response at 500 when JSON encoding fails
## Low priority
- [x] change error format to be filename.func so that only errors are separated by :
- [x] prepare limit and offset with squirrel, or at least make sure that its okay to not prepare them
- [x] change the string phrasing in storage.ErrMissingFields
- [x] add test to check that in car.CreateIfNotExists, creating a car that doesn't exist works
- [x] probably remove the return from `StartServer` function
- [x] start replacing time.Parse(str) with non-errorable time.Date(...) for brevity in permit\_repo\_test
- [x] add dateFormat to golang config
- [x] make routing its own thing in `api/`
- [x] update getoneadmin with sqlx semantics (use get instead of query.scan)
- [x] rename `limit` query parameter to `limit`
- [x] make CORS / acceptCredentials=true options only for dev and not prod environment if they're not necessary in prod. (cors and acceptCredentials=true is necessary in prod. CORS allows a front-end URL to send a request to the API URL, when they're different domains. acceptCredentials is necessary for the server to be able to read the cookie that comes with the request. [Ref here](https://web.dev/cross-origin-resource-sharing/). But, you can make the CORSALLOWEDORIGINS env variable a specific URL in prod, which makes it safe and appropriate)
- [x] change `username` instances to `id` for consistency
- [x] think of a way to define the resident regex once
- [✓] 500 errors come up when an account exists and the email is malformed. it would probably be better if the response was just: "if this account exists, password reset instructions have been sent to the email sent associated with this account". otherwise, a hacker could technically determine whether accounts exist 
- [✓] do proper status checking in permit_router_test (not just suite.Equal(http.StatusOK, statusCode)) but an actual if check that returns the error response error message
- [✓] make a `/visitors` endpoint return all resident visitors if an admin made the query and only a resident's visitors, if that resident made the query. same goes for `permits/*`
- [✓] change the way that the code connects to postgres from being a bunch of variables to just being a DATABASE\_URL. (start using .env variable and add it to .env.example)
- [✓] maybe use an interface like { GetOne(), SetPasswordFor() } in the auth_router. this might be better than making if statements where one branch does admin.GetOne(...) and the other does resident.GetOne(...) with redundant error checking logic
- [✓] remove constants from config files, just put them inline
- [✓] remove NewPermitArgs NewCarArgs from models. i'd rather send the args individually from the router to the repo, than have a bunch of functions like permitReq.toNewPermitArgs(args...) or newPermitArgs.ToPermit(args..)
- [✓] move all the business logic that routing funcs currently do into a new services package
- [✗] remove `highestVersion` variable from migrator.go (wont do bc dne anymore)
- [✗] make sure only residents can change their password (now that the reset password endpoint will be merged with the editResident endpoint) (wont do bc this isn't the case)
- [ ] probably remove getters from config files, too verbose, not much benefit (it makes sense in theory but not in practice. when are you really going to accidentally override a config value? the answer is probably never)
- [ ] add expiration JWT time to constants
- [ ] make python script also generate down migration file 
- [ ] make python script add line to `migrations/000001_schemas.down.sql` to drop table
- [ ] add logging when a non-null empty string is read from db (aka when NullString.Valid is true but NullString.string == '')
- [ ] explore using validator
## Testing
- [✓] add test that resident can have two active permits at one time, but no more
- [ ] add test to make sure residents can't create exception permits
- [ ] add test to make sure that residents can't make an API request to create a permit for another person
- [ ] add test to make sure that a permit from yesterday to today is counted as active today
- [ ] add test to make sure that a permit from today to tomorrow is counted as active today
- [ ] add `getAllVisitors` testing to visitor\_router
## Maybe going to do
- [✗] whether i should make empty-field checking a decorator in repo functions
- [✗] add `Validated<model-name>` type to prevent redundant calls to `<model-name>.Validate`. hard because everything coming out of the db won't be able to be of this type. (now, models types are validated by default)
- [x] change the argument that goes into permitRepo.Create func. rn it is a CreatePermit which has a CreateCar inside of it. but the CreateCar doesn't get used. so change it to a form of CreatePermit that doesn't have a CreateCar.
- [x] remove `admin/` and `resident/` prefix for adminonly and residentonly routes, respectively, since there are many routes that are shared between both
- [x] whether i should put all routing funcs in one file. or maybe put the admin/ routing funcs in api/admin
- [✗] remove `json` tags from models, since that is an api concern? (won't do, json tags are needed when returning Permits)
- [✗] make exceptionReason, make, and model, nullable pointer strings (won't do, there's no necessary distinction between an empty value and a null value for these fields)
- [✗] implement double-submit tokens
- [✗] whether to change carID to UUID type
- [✓] difference between using byte[8] for residentID for just string
## Not gonna do
- [✗] make models.Permit `make` and `model` fields nullable
    * it's probably fine: their existence as an empty string communicates their non-existence. there is no way that an empty string will ever be communicated as a valid existing value.
    * we could add the "omitEmpty" flag to the json tag if we ever wanted to distinguish. but that's not necessary now. this would only be nice for consistency reasons if we had future fields that could are nullable and did have valid empty values like `amtCars` (does 0 mean that there are no cars or that this field was never set?). but, that's not the case now.
- [✗] change all `id` fields in database to be actually `<model-name>_id`. not necessary. the way it is now is consistent (all database models have `id`, storage models have `<modelName>ID` and regular models have `ID`.) 
- [✗] Validate repo func decorator that could be defined in `models`
- [✗] make models.<model-name> struct fields private so that `models.<model-name>{}` initializations outside of `models` package can be prevented
- [✗] change car to not be embedded in storage.permit for consistency with models schema
- [✗] add a test to check that any combination of missing fields doesn't work when creating a car
- [✗] maybe make carRepo, permitRepo, adminRepo, ... fields on `storage.Database` and make all the receiving funcs of those repos receivers of storage.Database. that way repo funcs can easily call repo funcs of a different model
- [✗] make insert repo functions actually query the inserted values from the database instead of just returning their arguments. also test that the values are the same
- [✗] change WHERE db stmts in car_repo to be like `WHERE license_plate = ..` and not `WHERE car.license_plate = ...` same thing for `car.id`
- [✗] add a list of colors to use as a dropdown
- [✗] use type aliases for `models` datatype fields like LicensePlate Make, model, AddToAmtParkingDaysUsed, ..StartDate.., etc (this would prevent passing a licensePlate (string) as a `Make` (also string) argument accidentally
- [✗] move `migrations/` dir inside of `storage`
- [✗] make routing handlers receivers off of an injected struct (like in storage) to avoid func name conflicts
- [✗] delete Car.GetOne if it's not going to be used. it is used
## Conventions
- [ ] add CONVENTIONS doc and mention in it that the storage models use <model-name>ID for id fields
- [ ] mention in conventions that the error msg is `file_name.func_name: error: wrapped-error`. func name and wrapped-error are optional wrapped-error will be %v if it's a 3rd party error and %w if its an error defined within this code
- [ ] move comment about // check that they're equal not using suite.Equal because... to CONVENTIONS.md
