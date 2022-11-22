# grpc-service-user

This is a small basic grpc service for managing user data stored in postgres database.

## Building and Running Project

This project compiles to a binary that runs two different commands:
- `server` command to start the grpc server, and
- `e2e` command to run e2e basic tests.

Make file includes the following entries:

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

## Design

I decided to use postgres to store the user data. Postgres is pretty much a good choice for such a use case. For simplicity, I decided to use gorm lib to implement the data access layer. 

To allow other services get notified when changes to use data happens, I decided to implement web hook for that.
The service will fire `http post` request to list of web hooks (configured as command line arguments passed to `server` command).

Example:

When the service started with:
```sh
./app server -notify="https://service1.app,https://service2.app"
```
The service will fire post requests to the paths /add /delete /update, with the changed user data encoded as json to https://service1.app and https://service2.app hosts.

