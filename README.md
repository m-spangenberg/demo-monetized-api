# demo-monetized-api

This is a system design project showing a stripped-down monetized API running on my personal infrastructure. I've mocked up the API using FastAPI and used Go to mock a validator and billing service. This demo spins up a self-contained environment using Docker Compose that stacks Kong as a as API Gateway, the two Go services, Redis, and then SQLite (instead of PostgreSQL).

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
