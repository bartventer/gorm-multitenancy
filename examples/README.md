# Examples

This directory contains examples of how to use the `gorm-multitenancy` package with different frameworks:

-   [Echo](echo/README.md)
-   [net/http](nethttp/README.md)

## Getting Started

To run these examples, you have two options:

1. Clone the main repository and navigate to the relevant example directory:

    ```bash
    go get -u github.com/bartventer/gorm-multitenancy/v5
    ```

2. Or just clone the relevant example directory directly:

    _Echo example_:

    ```bash
    go get -u github.com/bartventer/gorm-multitenancy/v5/examples/echo
    ```

    _net/http example_:

    ```bash
    go get -u github.com/bartventer/gorm-multitenancy/v5/examples/nethttp
    ```

## Running the Server

Run the following command from the relevant example directory:

```bash
./run.sh
```

This will setup the database and run the server.

## API Usage

Please see [API Usage](USAGE.md) for more examples on how to interact with the server.
