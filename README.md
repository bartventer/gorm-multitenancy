# gorm-multitenancy

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v8)
[![Release](https://img.shields.io/github/release/bartventer/gorm-multitenancy.svg)](https://github.com/bartventer/gorm-multitenancy/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy/v8)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy/v8)
[![codecov](https://codecov.io/gh/bartventer/gorm-multitenancy/graph/badge.svg?token=6i0Pr1GFek)](https://codecov.io/gh/bartventer/gorm-multitenancy)
[![Tests](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gorm-multitenancy)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)

<p align="center">
  <img src="https://i.imgur.com/bOZB8St.png" title="GORM Multitenancy" alt="GORM Multitenancy">
</p>
<p align="center">
  <sub><small>Photo by <a href="https://github.com/ashleymcnamara">Ashley McNamara</a>, via <a href="https://github.com/ashleymcnamara/gophers">ashleymcnamara/gophers</a> (CC BY-NC-SA 4.0)</small></sub>
</p>

## Overview

Gorm-multitenancy provides a Go framework for building multi-tenant applications, streamlining
tenant management and model migrations. It abstracts multitenancy complexities through a unified,
database-agnostic API compatible with GORM.

## Multitenancy Approaches

There are three common approaches to multitenancy in a database:

- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

Depending on the database in use, this package utilizes either the _"shared database, separate schemas"_ or _"separate databases"_ strategy, ensuring a smooth integration with your existing database configuration through the provision of tailored drivers.

## Features

- **GORM Integration**: Simplifies [GORM](https://gorm.io/) usage in multi-tenant environments, offering a unified API alongside direct access to driver-specific APIs for flexibility.
- **Custom Database Drivers**: Enhances existing drivers for easy multitenancy setup without altering initialization.
- **HTTP Middleware**: Provides middleware for easy tenant context management in web applications.

## Supported Databases

| Database | Approach |
|----------|----------|
| PostgreSQL | Shared database, separate schemas |
| MySQL | Separate databases |

## Router Integration

- Echo - [Guide](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/echo/v8)
- Gin - [Guide](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/gin/v8)
- Iris - [Guide](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/iris/v8)
- Net/HTTP - [Guide](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8)

## Installation

Install the core package:

```bash
go get -u github.com/bartventer/gorm-multitenancy/v8
```

Install the database-specific driver:

```bash
# PostgreSQL
go get -u github.com/bartventer/gorm-multitenancy/postgres/v8

# MySQL
go get -u github.com/bartventer/gorm-multitenancy/mysql/v8
```

Optionally, install the router-specific middleware:

```bash
# Echo
go get -u github.com/bartventer/gorm-multitenancy/middleware/echo/v8

# Gin
go get -u github.com/bartventer/gorm-multitenancy/middleware/gin/v8

# Iris
go get -u github.com/bartventer/gorm-multitenancy/middleware/iris/v8

# Net/HTTP
go get -u github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8
```

## Getting Started

Check out the [pkg.go.dev](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v8) documentation for comprehensive guides and API references.

### Running the Example Application

For a practical demonstration, you can run [the example application](./_examples/README.md). It showcases various configurations and usage scenarios.

## Contributing

All contributions are welcome! See the [Contributing Guide](CONTRIBUTING.md) for more details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.