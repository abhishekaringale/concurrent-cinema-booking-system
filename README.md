# 🎬 Concurrent Cinema Booking System

[![CI Pipeline](https://github.com/abhishekaringale/concurrent-cinema-booking-system/actions/workflows/ci.yml/badge.svg)](https://github.com/abhishekaringale/concurrent-cinema-booking-system/actions)
[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)](https://golang.org)
[![Redis](https://img.shields.io/badge/Redis-Upstash%20Cloud-DC382D?style=flat&logo=redis)](https://upstash.com)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind-v3-38B2AC?style=flat&logo=tailwind-css)](https://tailwindcss.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, full-stack cinema reservation platform engineered in **Go** and **React** to solve real-world **concurrency race conditions** and double-booking anomalies during high-traffic ticket sales.

---

## 🌐 Live Deployments

* 🚀 **Live Web Application (Frontend):** [https://concurrent-cinema-booking-system.vercel.app](https://concurrent-cinema-booking-system.vercel.app)
* ⚡ **Production REST API (Backend):** [https://concurrent-cinema-booking-system.onrender.com](https://concurrent-cinema-booking-system.onrender.com)
* 🗄️ **Distributed Lock Engine:** [Upstash Redis Cloud](https://upstash.com)

---

## ⚡ Key Engineering Highlights

* **Distributed Atomic Locking:** Employs Redis `SetNX` (Set if Not eXists) to guarantee that when thousands of users click the exact same seat simultaneously, **exactly one user succeeds** while all conflicting requests are rejected atomically.
* **Automated TTL Hold Expiration:** Holds are provisioned with an atomic 2-minute Time-To-Live (TTL). If a user abandons checkout, Redis automatically evicts the lock with zero database bloat or manual cleanup jobs.
* **Real-Time Streaming with Server-Sent Events (SSE):** Built a native Go [`sseBroker`](backend/internal/booking/handler.go) utilizing thread-safe channels (`chan []SeatStatus`) to broadcast seat state updates to all active browser screens with **<50ms latency** (completely eliminating HTTP polling).
* **Stress-Tested Concurrency:** Verified with automated Go concurrency test suites simulating **100,000 parallel goroutine booking writes** in 0.44s with zero race conditions (`go test -race`).
* **Session-Isolated Architecture:** Built-in session tracking isolating holds per browser tab using `sessionStorage` and persistent cryptographic user IDs.

---

## 🏗️ Architecture & Real-Time Flow

```mermaid
sequenceDiagram
    autonumber
    actor User1 as Browser (User 1)
    actor User2 as Browser (User 2)
    participant Go as Go Backend (Render)
    participant Broker as SSE Broker (Go Channels)
    participant Redis as Redis (Upstash Cloud)

    Note over User1, Go: 1. Real-Time Stream Connection
    User1->>Go: GET /movies/{id}/seats/stream (SSE)
    Go->>Broker: Register client channel
    User2->>Go: GET /movies/{id}/seats/stream (SSE)
    Go->>Broker: Register client channel

    Note over User1, Redis: 2. Atomic Hold Acquisition
    User1->>Go: POST /movies/{id}/seats/{seatId}/hold
    Go->>Redis: SetNX "seat:{movie}:{seat}" (TTL: 120s)
    Redis-->>Go: Lock Acquired (1)
    Go-->>User1: 200 OK (Hold details + SessionID)

    Note over Go, User2: 3. Instant Multi-Client Broadcast
    Go->>Redis: ListBookings()
    Redis-->>Go: Current active statuses
    Go->>Broker: Broadcast(seatStatuses)
    Broker->>User1: Push SSE update (data: JSON)
    Broker->>User2: Push SSE update (data: JSON)
    Note over User2: Seat A4 instantly turns ORANGE in <50ms!
```

---

## 🛠️ Project Structure & Tech Stack

```text
concurrent-cinema-booking-system/
├── backend/
│   ├── cmd/
│   │   └── main.go                 # Server entrypoint, CORS, and routing
│   ├── internal/
│   │   └── booking/
│   │       ├── domain.go           # Domain entities (Booking, SeatStatus) & store interfaces
│   │       ├── handler.go          # HTTP Handlers & Real-Time SSE Broker
│   │       ├── service.go          # Core booking business logic
│   │       ├── redis_store.go      # Distributed Redis store (SetNX & TxPipeline)
│   │       ├── memory_store.go     # In-memory thread-safe store (RWMutex)
│   │       └── service_test.go     # Concurrency stress tests (100k goroutines)
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── App.jsx                 # Seat grid UI, EventSource SSE listener & timers
│   │   └── index.css               # Tailwind CSS styles
│   ├── vercel.json                 # Production API rewrite proxy
│   └── package.json
├── .github/
│   └── workflows/
│       └── ci.yml                  # Automated GitHub Actions CI pipeline
└── README.md
```

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Backend** | **[Go (Golang 1.22)](backend/cmd/main.go)** | Clean layered architecture (Handler $\rightarrow$ Service $\rightarrow$ Store) with native routing. |
| **Distributed Lock**| **[Redis (Upstash)](backend/internal/booking/redis_store.go)** | Atomic `SetNX` locking, 2-minute key expiration TTL, and `TxPipeline` transactions. |
| **Real-Time** | **[Server-Sent Events (SSE)](backend/internal/booking/handler.go)** | Low-overhead HTTP streaming using native Go channels and `http.Flusher`. |
| **Frontend** | **[React 19 + Vite](frontend/src/App.jsx)** | Single-page reactive interface with instant optimistic state synchronization. |
| **Styling** | **[Tailwind CSS v3](frontend/src/index.css)** | Responsive dark-mode cinema layout with visual seat maps and screen glows. |
| **CI/CD** | **[GitHub Actions](.github/workflows/ci.yml)** | Automated parallel testing (`go test -v`) and production bundle verification. |
| **Hosting** | **Render + Vercel** | Automated Git webhook deployment with environment secret injection. |

---

## 📡 REST API Documentation

| Method | Endpoint | Description | Status Code |
| :--- | :--- | :--- | :--- |
| `GET` | `/movies` | List all available movie screenings | `200 OK` |
| `GET` | `/movies/{movieID}/seats` | Fetch current seat statuses (Available, Held, Confirmed) | `200 OK` |
| `GET` | `/movies/{movieID}/seats/stream` | Establish persistent SSE real-time stream | `200 OK (Stream)` |
| `POST` | `/movies/{movieID}/seats/{seatID}/hold` | Temporarily hold an available seat for 2 minutes | `200 OK` / `409 Conflict` |
| `PUT` | `/sessions/{sessionID}/confirm?movie_id=...` | Permanently confirm a held seat reservation | `204 No Content` |
| `DELETE` | `/sessions/{sessionID}?movie_id=...` | Immediately release a held seat reservation | `204 No Content` |

---

## 🚀 Local Development Setup

### Prerequisites
* **Go** (v1.22 or higher)
* **Node.js** (v20 or higher)
* **Redis** (Local `redis-server` or free [Upstash](https://upstash.com) cloud account)

### 1. Clone the Repository
```bash
git clone https://github.com/abhishekaringale/concurrent-cinema-booking-system.git
cd concurrent-cinema-booking-system
```

### 2. Configure Environment Variables
Create a `.env` file in the `backend/` directory:
```env
# backend/.env
REDIS_URL=rediss://default:YOUR_PASSWORD@YOUR_ENDPOINT.upstash.io:6379
PORT=8080
```

### 3. Start the Go Backend
```bash
cd backend
go run cmd/main.go
# Server running on http://localhost:8080
```

### 4. Start the React Frontend
In a new terminal window:
```bash
cd frontend
npm install
npm run dev
# Frontend running on http://localhost:5173
```

---

## 🧪 Running Concurrency & Unit Tests

Run the full test suite with Go's **Race Detector** enabled:

```bash
cd backend
go test -race -v ./...
```

### Test Benchmark Output:
```text
=== RUN   TestConcurrentBooking_ExactlyOneWins
--- PASS: TestConcurrentBooking_ExactlyOneWins (0.44s)
PASS
ok      github.com/abhishekaringale/concurrent-cinema-booking/internal/booking  1.427s
```

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
