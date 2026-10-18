# CS 6650 HW05 - Product API

For this assignment, I built a simple Product API for an online store using Go and the Gin framework. The API follows the provided OpenAPI specification (`api.yaml`) and supports fetching product information and updating existing product details.

I first implemented and tested the API locally, then containerized it using Docker and deployed it to AWS ECS using Terraform. I also used Locust for load testing and Prometheus and Grafana to monitor the application.

## Repository Structure

```text
.
├── src/              Go server code and Dockerfile
├── terraform/        Terraform files for AWS infrastructure
├── locust/           Locust load testing scripts
├── observability/    Prometheus and Grafana configuration
└── docs/images/      Screenshots of API responses
```

## Part II: Product API

### API Endpoints

The API has two endpoints, as defined in the OpenAPI specification.

| Method | Endpoint                        | Description                               | Response Codes     |
| ------ | ------------------------------- | ----------------------------------------- | ------------------ |
| GET    | `/products/{productId}`         | Fetch product details using its ID        | 200, 404, 500      |
| POST   | `/products/{productId}/details` | Update the details of an existing product | 204, 400, 404, 500 |

### Product Storage

For this assignment, I used an in-memory Go map (`map[int]Product`) to store the products instead of using a database.

I added four sample products when the server starts so that I could test the API without having to create products manually.

| Product ID | SKU       | Manufacturer |
| ---------- | --------- | ------------ |
| 1100       | GO-1100   | Go Server    |
| 1200       | TS-1200   | Typescript   |
| 1300       | AMZ-1300  | Amazon       |
| 1400       | NODE-1400 | Node Server  |

Since the products are stored in memory, any changes made through the POST endpoint are lost when the server restarts. The server loads the original four products again on startup.

### Implementation Decisions

While implementing the API, I made a few decisions based on the OpenAPI specification and how the endpoints were expected to behave.

**1. Updating existing products**

The specification describes the POST endpoint as adding or updating product details, but it also includes a 404 response when a product is not found.

I decided to use POST only for updating products that already exist. If a product ID is not present in the map, the API returns 404.

Since there is no separate endpoint for creating products, I initialized the map with four sample products.

**2. Handling invalid product IDs**

For the GET endpoint, the specification only defines 200, 404, and 500 responses.

Because of this, I return 404 when the requested product ID is invalid or does not exist. For example, IDs such as `abc` or `0` cannot match a valid product.

For POST requests, invalid product IDs return 400 because that response code is included in the specification.

**3. Checking product IDs during updates**

The POST endpoint receives the product ID in both the URL and the request body.

I added a validation check to make sure both IDs match. If they are different, the API returns 400 instead of updating the wrong product.

**4. Handling concurrent requests**

I used a Go map because products are accessed using their IDs, which makes lookups fast.

However, Gin can handle multiple requests concurrently, and a regular Go map is not safe when multiple goroutines access it while writes are happening.

To handle this, I used `sync.RWMutex`.

- `RLock()` allows multiple requests to read product information at the same time.
- `Lock()` allows only one request to update the map at a time.

This helps prevent race conditions when multiple requests access the product data.

**5. Testing internal server errors**

Since the API uses in-memory storage, there are not many situations where a normal request would cause an internal server error.

To test the 500 response, I added support for the header:

`X-Simulate-Error: true`

When this header is included, the handler intentionally triggers a panic.

I also added custom recovery middleware to catch the panic and return a 500 response using the error format defined in the specification.

This also allows unexpected panics to be handled without crashing the entire server.

## Running the API Locally

The application can be run directly using Go or through Docker.

### Option 1: Using Go

Navigate to the `src` directory and start the server.

```bash
cd src
go run .
```

### Option 2: Using Docker

Build the Docker image:

```bash
cd src
docker build -t product-service .
```

Run the container:

```bash
docker run -p 8080:8080 product-service
```

In both cases, the API runs on port `8080`.

## Testing the API

I tested the API using `curl` to verify the different response codes mentioned in the OpenAPI specification.

The following commands can be used to test the API locally or against the deployed ECS service.

First, set the host URL:

```bash
HOST=http://localhost:8080
```

