# Solutional Homework Go REST API

This is a Go REST API project that replicates the functionality of a reference API running at https://homework.solutional.ee.

The reference API implements a very simple e-commerce cart/order flow, where products can be added, modified, and replaced within an order. The project also allows the listening port number to be configured without changing the code.

The project is not intended to be used in production, and the orders are not persisted between application restarts.
## Installation

To clone the project, run the following command:

```bash
git clone https://github.com/blackcloro/go-homework.git
```

## Usage

To run the project locally, you can use the `go run` command with the `--port` flag to specify a custom port. The default port is `3000`. For example:

```bash
go run ./cmd/api --port=":8080"
```

To run the project with Docker, you can use the `docker build` and `docker run` command with the `-p` flag to map the container port to the host port. You can also use the `-e` flag to set the `PORT` environment variable inside the container. For example:

```bash
docker build -t go-rest-api .
docker run -p 8090:8090 -e PORT=8090 go-rest-api
```

## API Endpoints

The project exposes the following API endpoints:

- `GET /api/products` - list of all available products
- `POST /api/orders` - create a new order
- `GET /api/orders/:order_id` - get order details
- `PATCH /api/orders/:order_id` - update an order
- `GET /api/orders/:order_id/products` - get order products
- `POST /api/orders/:order_id/products` - add products to the order
- `PATCH /api/orders/:order_id/products/:product_id` - update product quantity
- `PATCH /api/orders/:order_id/products/:product_id` - add a replacement product

## Testing

To run the tests, use the `go test` command:

```bash
go test ./...
```
