# gorm-multitenancy

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gorm-multitenancy/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gorm-multitenancy?branch=master)
[![Build](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)

There are three common approaches to multitenancy in a database:
- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

This package implements the shared database, separate schemas approach. It uses the [gorm](https://gorm.io/) ORM to manage the database and provides custom drivers to support multitenancy. It also provides HTTP middleware to retrieve the tenant from the request and set the tenant in context.

## Database compatibility
Current supported databases are listed below. Pull requests for other drivers are welcome.
- [PostgreSQL](https://www.postgresql.org/)

## Router compatibility
Current supported routers are listed below. Pull requests for other routers are welcome.
- [echo](https://echo.labstack.com/docs)
- [net/http](https://golang.org/pkg/net/http/)

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy
```

## Examples

- [PostgreSQL with echo](https://github.com/bartventer/gorm-multitenancy/tree/master/internal/examples/echo)
- [PostgreSQL with net/http](https://github.com/bartventer/gorm-multitenancy/tree/master/internal/examples/nethttp)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.