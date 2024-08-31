# Examples

This directory provides an example demonstrating the use of the `gorm-multitenancy` package with various configurations:

## Getting Started

To run this example on your local machine, follow these steps:

### Step 1: Clone the Repository

```bash
git clone https://github.com/bartventer/gorm-multitenancy.git
cd gorm-multitenancy
```

### Step 2: Run the Example

This example supports various configurations through command-line options.

#### Usage

```bash
go run -C examples . [options]
```

#### Options

-   `-server` string

    -   Description: Specifies the HTTP server to run (_and `gorm-multitenancy` middleware to use_)
    -   Options: [`echo`](../middleware/echo/README.md), [`gin`](../middleware/gin/README.md), [`iris`](../middleware/iris/README.md), [`nethttp`](../middleware/nethttp/README.md)
    -   Default: [`echo`](../middleware/echo/README.md)

-   `-driver` string
    -   Description: Specifies the `gorm-multitenancy` database driver.
    -   Options: [`postgres`](../postgres/README.md), [`mysql`](../mysql/README.md)
    -   Default: [`postgres`](../postgres/README.md)

#### Examples

-   Run with default options:

    ```bash
    go run -C examples .
    ```

-   Run with the `NetHTTP` server and `MySQL` driver:

    ```bash
    go run -C examples . -server=nethttp -driver=mysql
    ```

> [!NOTE]
> To enable debug logging, set the GMT_DEBUG environment variable to true. This can be helpful for troubleshooting or understanding the internal workings of the application.

## Interacting with the API

Please see [API Usage](USAGE.md) for more examples on how to interact with the server.

## Contributing

All contributions are welcome! See the [Contributing Guide](../CONTRIBUTING.md) for more details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../LICENSE) file for details.