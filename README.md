# Zapp - Payment Gateway

A distributed payment gateway built in Go with a React admin dashboard. Features saga pattern orchestration, exactly-once semantics, and double-entry bookkeeping.

![Go](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat&logo=go)
![React](https://img.shields.io/badge/react-18-61DAFB?style=flat&logo=react)
![TypeScript](https://img.shields.io/badge/typescript-5-3178C6?style=flat&logo=typescript)
![Status](https://img.shields.io/badge/status-active-green)

## Features

- **Payment Processing**: Create, authorize, capture, cancel, and refund payments
- **Settlement Management**: ACH and wire transfer processing
- **Double-Entry Ledger**: Full chart of accounts with balanced journal entries
- **Multi-Service Architecture**: Microservices for payments, settlements, and ledger
- **Modern Admin Dashboard**: React frontend with enterprise styling

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.21+, Chi router |
| Frontend | React 18, TypeScript, Vite, Tailwind |
| Database | CockroachDB (PostgreSQL-compatible) |
| Cache | Redis |
| Message Queue | Kafka |
| Tracing | OpenTelemetry, Jaeger |

## Quick Start

### Prerequisites

- Go 1.21+
- Docker Desktop
- Node.js 18+

### 1. Start Infrastructure

```bash
docker-compose -f configs/docker-compose.yaml up -d
```

Services started:
- CockroachDB (port 26257)
- Redis (port 6379)
- Kafka (port 9092)
- Jaeger (port 16686)

### 2. Run Migrations

```bash
go run cmd/migrate/main.go up
```

### 3. Start Backend Services

```bash
# Terminal 1 - API Gateway (port 8080)
go run cmd/api/main.go

# Terminal 2 - Payment Service (port 8082)
go run cmd/payment/main.go

# Terminal 3 - Ledger Service (port 8081)
go run cmd/ledger/main.go

# Terminal 4 - Settlement Service (port 8083)
go run cmd/settlement/main.go
```

### 4. Start Frontend

```bash
cd zapp
npm install
npm run dev -- --host
```

### 5. Access

- **Frontend**: http://localhost:5173
- **API Gateway**: http://localhost:8080
- **Payment Service**: http://localhost:8082
- **Ledger Service**: http://localhost:8081
- **Settlement Service**: http://localhost:8083
- **CockroachDB SQL**: localhost:26257

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ZAPP FRONTEND (React)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Accounts    │  │ Payments     │  │ Settlements  │  │ Ledger       │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                               REST API
                                     │
        ┌────────────────────────────┼────────────────────────────┐
        ▼                            ▼                            ▼
┌───────────────┐          ┌───────────────┐          ┌───────────────┐
│   Payment     │          │   Settlement  │          │    Ledger     │
│   Service     │          │   Service     │          │    Service    │
│  (port 8082) │          │  (port 8083)  │          │  (port 8081)  │
└───────────────┘          └───────────────┘          └───────────────┘
        │                            │                            │
        └────────────────────────────┼────────────────────────────┘
                                     ▼
                            ┌───────────────┐
                            │ CockroachDB   │
                            │   (PostgreSQL)│
                            └───────────────┘
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | Entry point, proxies to microservices |
| Payment Service | 8082 | Payment intents, authorization, capture |
| Ledger Service | 8081 | Chart of accounts, journal entries |
| Settlement Service | 8083 | ACH/wire payout processing |

## Dashboard

The frontend provides a modern admin interface with slate gray and lightning yellow styling:

- **Accounts**: View holdings by asset type, payments incoming, settlements outgoing
- **Payments**: Create, authorize, capture, cancel, refund transactions
- **Settlements**: Create and manage ACH/wire payouts
- **Ledger**: View chart of accounts and journal entries

## API Endpoints

### Payment Service

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/payments` | Create payment |
| GET | `/api/v1/payments?merchant_id=` | List payments |
| GET | `/api/v1/payments/:id` | Get payment |
| POST | `/api/v1/payments/:id/authorize` | Authorize payment |
| POST | `/api/v1/payments/:id/capture` | Capture payment |
| POST | `/api/v1/payments/:id/cancel` | Cancel payment |
| POST | `/api/v1/payments/:id/refund` | Refund payment |
| GET | `/api/v1/payments/:id/charges` | Get charges |

### Settlement Service

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/settlements` | Create settlement |
| GET | `/api/v1/settlements?merchant_id=` | List settlements |
| GET | `/api/v1/settlements/:id` | Get settlement |
| POST | `/api/v1/settlements/:id/process` | Process settlement |
| POST | `/api/v1/settlements/:id/complete` | Mark complete |
| POST | `/api/v1/settlements/:id/cancel` | Cancel settlement |

### Ledger Service

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/ledger/accounts` | List accounts |
| GET | `/api/v1/ledger/accounts/:id` | Get account |
| GET | `/api/v1/ledger/accounts/:id/balance` | Get balance |
| GET | `/api/v1/ledger/balances` | Get all balances |
| GET | `/api/v1/ledger/journals` | List journals |
| GET | `/api/v1/ledger/journals/:id` | Get journal with lines |
| POST | `/api/v1/ledger/journals` | Create journal entry |

## Project Structure

```
zapp/
├── cmd/                         # Service entry points
│   ├── api/                     # API Gateway
│   ├── payment/                 # Payment service
│   ├── ledger/                  # Ledger service
│   ├── settlement/             # Settlement service
│   ├── saga/                    # Saga orchestrator
│   └── migrate/                 # Database migrations
├── services/                    # Business logic layer
│   ├── payment/
│   ├── settlement/
│   ├── ledger/
│   └── saga/
├── internal/                    # Shared packages
│   ├── idempotency/            # Redis idempotency
│   ├── outbox/                 # Outbox pattern
│   ├── circuitbreaker/         # Circuit breaker
│   ├── ratelimit/              # Rate limiting
│   ├── tracing/                # OpenTelemetry
│   └── messaging/              # Kafka producers
├── migrations/                 # SQL migrations
├── configs/                    # Docker configs
└── zapp/                       # React frontend
    ├── src/
    │   ├── api/                # API client
    │   ├── components/         # UI components
    │   ├── pages/              # Dashboard pages
    │   └── types/              # TypeScript types
    └── dist/                   # Built output
```

## Design Principles

1. **Integer Currency**: All amounts use `int64` in cents, never float
2. **Double-Entry**: Every transaction creates balanced debit/credit entries
3. **Outbox Pattern**: Reliable event publishing for exactly-once delivery

## License

MIT
