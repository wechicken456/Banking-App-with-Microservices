# 🏗️ Event-Driven Banking Microservices - Architecture Overview (Golang + PostgreSQL)

## 📌 Architecture Overview

This project is a local-first, event-driven banking backend written in **Golang**, focusing on **scalability**, **reliability**, and **eventual consistency**. The system uses **gRPC** for inter-service communication, **Redis** for background job queueing, and **PostgreSQL** (self-hosted) for data storage. Everything is containerized via **Docker**, with future expansion to **Kubernetes**.

---

## 🧩 Microservices

| Service              | Responsibilities                                                                 |
|----------------------|-----------------------------------------------------------------------------------|
| **auth-service**     | User signup/login, token generation (JWT), password hashing, token verification   |
| **user-service**     | Account profiles, user metadata, account state management                         |
| **payment-service**  | Initiate payments, handle authorization, emit events to `transaction-service`     |
| **transaction-service** | Ensure ACID transactions and consistency, log events, rollback via outbox/saga |
| **notification-service** | Send async email alerts via Gmail SMTP using Redis queue                       |
| **api-gateway**      | Entry point for frontend; routes HTTP requests to services, rate-limits, etc.     |

---

## 🧱 Architecture Patterns

- **Communication**:  
  - gRPC (intra-service)  
  - HTTP via API Gateway (external/client-facing)

- **Asynchronous Processing**:  
  - Redis queue for background jobs (e.g. email notifications, retries)

- **Database**:  
  - PostgreSQL for all services (isolated schema per service)
  - Connection via `sqlx + `sqlc`  
  - Migrations managed with `goose`

- **Transactional Integrity**:  
  - Saga pattern for distributed transactions  
  - Outbox pattern to ensure message delivery  
  - Idempotency keys for retry-safe actions

---

## 🗂️ Project Directory Structure

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
│   ├── proto/                  # gRPC service definitions
│   ├── handler/                # HTTP/gRPC handler logic
│   ├── service/                # Business logic
│   ├── db/                     #  goose migration files, sql schema and queries, sqlc
|   |   |--- schema
|   |   |--- queries
|   |   |--- sqlc                   
│   ├── config/                 # Env, config loader
│   ├── tests/                  # Unit/integration tests
│   ├── Dockerfile
│   └── main.go
│
├── user-service/
│   └── (same structure as above)
│
├── payment-service/
│   └── (same structure as above)
│
├── transaction-service/
│   └── (same structure as above)
│
├── notification-service/
│   ├── worker/                 # Redis consumers
│   ├── mailer/                 # Email logic
│   └── (rest same as above)
│
├── proto/                      # Shared gRPC definitions
│
|── docker-compose.yml
|── .env

```


## Testing

### Postgres on Docker and Goose

Each microservice will have its own database, and each of them will run on a different docker container.

To test connections database and verify `goose` migrations:

1. `docker compose up`: to start all the containers.
2. `goose postgres "postgres://{USER}:{PASSWORD}@localhost:5432/{DATABASE_NAME}?sslmode=disable"` where each of the variables can be found in the `.env` file.