For testing on AWS ECS, replace `localhost` with the public IP address of the running service.

### 1. GET - 200 OK

Fetch details of an existing product.

```bash
curl -i $HOST/products/1100
```

Expected response:

```json
{
  "product_id": 1100,
  "sku": "GO-1100",
  "manufacturer": "Go Server",
  "category_id": 1,
  "weight": 500,
  "some_other_id": 10
}
```

**Status Code:** `200 OK`

### 2. GET - 404 Not Found

Try fetching a product that does not exist.

```bash
curl -i $HOST/products/9999
```

Expected response:

```json
{
  "error": "NOT_FOUND",
  "message": "Product not found",
  "details": "No product exists with ID 9999"
}
```

**Status Code:** `404 Not Found`

### 3. GET - 404 Invalid Product ID

Send a request with an invalid product ID.

```bash
curl -i $HOST/products/abc
```

**Status Code:** `404 Not Found`

Since the product ID is not a valid positive integer, the API returns an error message.

### 4. GET - 500 Internal Server Error

I used the custom header to simulate an internal server error.

```bash
curl -i $HOST/products/1100 \
  -H "X-Simulate-Error: true"
```

**Status Code:** `500 Internal Server Error`

The recovery middleware catches the panic and returns an error response.

### 5. POST - 204 No Content

Update the details of an existing product.

```bash
curl -i -X POST $HOST/products/1100/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1100,
    "sku": "GO-1100-V2",
    "manufacturer": "Go Server Inc",
    "category_id": 1,
    "weight": 750,
    "some_other_id": 10
  }'
```

**Status Code:** `204 No Content`

The product details are updated successfully. Since the response is 204, no response body is returned.

### 6. POST - 400 Invalid URL ID

Try updating a product using an invalid ID in the URL.

```bash
curl -i -X POST $HOST/products/abc/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1100,
    "sku": "GO-1100",
    "manufacturer": "Go Server",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 10
  }'
```

**Status Code:** `400 Bad Request`

### 7. POST - 400 Invalid Request Body

Send a request with an empty SKU and a negative weight.

```bash
curl -i -X POST $HOST/products/1100/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1100,
    "sku": "",
    "manufacturer": "Go Server",
    "category_id": 1,
    "weight": -5,
    "some_other_id": 10
  }'
```

**Status Code:** `400 Bad Request`

The request fails because the product details do not pass validation.

### 8. POST - 400 Product ID Mismatch

Send different product IDs in the URL and request body.

```bash
curl -i -X POST $HOST/products/1100/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1200,
    "sku": "GO-1100",
    "manufacturer": "Go Server",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 10
  }'
```

**Status Code:** `400 Bad Request`

The API checks that both IDs match before updating the product.

### 9. POST - 404 Product Not Found

Try updating a product that is not present in the map.

```bash
curl -i -X POST $HOST/products/9999/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 9999,
    "sku": "NEW-9999",
    "manufacturer": "Nobody",
    "category_id": 1,
    "weight": 100,
    "some_other_id": 1
  }'
```

**Status Code:** `404 Not Found`

Since product `9999` does not exist, the API does not create a new product.

### 10. POST - 500 Internal Server Error

Simulate a server error while updating a product.

```bash
curl -i -X POST $HOST/products/1100/details \
  -H "Content-Type: application/json" \
  -H "X-Simulate-Error: true" \
  -d '{
    "product_id": 1100,
    "sku": "GO-1100",
    "manufacturer": "Go Server",
    "category_id": 1,
    "weight": 500,
    "some_other_id": 10
  }'
```

**Status Code:** `500 Internal Server Error`

The custom recovery middleware handles the simulated error.

## API Testing Screenshots

Below are the screenshots of the API responses tested against the deployed ECS service.

<!-- Add screenshots here -->

<!-- Example:
![GET 200](docs/images/get-200.png)
![GET 404](docs/images/get-404.png)
![POST 204](docs/images/post-204.png)
-->

<img width="730" height="794" alt="Screenshot 2026-10-17 at 9 42 18 PM" src="https://github.com/user-attachments/assets/ff3f3455-1baf-4381-a953-56d2c9382d1e" />

