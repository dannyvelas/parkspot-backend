# Park Spot backend

This is the backend HTTP API service for the ParkSpot application. ParkSpot is a parking reservation and administration app that I developed for residential communities in Miami.

This service interfaces with a database which holds a list of residents, permits, visitors, and cars that belong to a community.

## Authorization
* API routes have a middleware which checks that the requesting user has the authorization to hit that endpoint.

Administrators can:
* Create/Read/Update/Delete the residents of their community
* Create/Read/Update/Delete the parking permits of their community
* Read the visitors of their community
* Create/Read/Update/Delete the cars of their community

Security can:
* Read the residents of their community
* Read the parking permits of their community
* Read the visitors of their community
* Read the cars of their community

Residents can:
* Create/Read their own parking permits
* Create/Read/Update/Delete their visitors
* Create/Read/Update/Delete their cars

All users can:
* Create a session
* Close their session
* Reset their password

## Authentication
* This service uses [refresh tokens](https://auth0.com/blog/refresh-tokens-what-are-they-and-when-to-use-them/) to track sessions.
* The Refresh token of a `resident` or `admin` can be revoked by incrementing the user's `token_version` field in the `resident` or `admin` table, respectively.
* By default, access tokens are set to expire after 15 minutes, and refresh tokens are set to expire after 7 days, but this is configurable.
* If a client has an expired refresh token, this client can make a `POST` request to the `/refresh-tokens` endpoint. This request will update the refresh token cookie in the client. The response will have a new access token.

## More session information
* Some user fields are purposely not exposed via HTTP, like `token_version` or `password`.
* `password`s are always stored after hashing and salting with `bcrypt`.

## Setup
1. Install docker
2. Run docker
3. Run PostgreSQL instance: `docker compose up -d`
4. Create database models and seed them with sample data: `make migrate_up`
5. Create an `.env` file: `cp .env.example .env`
6. Run the service: `go run -v main.go`


## Local development
* All residents/admins in the sample data have the password: `notapassword`.
