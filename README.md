# demo-monetized-api

This is a system design project showing a stripped-down monetized API running on my personal infrastructure. It is composed of Go microservices (API, Validator, and Billing) fronted by Kong as the API Gateway. Redis is used for caching authentication and credit information, with SQLite as persistence on the API service. The system is designed to demonstrate a pay-per-use API model, where consumers are charged credits for each API call based on the complexity of the request.

## System Architecture

```mermaid
sequenceDiagram
    participant User as Internet / Client
    participant Kong as Kong (API Gateway)
    participant Validator as Go Validator Service
    participant API as Go API Service
    participant Billing as Go Billing Service
    participant Redis as Redis (Auth/Credits)

    User->>Kong: Request with X-API-Key Header
    
    Kong->>Validator: GET /api/v1/validate (via pre-function plugin)
    Validator->>Redis: GET <api-key>
    Redis-->>Validator: resCredits > 0
    Validator-->>Kong: 200 OK

    Kong->>API: Proxy Request to /service
    API-->>Kong: 200 OK (Data)
    Kong-->>User: Final Response

    Note over User,Redis: Billing settlement is triggered manually in this demo
    User->>Billing: POST /api/v1/billing?amount=X
    Billing->>Redis: SET <api-key> <new-balance>
```

## Getting Started

### Prerequisites
- Docker and Docker Compose installed.

### Start the Demo
1. Clone the repository and navigate to the project root.
2. Spin up the entire stack:
   ```bash
   docker compose up --build
   ```
3. The services will be available at:
   - **Kong Gateway**: http://localhost:8000
   - **Go API**: http://localhost:8082
   - **Go Billing**: http://localhost:8081
   - **Go Validator**: http://localhost:8080

## Demo Endpoints

You can test the monetization flow using the following `curl` commands. The system initializes with a demo key containing **100 credits**.

**Demo API Key:** `demo-api-key-123`

### 1. Access the Protected API
Access the API through the Kong Gateway. The gateway will automatically call the validator service:
```bash
curl -H "X-API-Key: demo-api-key-123" http://localhost:8000/service
```

### 2. Settle Billing (Deduct Credits)
Manually trigger a billing settlement to simulate usage costs:
```bash
curl -X POST -H "X-API-Key: demo-api-key-123" "http://localhost:8081/api/v1/billing?amount=10"
```

### 3. Verify Insufficient Credits
If you deduct enough credits to reach zero, the Kong gateway (via the validator) will block further requests with a `402 Payment Required` error.

## Data Flow

Redis is used as a shared cache across the Validator and Billing services to manage API key states and credit balances in real-time. The API service interacts with a persistent database (SQLite) for application data and also synchronizes credit usage and ledger entries with Redis to ensure consistency across the system. The Billing service processes usage events asynchronously, allowing for eventual consistency in credit deductions while maintaining a responsive API experience.

```mermaid
graph TD
    User[Internet / Client]
    Kong[Kong API Gateway]
    Redis[Redis Cache]
    
    subgraph "Go API Service"
        APILogic[Core API Logic]
        APIDB[SQLite]
    end

    subgraph "Go Validator Service"
        AuthLogic[Auth & Credit Validation]
    end

    subgraph "Go Billing Service"
        BillingLogic[Credit Ledger & Usage Settlement]
        BillingDB[SQLite]
    end

    User -->|1. API Request Bearer Token | Kong
    Kong -->|2a. ForwardAuth Check| AuthLogic
    AuthLogic <-->|2b. Read Auth State & Balance| Redis
    AuthLogic -.->|2c. Respond: Allow/Deny| Kong

    Kong -->|3a. Proxy Authorized Request| APILogic
    APILogic <-->|3b. Read/Write Service Data| APIDB
    APILogic -->|3c. Reconcile API Metrics| Redis

    Kong -.->|6a. Async Billing Event| BillingLogic
    BillingLogic -->|6b. Reconcile Balance| Redis
    BillingLogic <-->|6c. Read/Write Credit Ledger| BillingDB
```

## Demo Endpoints

You can test the monetization flow using the following `curl` commands. The system initializes with a demo key containing **100 credits**.

**Demo API Key:** `demo-api-key-123`

### 1. Access the Protected API
Access the API through the Kong Gateway. The gateway will automatically call the validator service:
```bash
curl -H "X-API-Key: demo-api-key-123" http://localhost:8000/service
```

### 2. Settle Billing (Deduct Credits)
Manually trigger a billing settlement to simulate usage costs:
```bash
curl -X POST -H "X-API-Key: demo-api-key-123" "http://localhost:8081/api/v1/billing?amount=10"
```

### 3. Verify Insufficient Credits
If you deduct enough credits to reach zero, the Kong gateway (via the validator) will block further requests with a `402 Payment Required` error.

`GET api/v1/info` - Returns basic information about the API and its usage.

```curl
curl -H "Authorization: Bearer demo-api-key-123" http://localhost:8998/api/v1/info
```

```json
{
  "name": "Demo Monetized API",
  "version": "0.0.1",
  "description": "A demo API. See more at https://api.domain.tld/docs"
}
```

### Billing

`GET api/v1/usage/<period>` - Returns a history of API usage and charges. Defaults to `last_24h` if no period is specified.

```curl
curl -H "Authorization: Bearer demo-api-key-123" http://localhost:8998/api/v1/usage/last_24h
```

```json
{
  "timestamp": "2024-06-01T12:00:00Z",
  "credits_used": 5,
  "status": "success",
  "balance": 95,
  "usage_history": [
    {
      "timestamp": "2024-06-01T11:00:00Z",
      "credits_used": 5,
      "status": "success"
    },
    {
      "timestamp": "2024-06-01T10:00:00Z",
      "credits_used": 10,
      "status": "success"
    }
  ]
}
```

### Health

`GET api/v1/health` - Returns the health status of the API.

```curl
curl -H "Authorization: Bearer demo-api-key-123" http://localhost:8998/api/v1/health
```

```json
{
  "status": "healthy",
  "latency_ms": 20,
  "timestamp": "2024-06-01T12:00:00Z",
  "uptime": "72h"
}
```

### Consumables

`POST api/v1/work/<level>` - Performs a unit of work, consuming credits, returns data.

The `<level>` parameter can be `easy`, `medium`, or `hard`, and determines the complexity and credit cost of the work.

```curl
curl -X POST -H "Authorization: Bearer demo-api-key-123" http://localhost:8998/api/v1/work/medium
```

```json
{
    "credits_used": 5,
    "status": "success",
    "timestamp": "2024-06-01T12:00:00Z",
    "duration_ms": 150,
    "result": "Base64-encoded data string"
}
```
