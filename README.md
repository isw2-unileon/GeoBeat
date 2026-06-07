# Description

GeoBeat is a web application based on a guessing game about the most prominent music genres in different countries around the world. The objective of the game is to win the daily challenge of the selected game mode. The user will be able to view a full globe of the planet where the countries involved in the game will be marked:

Daily Mode - A daily challenge where you have to guess the predominant genre in a predefined country, with a total of 5 attempts and one clue for each failed attempt.
Time Trial Mode - A daily challenge where the objective is to guess the predominant genre of a group of countries and do it faster than the rest of the players. In this mode, there are no clues.
Inverse Mode - A daily challenge that involves choosing the country to which a predefined song belongs.

## Prerequisites

Ensure the following languages, tools, and minimal versions are installed and configured on your local system before attempting to build or run the project. Failure to meet these versions will result in compilation or runtime errors.

- Go (Golang): v1.25 or higher. Required for compiling the backend server, running the pgxpool drivers, and executing the core test suite.

- Node.js & npm: v22.0.x or higher. Required to install dependencies, serve the frontend SPA (Vite/React), and execute End-to-End (E2E) tests via Playwright.

- Docker & Docker Compose: Required to provision the isolated local PostgreSQL database environment for development and testing without polluting the host machine.

- GNU Make: Required to execute the automation scripts and operational targets defined in the Makefile (e.g., make run, make db-up).

- Last.fm API Account: A registered developer account is strictly necessary to obtain the LASTFM_API_KEY and LASTFM_API_SECRET required for the music metadata ingestion engine.

- Free Google Account: To access the Google Cloud Console and create OAuth credentials.

## 🚀 Cloning, Configuration, and Local Execution

Follow these steps to set up the complete development environment (Backend, Frontend, and Database) on your local machine.

### 1. Clone the repository
Download the source code and navigate to the project directory:
```bash
git clone <REPOSITORY_URL>
cd Geobeat
```

### 2. Install dependencies
The project includes an automated command to install all Go tools (like `air` for hot-reload and `golangci-lint`) as well as the Node.js packages for both the frontend and the E2E tests.
```bash
make install
```

### 3. Configure environment variables (`.env`)

This project requires two separate environment files: one for the backend server and one for the Vite frontend.

**Backend Configuration (`./backend/.env`)**
Create a `.env` file inside the `backend` directory. This file handles your database connection, security secrets, OAuth, and external APIs.

```env
# Server Configuration
SERVER_PORT=8080
ENV=development

# Database Connection (Local Docker)
DATABASE_URL=postgresql://postgres:supersecret@localhost:5432/geobeat_local?sslmode=disable

# Security & Authentication
JWT_TOKEN=your_jwt_token
GOOGLE_CLIENT_ID=your_google_oauth_client_id
GOOGLE_SECRET=your_google_oauth_secret
REDIRECT_URL=http://localhost:8080/api/auth/login/callback/

# CORS & Routing
CORS_ALLOW_ORIGIN=http://localhost:5173
FRONTEND_URL=http://localhost:5173

# Last.fm API (Mandatory for game mechanics)
LASTFM_API_KEY=your_lastfm_api_key
```

**Frontend Configuration (`./frontend/.env`)**
Create a `.env` file inside the `frontend` directory. Vite requires variables to be prefixed with `VITE_` to securely expose them to the client-side code.

```env
# Backend API Endpoints
VITE_BACKEND_URL=http://localhost:8080
VITE_GOOGLE_LOGIN=http://localhost:8080/api/auth/login/google
```

### 4. Start and prepare the Local Database
We will use Docker to spin up an isolated PostgreSQL instance for development (with data persistence).

```bash
# Start the local database container in the background
make db-local-up

# Apply migrations to create the database tables
make migrate-up

# IMPORTANT: Run the seeder to populate the music genres table
# (Replace this path with the actual location of your seeder if different)
go run backend/cmd/seeder/main.go 
```

### 5. Run the Application
With the database ready and seeded, you can start the development servers. For the best workflow, we recommend opening two separate terminals:

