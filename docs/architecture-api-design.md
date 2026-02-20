# API Design & Error Handling

This document outlines the design principles, architectural patterns, and error handling strategies for the stock service API. Our goal is to create a robust, predictable, and developer-friendly API that is both resilient and easy to maintain.

##  philosophies

- **RESTful Principles**: We adhere to RESTful principles for resource naming, HTTP verb usage, and status code representation.
- **Statelessness**: The API is stateless, meaning each request is independent and contains all necessary information. This is critical for horizontal scalability.
- **Developer Experience**: We prioritize a positive developer experience with clear documentation, predictable behavior, and informative error messages.
- **Security First**: All endpoints are designed with security in mind, including input validation, secure defaults, and protection against common vulnerabilities.

## API Design

### Endpoints

The API exposes the following endpoints:

- `GET /`: Retrieves stock data for the default symbol (e.g., MSFT).
- `GET /{symbol}`: Retrieves stock data for a specific symbol.
- `GET /{symbol}/{days}`: Retrieves stock data for a specific symbol and number of days.
- `GET /health`: Liveness probe for Kubernetes.
- `GET /ready`: Readiness probe for Kubernetes.
- `GET /metrics`: Prometheus metrics endpoint.
- `GET /docs`: Interactive API documentation (Scalar).
- `GET /circuit-breaker`: Returns the current state of the circuit breaker.

### Request & Response

- **Requests**: Requests use URL parameters for symbol and days, which are validated to prevent invalid input.
- **Responses**: All responses are in JSON format. Successful responses contain the requested data, while error responses provide a clear error message.

### Input Validation

- **Symbol**: Validated to ensure it contains only alphanumeric characters.
- **Days**: Validated to ensure it is a positive integer within a reasonable range (e.g., 1-100).

## Error Handling

Our error handling strategy is designed to be consistent, informative, and secure.

### Error Types

We classify errors into two main categories:

- **Client Errors (4xx)**: Errors caused by invalid client input, such as a malformed request or invalid stock symbol.
- **Server Errors (5xx)**: Errors caused by a server-side issue, such as a database connection failure or an external API outage.

### Error Response Format

All error responses follow a consistent JSON format:

```json
{
  "error": "a brief, human-readable error message",
  "details": "optional, more detailed information for debugging"
}
```

This format provides a clear error message for the client, with optional details for debugging purposes. In production, detailed error messages are logged on the server but not exposed to the client to prevent information leakage.

### Centralized Error Handling

We use a centralized error handling mechanism in our HTTP handlers to ensure all errors are handled consistently. The `sendJSON` helper function is used to send both successful and error responses, which simplifies the handler logic and ensures a consistent response format.

### Graceful Degradation

In the event of a failure with the external Alpha Vantage API, the service will:

1.  **Serve from Cache**: If the requested data is available in the cache, it will be served to the client, ensuring continued service availability.
2.  **Circuit Breaker**: If the external API is down, the circuit breaker will open, preventing further calls and reducing the load on the external service.
3.  **Return 503**: If the data is not in the cache and the circuit breaker is open, the service will return a `503 Service Unavailable` error, indicating that the request cannot be fulfilled at this time.

## Logging

We use structured logging with [Zap](https://github.com/uber-go/zap) to ensure all log entries are machine-readable and contain rich context. All logs are written to `stdout` in JSON format, which is ideal for containerized environments and log aggregation platforms like Elasticsearch or Splunk.

### Log Fields

Each log entry includes:

- `level`: Log level (e.g., `info`, `warn`, `error`)
- `ts`: Timestamp
- `msg`: Log message
- `method`: HTTP request method
- `path`: HTTP request path
- `duration`: Request processing time
- `status`: HTTP response status code

This structured approach allows us to easily search, filter, and analyze logs, which is critical for debugging and incident response.
