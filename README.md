# Friendly Octo Guacamole

helmfile --environment production diff
helmfile --environment production sync


# Food Delivery Platform - SRE Interview Exercise

## Overview

This is a minimal Go application designed for SRE interview exercises. It simulates a food delivery platform's menu service API with intentional failure scenarios.

## Application Endpoints

| Endpoint | Method | Description | Response Codes |
|----------|--------|-------------|----------------|
| `/health` | GET | Health check endpoint | `200` (always) |
| `/api/menu` | GET | List all menu items | `200`, `500` (10% chance) |
| `/api/menu/{id}` | GET | Get menu item by ID | `200`, `404` |

## Running the Application

### Local Development

```bash
# Run directly with Go
go run main.go

# Or build and run
go build -o mock-service
./mock-service
```

The server will start on port `8080`.

### Using Docker

```bash
# Build the image
docker build -t mock-service .

# Run the container
docker run -p 8080:8080 mock-service
```

### Testing Endpoints

```bash
# Health check - Always returns 200
curl -i http://localhost:8080/health

# List menu items - Usually 200, sometimes 500
curl -i http://localhost:8080/api/menu

# Get existing menu item - Returns 200
curl -i http://localhost:8080/api/menu/1

# Get non-existent menu item - Returns 404
curl -i http://localhost:8080/api/menu/999
```

## Structured Logging

All logs are output in JSON format with the following fields:

```json
{
  "timestamp": "2025-11-22T10:30:45Z",
  "level": "INFO",
  "method": "GET",
  "path": "/api/products",
  "status_code": 200,
  "duration_ms": 2.34,
  "request_id": "req-1234567890-1",
  "menu_item_id": "123",
  "message": "Listed 5 menu items",
  "error": "Database connection timeout"
}
```

### Log Fields Explanation

- `timestamp` - RFC3339 formatted timestamp
- `level` - Log level (INFO, WARN, ERROR)
- `method` - HTTP method
- `path` - Request path
- `status_code` - HTTP response code
- `duration_ms` - Request processing time
- `request_id` - Unique request identifier
- `menu_item_id` - Menu item ID (when applicable)
- `message` - Human-readable message
- `error` - Error details (only on failures)