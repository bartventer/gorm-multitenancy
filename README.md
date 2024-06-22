# gorm-multitenancy

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v7)
[![Release](https://img.shields.io/github/release/bartventer/gorm-multitenancy.svg)](https://github.com/bartventer/gorm-multitenancy/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy/v7)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy/v7)
[![codecov](https://codecov.io/gh/bartventer/gorm-multitenancy/graph/badge.svg?token=6i0Pr1GFek)](https://codecov.io/gh/bartventer/gorm-multitenancy)
[![Build](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gorm-multitenancy)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy?ref=badge_shield&issueType=license)

<p align="center">
  <img src="https://i.imgur.com/bOZB8St.png" title="GORM Multitenancy" alt="GORM Multitenancy">
</p>
<p align="center">
  <sub><small>Photo by <a href="https://github.com/ashleymcnamara">Ashley McNamara</a>, via <a href="https://github.com/ashleymcnamara/gophers">ashleymcnamara/gophers</a> (CC BY-NC-SA 4.0)</small></sub>
</p>

## Introduction

Gorm-multitenancy is a Go package that provides a framework for implementing multitenancy in applications using GORM.

## Multitenancy Approaches

There are three common approaches to multitenancy in a database:

- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

This package adopts the 'shared database, separate schemas' approach, providing custom drivers for seamless integration with your existing database setup.

## Features

- **GORM Integration**: Leverages the [gorm](https://gorm.io/) ORM to manage the database, facilitating easy integration with your existing GORM setup.
- **Custom Database Drivers**: Provides drop-in replacements for existing drivers, enabling multitenancy without the need for initialization reconfiguration.
- **HTTP Middleware**: Offers middleware for seamless integration with popular routers, making it easy to manage tenant context in your application.

## Database compatibility

The following databases are currently supported. Contributions for other drivers are welcome.

- PostgreSQL

## Router Integration

This package includes middleware that can be used with the routers listed below for seamless integration with the database drivers. While not a requirement, these routers are fully compatible with the provided middleware. Contributions for other routers are welcome.

- Echo
- Net/HTTP

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy/v7
```

## Getting Started

### Drivers

- PostgreSQL [Guide](./postgres/README.md)

### Middleware

- Echo [Guide](./middleware/echo/README.md)
- Net/HTTP [Guide](./middleware/nethttp/README.md)

## Contributing

All contributions are welcome! See the [Contributing Guide](CONTRIBUTING.md) for more details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy?ref=badge_large&issueType=license)
