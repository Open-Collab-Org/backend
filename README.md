
## Development environment setup

This section assumes your terminal's current working directory is the project's directory.

Required tools:
- python
- [pip](https://pip.pypa.io/en/stable/installing/#installing-with-get-pip-py)

1. Install [`golangci-lint`](https://golangci-lint.run/usage/install/#local-installation)
2. Install `pre-commit`
```
pip install pre-commit && pre-commit install
```
3. Install [`docker`](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository)
4. Install [`docker-compose`](https://docs.docker.com/compose/install/#install-compose-on-linux-systems)


> **Tip:** Install the Go Linter plugin for GoLand/VSCode


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
go run src/main.go
```