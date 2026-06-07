# System Architecture

GeoBeat is built on a decoupled Client-Server model using a RESTful API. The backend is written in Go (Golang) and strictly follows **Clean Architecture** (Layered Architecture) principles to ensure separation of concerns, testability, and maintainability.

## 🏗️ Layered Backend Structure

The backend is divided into three primary layers. Dependencies only point inwards (Transport -> Service -> Repository).

### 1. Transport Layer (Handlers & HTTP)
* **Responsibility:** Handles all incoming HTTP requests, extracts parameters/payloads, and formats JSON responses.
* **Security:** Enforces middleware (CORS, JWT validation, OAuth state checking).
* **Rule:** This layer contains **zero business logic**. It only knows how to speak HTTP (returning 200 OK, 400 Bad Request, etc.).

### 2. Service Layer (Domain Logic)
* **Responsibility:** The "brain" of the application. It processes game rules, calculates scores, validates user answers, and orchestrates the daily challenges.
* **Rule:** This layer does not know about HTTP or SQL. It receives Go structs from the Handlers, applies logic, and calls the Repositories.

### 3. Repository Layer (Persistence)
* **Responsibility:** Direct interaction with the PostgreSQL database.
* **Implementation:** We strictly use **`pgxpool`** (the native Go driver for PostgreSQL) instead of heavy ORMs (like GORM). 
* **Rule:** Executes raw SQL queries and maps the rows back into Go structs.

---

## 📐 Design Decisions & Tech Stack

### Why Go (Golang)?
Go was selected for its strong typing, compiled performance, and native concurrency model. Goroutines allow us to handle background tasks (like updating leaderboards or pre-fetching Last.fm data) without blocking the main HTTP server thread.

### Why pgxpool over an ORM?
ORMs often introduce hidden performance bottlenecks (N+1 query problems) and abstract away the database too much. By using `pgxpool`, we maintain absolute control over connection pooling, query execution plans, and memory allocation under heavy concurrent load.

### Frontend Architecture (React/Vite)
The frontend is a Single Page Application (SPA). State management is kept local to components where possible. Map interactions utilize `@vis.gl/react-maplibre` with derived state patterns to avoid asynchronous rendering bugs (ensuring the map updates instantly when new geographical coordinates are received).
