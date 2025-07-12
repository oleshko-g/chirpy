# Chirpy

## What is it?

It's an HTTP-server of a fictional Chirpy web app.

## The features

### Serve web pages

* Chirpy's main page
* The number of the file server hits
* "Assets" directory
* The logo of Chirpy

### Manage users

* Register
* Login with email and passed and sign access JWT
* Manage users' access JWTs through refresh tokens
* Update the user's login or password
* Track "Chirpy Red" premium subscription status through the Polka web-hook
* Dev only option: delete all the users and their data

### Chirps

* Post a chirp
* Delete the chirp
* Get chirps
    * Optionally: filter by author
    * Optionally: sort by the creation date

### Maintenance

* Check if the server is up

## How to set up

First of all create the file for the environment variables and add it to the `.gitignore` file.
Run in the terminal in the root of the repo:

```bash
touch .env && printf '.env' >> .gitignore
```

### Database

1. Install PostgreSQL
2. Install Go 1.24+ 
3. Install goose.

    Run in the terminal:
      ```bash
      go install github.com/pressly/goose/v3/cmd/goose@latest
      ``` 

4. Set goose env variables in the `.env` file:

    1. Run in the terminal to set goose driver and the migrations directory:
    ```bash
    printf '\n# goose env variables\nGOOSE_DRIVER="postgres" \nGOOSE_MIGRATION_DIR="./sql/schema"'>> .env
    ```
    2. Run in the terminal and manually replace `USER` with your PostgreSQL credentials to set database connection string:
    ```bash
    GOOSE_DBSTRING="postgres://USER:@localhost:5432/chirpy"`
    ```

5. Run goose migrations. Run in the terminal to create the database tables:
    ```bash
    goose up
    ```

6. Set database connection string.
    Run in the terminal:
    ```bash
    printf '\n # Database\nDB_URL="postgres://USER:@localhost:5432/chirpy?sslmode=disable"' >> .env
    ```

### Authentication 

Generate the JWT secret to sign users' access JWTs.

Run in the terminal: 
```bash
printf '# JWT Secret\nJWT_SECRET="%s"\n' "$(openssl rand -base64 64)" >> .env
```

### Options

If you want to reset the users data – set the platform variable.

Run in the terminal`: 
```bash
printf '\n# Options\nPLATFORM="dev"' >> .env
```

### Polka

For Polka web-hook to work set Polka API key.

Run in the terminal and paste Polka API key from "Learn HTTP servers in Go" course on boot\.dev":

```bash
printf '\n# Polka\nPOLKA_API_KEY="POLKA_API_KEY"' >> .env
```

## How to run?

Run in the terminal in the root of the repo:

```bash
go run .
```

OR

```bash
go build -o chirpy && ./chirpy
```

## HTTP endpoints

### Web pages

#### GET /app/

Serves the main page of Chirpy as an HTML-file.

#### GET /app/assets/

Serves the "assets" directory as an HTML-file.

#### GET /app/assets/logo.png

Serves the logo of Chirpy as an HTML-file.

#### GET /admin/metrics

Shows the number of file server hits as an HTML-file. 

### Users

#### POST /api/users

Registers a user with the email and password.

##### Request

```json
{
  "email": "saul@bettercall.com",
  "password": "123456"
}
```

##### Response

```json
{
  "id": "50746277-23c6-4d85-a890-564c0044c2fb",
  "created_at": "2021-07-07T00:00:00Z",
  "updated_at": "2021-07-07T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

#### PUT /api/users

Updates the user's both email and password.

##### Authentication

Headers: `Authorization: Bearer {the user's JWT}`

##### Request
```json
{
  "email": "saul@bettercall.com",
}
```

##### Response

```json
{
  "id": "50746277-23c6-4d85-a890-564c0044c2fb",
  "created_at": "2021-07-07T00:00:00Z",
  "updated_at": "2021-07-07T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

#### POST /api/login

1. Logs in the user with the email and password. 
2. Signs and responds with the user's access JWT.
3. Generates and responds with the user's refresh token.
   
##### Request

```json
{
  "password": "04234",
  "email": "lane@example.com"
}
```

##### Response
```json
{
  "id": "5a47789c-a617-444a-8a80-b50359247804",
  "created_at": "2021-07-01T00:00:00Z",
  "updated_at": "2021-07-01T00:00:00Z",
  "email": "lane@example.com",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
  "refresh_token": "56aa826d22baab4b5ec2cea41a59ecbba03e542aedbb31d9b80326ac8ffcfa2a"
}
```

#### POST /api/refresh

Refreshes and responds with the users's access JWT by the passed user's refresh token.

##### Authentication

Headers: `Authorization: Bearer {the user's refresh token}`

##### Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
}
```

#### POST /api/revoke

Revokes the passed refresh token.

##### Authentication

Headers: `Authorization: Bearer {the user's refresh token}`

### Chirps

#### POST /api/chirps

Posts a chirp as the authenticated user.

##### Authentication

Headers: `Authorization: Bearer {the user's JWT}`

##### Request

```json
{
  "body": "Mr President...."
}
```

##### Response

```json
{
  "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
  "created_at": "2021-01-01T00:00:00Z",
  "updated_at": "2021-01-01T00:00:00Z",
  "body": "Hello, world!",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### GET /api/chirps

Responds with the list of chirps.

##### OPTIONAL Query parameters

* If the `author_id={user_id}` is set then the list of chirps will contain the author's chirps only.
* If the `sort=asc` is set or NOT set then the list of chirps will be sorted in the ascending order by `created_at` field.
* If the `sort=desc` is set then the list of chirps will be sorted in the descending order by `created_at` field.

##### Response

```json
[
  {
    "id": "94b7e44c-3604-42e3-bef7-ebfcc3efff8f",
    "created_at": "2021-01-01T00:00:00Z",
    "updated_at": "2021-01-01T00:00:00Z",
    "body": "Yo fam this feast is lit ong",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  },
  {
    "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
    "created_at": "2022-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "body": "What's good king?",
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  }
]
```

#### GET /api/chirps/{chirp_id}

Responds with the chirp by its id.

##### Response

```json
{
  "id": "f0f87ec2-a8b5-48cc-b66a-a85ce7c7b862",
  "created_at": "2022-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "body": "What's good king?",
  "user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

#### DELETE /api/chirps/{chirp_id}

If the passed user is authenticated deletes the chirp by its id.

##### Authentication

Headers: `Authorization: Bearer {the user's JWT}`

### Web-hooks

#### POST /api/polka/webhooks

If the `event` field is `"user.upgraded"` then the passed user is upgraded to "Chirpy Red" premium subscription.

##### Authentication

Headers: `Authorization: ApiKey POLKA_API_KEY`

##### Request

```json
{
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000"
  },
  "event": "user.upgraded"
}
```

### Maintenance

#### GET /api/healthz

If the server is up returns "OK" as plain text .
