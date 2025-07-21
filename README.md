# Banking App with Microservices - Architecture Overview (Golang + PostgreSQL)

## Architecture Overview

Banking App written in **Golang** to learn about implementation of microservices. The system uses **gRPC** for inter-service communication, **Redis** for background job queueing, and **PostgreSQL** (self-hosted) for data storage. Everything is containerized via **Docker**, with future expansion to **Kubernetes**.

**Current progress**: 

1. Finished setting up PostgreSQL on **docker**.
2. Finished most features for account and auth microservices (including JWT auth). 
3. Finished gRPC for the `auth` microservice. Setting up gRPC for the `account` microservice.


Check [JOURNAL.md](./JOURNAL.md) for a more detailed 

---
## Microservices

| Service                   | Responsibilities                                                                 |
|---------------------------|-----------------------------------------------------------------------------------|
| **auth-service**          | User signup/login, JWT issuance, password hashing                 |
| **account-service**       | Manages account details, balances, deposits, withdrawals, and transaction history (ACID guarantees) |
| **transfer-service**      | Orchestrates fund transfers: validates source/target accounts, invokes debit/credit via gRPC, ensures consistency |
| **notification-service**  | Sends asynchronous email alerts (e.g. transfer success/failure) using Redis + Gmail SMTP |
| **api-gateway**           | Entry point for client requests; JWT validation, routes HTTP to internal services, handles rate limiting, auth forwarding |

---

## Architecture Patterns

- **Communication**:  
  - gRPC (internal service-to-service)  
  - HTTP (client-facing via API Gateway)

- **Asynchronous Processing** (TODO):  
  - Redis queue for background tasks like email alerts and retries

- **Database**:  
  - PostgreSQL (isolated schema per service)  
  - Accessed via `sqlx` and `sqlc`  
  - Migrations managed with `goose`

- **Transactional Integrity**:  
  - Core services maintain ACID within their DBs  
  - Saga pattern used for distributed operations (e.g., fund transfers)  
  - Outbox pattern ensures eventual message delivery  
  - Idempotency keys prevent duplicate processing

---

![Flowchart](https://www.mermaidchart.com/raw/48a2029d-139d-4572-b015-3b6bcbcac784?theme=light&version=v0.1&format=svg)

---

## Project Directory Structure

```text
banking-microservices/
│
├── api-gateway/
│   ├── main.go
│   ├── router/
│   ├── handlers/
│   └── Dockerfile
│
├── auth-service/
│   ├── proto/
│   ├── handler/
│   ├── service/
│   ├── db/
│   │   ├── schema/
│   │   ├── queries/
│   │   └── sqlc/
│   ├── config/
│   ├── tests/
│   └── Dockerfile
│
├── user-service/
│   └── (same structure as above)
│
├── account-service/
│   └── (same structure; handles both accounts and transactions)
│
├── transfer-service/
│   └── (same structure; orchestrates debit/credit with gRPC calls)
│
├── notification-service/
│   ├── worker/       # Redis consumers
│   ├── mailer/       # Email logic
│   └── (rest same as above)
│
├── proto/            # Shared gRPC definitions
│
├── docker-compose.yml
├── .env



## Testing

### Postgres on Docker and Goose

Each microservice will have its own database, and each of them will run on a different docker container.

To test connections database and verify `goose` migrations:

1. `docker compose up`: to start all the containers.
2. `goose postgres "postgres://{USER}:{PASSWORD}@localhost:5432/{DATABASE_NAME}?sslmode=disable"` where each of the variables can be found in the `.env` file.
3. `docker compose down -v`: stop all containers and remove all volumes.

OR using the `Makefile`:
1. `docker compose up`: to start all the containers.
2. `make goose-up`: apply all up migrations.
3. `make goose-down`: apply all down migrations.
4. `docker compose down -v`: stop all containers and remove all volumes.
