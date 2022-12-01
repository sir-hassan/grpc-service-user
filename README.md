# grpc-service-user

This is a small basic grpc service for managing user data stored in postgres database.

## Building and Running Project

This project compiles to a binary that runs two different commands:
- `server` command to start the grpc server, and
- `e2e` command to run e2e basic tests.

You don't have to compile the binary manually as the project ships a Makefile which automate all the building and 
testing with docker-compose environment:

```shell
# to run unit tests:
make test

# to run e2e tests:
make test-e2e

# to lint project:
make lint

# to build a naked binary:
make build

# to generate protobuf files:
make generate

# to run project in a docker-compose env:
make composer-up
```

## GRPC Endpoints:

The following endpoint are implemented:
```shell
+-----------+-------------+--------------------+------------------+
|  SERVICE  |     RPC     |    REQUEST TYPE    |  RESPONSE TYPE   |
+-----------+-------------+--------------------+------------------+
| UserStore | CheckHealth | CheckHealthRequest | CheckHealthReply |
| UserStore | AddUser     | AddUserRequest     | AddUserReply     |
| UserStore | UpdateUser  | UpdateUserRequest  | UpdateUserReply  |
| UserStore | DeleteUser  | DeleteUserRequest  | DeleteUserReply  |
| UserStore | ListUsers   | ListUsersRequest   | User             |
+-----------+-------------+--------------------+------------------+
```

Refer to `api/user.proto` for more details about the endpoints and the requests and replies structures.

## Server Configurations

The server expects configurations via env vars. The following golang struct explains all the expected environment vars:
```go
type envVars struct {
    Port int `env:"PORT" envDefault:"8080"`
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"admin"`
	PostgresPassword string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"userdb"`

	NotifierWebHooks []string `env:"NOTIFIER_WEBHOOKS" envSeparator:","`
}
```

`NOTIFIER_WEBHOOKS` is a comma seperated string for all web hooks urls used to notify other systems upon user data changes.

## Design

### 1. Storing User Data

I decided to use postgres to store the user data. Postgres is pretty much a good choice for such a use case.
I used gorm library to simplify the data access layer code. I never used this lib before but wanted to give it a try.

### 2. Providing API to do CRUD

Only two options to choose from (`rest` and `grpc`). Typically, 
I would use `rest` for use case unless wire efficiency and latency is crucial.
Here, I used `grpc` just to be cool :).

### 3. Pagination and Filtering Endpoint

This feature is implemented via `UserStore.ListUsers` endpoint. Nothing special to design here.
Example call:

```shell

# using grpcurl tool:
grpcurl -plaintext -d '{"page": 2, "page_size": 6, "filters": {"email": "me@example.com"}  }'  localhost:8080  api.UserStore.ListUsers
```

##$ 2. Asynchronous Notification Mechanism

To allow other services getting notified when changes to use data happens, I decided to implement web hook for that.
A comma seperated string of webhooks should be configured via `NOTIFIER_WEBHOOKS`. 
The server will fire post requests asynchronously.
These requests encode both the changed user data (via request body) and the type of the change (via /add /delete /update) paths.

For simplicity, I used a very simple channel + single goroutine to implement a FIFO queue to queue the notifications to be sent asynchronously.
For larger and more serious systems, an auto-scaling jobs queue with more goroutines need to be implemented.