**Terminal 1 (Backend):**
Start the Go server with hot-reload support.
```bash
make run-backend
```

**Terminal 2 (Frontend):**
Start the Vite/React development server.
```bash
make run-frontend
```

The User Interface will be available at [http://localhost:5173](http://localhost:5173) and the backend API will respond on the configured port (usually `:8080`).

## 🧪 Running Tests

GeoBeat uses a comprehensive testing strategy covering the backend, frontend, and end-to-end (E2E) workflows. All test execution is streamlined through the `Makefile`.

### 1. Prepare the Test Database
Backend integration tests require an isolated database to prevent overwriting your local development data. The test database runs entirely in RAM (`tmpfs`) for maximum speed and automatic cleanup.
```bash
# Start the ephemeral test database in the background
make db-test-up
```

### 2. Unit and Integration Tests
Run the core test suites for both the Go backend and the React frontend. The backend tests will automatically connect to the test database you just started.
```bash
make test
```
*Note: Once you are finished running tests, you can shut down and destroy the test database instance by running `make db-test-down`.*

### 3. Code Quality & Linters
Before submitting a Pull Request, ensure your code complies with the project's strict styling rules and cognitive complexity limits (`gocognit`).
```bash
make lint
```

### 4. End-to-End (E2E) Tests
E2E tests use **Playwright** to simulate real user interactions across the application. 

**⚠️ Important:** Playwright requires the application to be live. You must have both the backend (`make run-backend`) and frontend (`make run-frontend`) servers running in separate terminals before executing this command.
```bash
make e2e
```

## 🤝 How to Contribute

We welcome community contributions! To maintain code quality and production stability, please follow our strict development workflow.

### 1. Branching Strategy
All new features, bug fixes, or documentation changes must be developed in an isolated branch created from `main`. Use the following naming conventions:
* `feature/<feature-name>` (e.g., `feature/time-trial-scoring`)
* `bugfix/<bug-name>` (e.g., `bugfix/duration-calc-error`)
* `hotfix/<critical-issue>` (For urgent production fixes)
* `docs/<topic>` (For documentation updates)

### 2. Commit Conventions
We enforce [Conventional Commits](https://www.conventionalcommits.org/). Your commit messages must be structured and descriptive so the project history remains readable:
* `feat: ...` for a new feature.
* `fix: ...` for a bug fix.
* `docs: ...` for documentation changes.
* `refactor: ...` for code changes that neither fix a bug nor add a feature.
* `test: ...` for adding missing tests or correcting existing ones.

*Example:* `feat: add redis caching layer for daily challenges`

### 3. Pull Request (PR) Process
Before your code can be merged into the `main` branch, it must pass our quality gates:

1. **Local Checks:** Ensure your branch compiles successfully and passes all local tests and linters before pushing.
   ```bash
   make lint
   make test
   ```
2. **Open a PR:** Open a Pull Request targeting the `main` branch. Provide a clear description of the changes and link any related issue trackers.
3. **Automated CI/CD Pipeline:** GitHub Actions will automatically run our CI pipeline. **Your PR will be blocked** if:
   * The code fails to compile.
   * Any unit, integration, or E2E tests fail.
   * The linter detects style violations or if any Go function exceeds the cognitive complexity threshold of 20 (`gocognit`).
4. **Peer Review:** At least one core team member must review and approve your PR before it can be merged.

## 📖 Technical Documentation

For a deep dive into the system's design, architectural patterns, and data flow, please refer to our dedicated documentation directory. 

* [📚 Read the Technical Documentation Index](./docs/README.md)

Inside the `/docs` folder, you will find detailed explanations regarding:
* **System Architecture:** Clean Architecture layers and dependency injection.
* **Database Schema:** PostgreSQL relational models, UUID strategies, and `pgxpool` configuration.
* **API Specification:** RESTful JSON contracts and endpoint security.
* **Core Algorithms:** The mathematical randomization engine and state management for daily challenges.
