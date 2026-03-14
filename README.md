# Zapp - Payment Gateway

A distributed payment gateway with saga pattern orchestration, exactly-once semantics, and double-entry ledger built in Go, with a React admin dashboard.

## Prerequisites

1. **Go 1.24+** - [Install Go](https://golang.org/doc/install)
2. **Docker Desktop** - For running infrastructure services
3. **Node.js 18+** - For the frontend

## Quick Start

### 1. Start Infrastructure

```bash
docker-compose -f configs/docker-compose.yaml up -d
```

### 2. Run Migrations

```bash
go run cmd/migrate/main.go up
```

### 3. Start Backend Services

```bash
# Terminal 1 - API Gateway
go run cmd/api/main.go

# Terminal 2 - Payment Service
go run cmd/payment/main.go

# Terminal 3 - Ledger Service
go run cmd/ledger/main.go

# Terminal 4 - Settlement Service
go run cmd/settlement/main.go
```

### 4. Start Frontend

```bash
cd zapp
npm install
npm run dev
```

### 5. Access Zapp

- **Frontend**: http://localhost:5173
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Gateway**: http://localhost:8080
- **Payment Service**: http://localhost:8082
- **Ledger Service**: http://localhost:8081
- **Settlement Service**: http://localhost:8083
- **CockroachDB**: localhost:26257
- **Redis**: localhost:6379
- **Kafka**: localhost:9092

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ZAPP FRONTEND (React)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │ Payments    │  │ Settlements │  │ Ledger      │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                              REST API
                                    │
┌─────────────────────────────────────────────────────────────────────────────┐
│                           API GATEWAY (Port 8080)                          │
│  • Reverse proxy to services                                               │
│  • CORS support                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | Entry point, proxies to services |
| Payment Service | 8082 | Payment intents, transactions |
| Ledger Service | 8081 | Double-entry bookkeeping |
| Settlement Service | 8083 | ACH, wire transfers |
| Saga Orchestrator | 8084 | Distributed transactions |

## Zapp Dashboard

The Zapp frontend provides a modern admin interface:

- **Payments**: Create, authorize, capture, cancel, refund
- **Settlements**: Create and manage ACH/wire transfers
- **Ledger**: View accounts and journal entries

### Running the Frontend

```bash
cd zapp
npm install
npm run dev
```

The frontend connects to the API Gateway at `http://localhost:8080` (configurable via `VITE_API_URL` in `.env`).

## API Endpoints

### Payment Service

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/payments` | Create payment |
| GET | `/api/v1/payments?merchant_id=` | List payments |
| GET | `/api/v1/payments/:id` | Get payment |
| POST | `/api/v1/payments/:id/authorize` | Authorize payment |
| POST | `/api/v1/payments/:id/capture` | Capture authorized payment |
| POST | `/api/v1/payments/:id/cancel` | Cancel payment |
| POST | `/api/v1/payments/:id/refund` | Refund payment |

### Settlement Service

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/settlements` | Create settlement |
| GET | `/api/v1/settlements?merchant_id=` | List settlements |
| GET | `/api/v1/settlements/:id` | Get settlement |
| POST | `/api/v1/settlements/:id/process` | Start processing |
| POST | `/api/v1/settlements/:id/complete` | Mark complete |
| POST | `/api/v1/settlements/:id/cancel` | Cancel settlement |

### Ledger Service

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ledger/accounts` | List accounts |
| GET | `/api/v1/ledger/accounts/:id` | Get account |
| GET | `/api/v1/ledger/accounts/:id/balance` | Get balance |
| GET | `/api/v1/ledger/journals` | List journals |
| GET | `/api/v1/ledger/journals/:id` | Get journal with lines |

## Testing

### Swagger UI

Interactive API documentation is available at: http://localhost:8080/swagger/index.html

All endpoints are documented and can be tested directly from the browser.

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Project Structure

```
zapp/
├── cmd/                    # Service binaries
│   ├── api/                # API Gateway
│   ├── payment/            # Payment service
│   ├── ledger/             # Ledger service
│   ├── settlement/         # Settlement service
│   ├── saga/               # Saga orchestrator
│   └── migrate/            # DB migrations
├── internal/               # Internal packages
│   ├── idempotency/        # Idempotency middleware (Redis)
│   ├── outbox/             # Outbox pattern
│   ├── circuitbreaker/     # Circuit breaker
│   ├── ratelimit/         # Rate limiting
│   ├── tracing/           # OpenTelemetry
│   └── messaging/          # Kafka abstraction
├── services/              # Business logic
│   ├── payment/
│   ├── ledger/
│   ├── settlement/
│   └── saga/
├── migrations/             # Database migrations
├── pkg/                   # Shared packages
├── configs/               # Docker configs
└── zapp/                 # React frontend
    ├── src/
    │   ├── api/           # API client
    │   ├── components/    # UI components
    │   ├── pages/        # Dashboard pages
    │   └── types/        # TypeScript types
    └── dist/             # Built output
```

## Key Design Decisions

1. **Money Handling**: Uses `int64` in smallest currency unit (cents), NEVER float
2. **Double-Entry Ledger**: All transactions create balanced debit/credit entries
3. **Saga Pattern**: Distributed transactions with compensation
4. **Outbox Pattern**: Exactly-once event delivery
5. **Idempotency**: Redis-backed idempotency keys for safe retries

## License

MIT
