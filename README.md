# Insider Assessment Project - Automatic Message Sending System

## Project Overview

This project implements an automatic message sending system in Go that:

* Retrieves unsent messages from a PostgreSQL database,
* Sends 2 messages every 2 minutes to their respective recipients via an HTTP API,
* Ensures each message is sent only once,
* Supports dynamic addition of new messages,
* Provides API endpoints to start/stop the sending process and to retrieve the list of sent messages. 
* Caches message send metadata (message ID and timestamp) in Redis for quick reference.

---

## Features

* **Database-driven message retrieval** with character limits enforced
* **Automatic scheduling** implemented in Go using time.Ticker
* **Two REST API endpoints:**
    * `POST "/scheduler"` - Start/stop the automatic message sending process
    * `GET /messages/sent` - Retrieve list of sent messages
* Swagger API documentation
* Docker-friendly setup with PostgreSQL and Redis containers
* Clean architecture and maintainable codebase

---

## Technologies Used

* Golang
* PostgreSQL (Docker)
* Redis (Docker)
* Swagger for API docs
* HTTP client for sending messages

---

## Getting Started

### Prerequisites

* Docker
* Go (1.20+ recommended)
* `goose` migration tool

---

### Setup Instructions

1. **Run PostgreSQL:**

```bash
docker run --rm -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_DB=postgres -e POSTGRES_PASSWORD=postgres -d postgres:17
```

2. **Run database migrations:**

```bash
go install github.com/pressly/goose/cmd/goose@v3.24.3
goose -dir migrations up
```

3. **Populate the database with messages:**

```bash
go run ./cmd/message-seeder -n 100
```

4. **Run Redis:**

```bash
docker run --rm -p 6379:6379 -d redis
```

5. **Start the message processor service:**

```bash
go run ./cmd/message-processor
```

6. **Start/Stop automatic sending via API:**

* Start sending:

```bash
curl -X POST http://localhost:8000/scheduler -H "Content-Type: application/json" -d '{"action": "start"}'
```

* Stop sending:

```bash
curl -X POST http://localhost:8000/scheduler -H "Content-Type: application/json" -d '{"action": "stop"}'
```

7. **Get list of sent messages:**

```bash
curl -X GET http://localhost:8000/messages/sent
```

---

## Project Structure

```
├── cmd/
│   ├── message-processor/        # Main app to process and send messages
│   ├── message-seeder/            # CLI tool to seed DB with messages
├── migrations/                   # Database migration files
├── internal/                     # Application logic (services, handlers, models)
├── docs/                        # Swagger API docs
├── README.md                    # Project documentation
└── ...
```

---

## Notes

* The system uses a custom Go scheduler with a 2-minute interval, without external cron tools.
* Sent messages are tracked to avoid duplication.
* Redis caching stores the `messageId + msisdn` as key and sending time after each successful send, as free version of webhook.site does not provide generated content as response to have unique messageId I have decided to use messageId + msisdn as key
* The message sending is simulated via a configurable webhook URL (default to webhook.site).
