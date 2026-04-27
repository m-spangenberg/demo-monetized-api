# demo-monetized-api

This is a system design project showing a stripped-down monetized API running on my personal infrastructure. I've mocked up the API using FastAPI and used Go to mock a validator and billing service. This demo spins up a self-contained environment using Docker Compose that stacks Kong as the API Gateway, the two Go services, Redis, and then SQLite (instead of PostgreSQL).

```mermaid
sequenceDiagram
    participant User as Internet / Client
    participant CF as Cloudflare
    participant Kong as Kong (API Gateway)
    participant Validator as Go Validator Service
    participant Billing as Go Billing Service
    participant Redis as Redis (Auth/Credits)
    participant DB as Usage Database
    participant API as api.domain.tld

    User->>CF: Request with Bearer Token
    CF->>Kong: Forward through CF Tunnel

    Kong->>Validator: ForwardAuth (Check Headers / API Key)

    Validator->>Redis: Get Key Status & Credit Balance
    Redis-->>Validator: Balance > 0

    alt Credits OK

        Validator->>Redis: Reserve / Deduct Credits
        Validator-->>Kong: 200 OK + Consumer Context

        Kong->>API: Proxy Request

        alt API Error
            API-->>Kong: 4** or 5** Error
            Kong->>Billing: Send Failed Usage Event
            Billing->>DB: Record Attempt / Reconcile Charges
            Kong-->>User: Forward Error
        else API Success
            API-->>Kong: 200 OK (Data)
            Kong->>Billing: Send Usage / Settlement Event
            Billing->>DB: Persist Usage Ledger Entry
            Kong-->>User: Final Response
        end

    else Insufficient Credits or Invalid Key

        Validator-->>Kong: 402 Payment Required / 401 Unauthorized
        Kong-->>User: Error Response

    end
```

## Demo Endpoints

### Info

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

## Demo Requests

Here's a few example commands you can `curl` to test the system.
The system starts off with 100 credits for the demo API key:

Demo API Key: `demo-api-key-123`
Demo API Endpoint: `localhost:8998/api`

```bash
# Valid request with sufficient credits
curl -H "Authorization: Bearer demo-api-key-123" http://localhost:8998/api/v1/work
```