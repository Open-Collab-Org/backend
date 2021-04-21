[![Go Report Card](https://goreportcard.com/badge/github.com/open-collaboration/server)](https://goreportcard.com/report/github.com/open-collaboration/server)

## Development environment setup

This section assumes your terminal's current working directory is the project's directory.

Required tools:
- python
- [pip](https://pip.pypa.io/en/stable/installing/#installing-with-get-pip-py)

1. Install `gocyclo` and `gocritic`
```
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/go-critic/go-critic/cmd/gocritic@latest
```
2. Install `pre-commit`
```
pip install pre-commit && pre-commit install
```
3. Install [`docker`](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository)
4. Install [`docker-compose`](https://docs.docker.com/compose/install/#install-compose-on-linux-systems)



## Running the project

To start the database:
```
docker-compose up -d
```

To stop the database:
```
docker-compose down
```

To run the server:
```
go run .
```

## Contribution guidelines

### Modifying the database's schema
To modify the database's schema, you should create a migration in the [`migrations` package](./migrations).
Do not delete or modify existing migrations in `migrations.go`, only add more migrations to it with an increasing
id.

### Creating routes
Don't create gin routes with "pure" gin route handlers (functions with the signature `func(*gin.Context)`), instead
use the `createRouteHandler` method, which will be able to provide your handler with a database connection and
automatic error handling.

### Globals
Don't use globals. Ever. They make it harder to test the code. Instead, use depencency injection.

In a nutshell, dependency injection basically means receiving all of a method's *dependencies* (e.g. database connections
like `*gorm.DB`) as parameters. This way the caller of the function has to provide it with its dependencies and testing
the function later on is easier, we just have to pass it mocked or real dependencies as parameters, no need to fiddle
around with singletons and globals and whatnot.

### DTOs
Use DTOs (Data Transfer Objects) to send and receive data on routes. DTOs are basically just plain structs with
fields and field tags for validation (take a look at `NewUserDto`).


