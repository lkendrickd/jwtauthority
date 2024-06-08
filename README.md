## JWTAuthority

<img src="images/echo-server.webp" alt="Echo Server Logo" width="400"/>

This is a jwt authorization server that generates jwt tokens for user accounts.

### Features:

- [x] JWT HTTP Server
- [x] Routing
- [x] Middleware
- [x] Structured Logging
- [x] Prometheus Metrics
- [x] Flag and Environment Variable Configuration

### Endpoints:

- `GET /health`: Returns the health of the server
- `POST api/v1/login`: Returns a JWT token for the user
- `GET api/v1/protected`: Endpoint to test the JWT token

### Usage:

```bash
go run cmd/jwtauthorizor.go
```

### Configuration:

Note that environment variables for PORT and LOG_LEVEL take precedence over the flags.

### Make Native Go Execution:

```bash
make build
make run
```

#### Docker Execution:

```bash
make docker-run
```

#### Curl Examples:
Health Check:
```bash
curl http://localhost:8080/health
```
JWT Token Acquisition:
```bash
curl -X POST -d '{"username":"admin","password":"adminpassword"}' -H "Content-Type: application/json" http://localhost:8080/login
```
Protected Endpoint:
```bash
TOKEN="your_jwt_token"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/protected
```

Metrics:
```bash
curl http://localhost:8080/metrics
```

### Cleanup - When done with Docker Execution:

```bash
make docker-clean
```

### Expansion:

To add on or remove an endpoint just manipulate this section under server/server.go

```go
func (s *Server) SetupRoutes() {}
```

Then add a handler for your route under handlers/handlers.go it's that simple.

This is to show that frameworks really are unnecessary for microservices with the new features in Go 1.22.
