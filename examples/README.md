# Examples

This directory provides examples demonstrating the use of the `gorm-multitenancy` package with various frameworks:

- [Echo](echo/README.md)
- [net/http](nethttp/README.md)

## Getting Started

To run these examples on your local machine, follow these steps:

### Step 1: Clone the Repository

```bash
git clone https://github.com/bartventer/gorm-multitenancy.git
cd gorm-multitenancy
```

### Step 2: Execute the Desired Example

Choose an example to run. This will spin up the necessary services and start the server.

#### For `Echo`

```bash
make echo_example
```

#### For `net/http`

```bash
make nethttp_example
```

## Interacting with the API

Please see [API Usage](USAGE.md) for more examples on how to interact with the server.
