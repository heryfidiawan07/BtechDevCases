# Outpost

User registration, JWT login, and wallet transfers. Backend is Go (Fiber) with PostgreSQL. Frontend is Next.js.

## Requirements

- Docker Engine and Docker Compose v2

For local development without Docker: Go 1.23+, Node.js 20+, and PostgreSQL 16+.

## Run with Docker

From the repository root:

```bash
cp .env.example .env
docker compose up --build
```

| Service | URL |
| --- | --- |
| App | http://localhost:3000 |
| API | http://localhost:8080 |
| Health | http://localhost:8080/api/v1/health |
| Postgres | `127.0.0.1:5433` |

Postgres is published on **5433** so it does not conflict with a local PostgreSQL on 5432. Containers still talk to Postgres on internal port 5432.

Use `sudo docker compose up --build` if your user is not in the `docker` group.

`docker compose up --build` rebuilds images and keeps the Postgres volume. User balances and transfer history stay.

Stop the stack with `docker compose down`. Add `-v` to delete the database volume. That does **not** clear the browser outbox (IndexedDB). Sign out, or hard-refresh after a code change that bumps the outbox version.

### Demo accounts

| Email | Password | Balance |
| --- | --- | --- |
| fidiawan07@gmail.com | 12345678 | Rp 100.000,00 |
| heryfidiawan07@gmail.com | 12345678 | Rp 100.000,00 |

New users also start with Rp 100.000,00. If a demo account already exists with a lower balance, seed raises it to this amount on API startup. Idle sessions expire after 15 minutes.

### Money and offline transfers

The UI uses Indonesian rupiah formatting: `.` thousands, `,` decimals, always two decimal places (`Rp 100.000,00`). The amount field accepts digits and an optional comma only (`50000` or `50,00`). There is no form draft in `sessionStorage`.

The **Outbox** keeps only transfers that failed because the network dropped. They retry when the browser fires `online`. Business errors (recipient not found, insufficient funds, self-transfer) appear on the form and are not stored. Sign out clears the access token and the IndexedDB outbox.

## Environment files

There are three env files, each for a different process:

| File | Used by | When |
| --- | --- | --- |
| `.env` | Docker Compose | `docker compose up` |
| `backend/.env` | Go API | `go run` without Compose |
| `frontend/.env.local` | Next.js | `npm run dev` without Compose |

Docker only needs the root `.env`. The backend and frontend files are for running those apps on the host.

Do not commit `.env` or `.env.local`. Copy the matching `*.example` file.

### Compose (`.env`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `POSTGRES_USER` | no | `wallet` | Database user |
| `POSTGRES_PASSWORD` | no | `wallet` | Database password |
| `POSTGRES_DB` | no | `wallet` | Database name |
| `JWT_SECRET` | no | see `.env.example` | Minimum 32 characters |
| `JWT_TTL` | no | `15m` | Access token lifetime |
| `INITIAL_BALANCE` | no | `100000` | Starting balance for new users |
| `CORS_ORIGIN` | no | `http://localhost:3000` | Allowed frontend origin |
| `NEXT_PUBLIC_API_URL` | no | `http://localhost:8080` | API URL baked into the frontend image |

### Backend (`backend/.env`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DATABASE_URL` | yes | | PostgreSQL connection string |
| `JWT_SECRET` | yes | | Minimum 32 characters |
| `JWT_TTL` | no | `15m` | Access token lifetime |
| `INITIAL_BALANCE` | no | `100000` | Starting balance for new users |
| `HTTP_PORT` | no | `8080` | API port |
| `CORS_ORIGIN` | no | `http://localhost:3000` | Allowed origin |
| `APP_ENV` | no | `development` | `development` or `production` |

If the API runs on the host and Postgres runs in Compose, use port **5433**:

```
DATABASE_URL=postgres://wallet:wallet@127.0.0.1:5433/wallet?sslmode=disable
```

### Frontend (`frontend/.env.local`)

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `NEXT_PUBLIC_API_URL` | no | `http://localhost:8080` | API base URL |

`NEXT_PUBLIC_*` values are public. Do not put secrets there.

## Run without Docker

Start PostgreSQL, create a database named `wallet`, then from the repository root:

```bash
cp backend/.env.example backend/.env
cd backend
go run ./cmd/api
```

```bash
cp frontend/.env.example frontend/.env.local
cd frontend
npm install
npm run dev
```

The app is at http://localhost:3000 and the API at http://localhost:8080.

## API

| Method | Path | Auth |
| --- | --- | --- |
| `GET` | `/api/v1/health` | no |
| `POST` | `/api/v1/auth/register` | no |
| `POST` | `/api/v1/auth/login` | no |
| `POST` | `/api/v1/auth/refresh` | yes |
| `POST` | `/api/v1/auth/logout` | yes |
| `GET` | `/api/v1/me` | yes |
| `GET` | `/api/v1/wallet` | yes |
| `POST` | `/api/v1/wallet/transfers` | yes |

Login returns a JWT (`sub` = user id, `email`). Transfers require an `Idempotency-Key` header so a retry never double-debits.

Amount may be `"25.50"`, `"25,50"`, or `"25"`.

```json
{
  "recipient": "heryfidiawan07@gmail.com",
  "amount": "25,50",
  "notes": "lunch"
}
```

## Tests

```bash
cd backend
go test ./...
```

The integration test in `internal/server/flow_test.go` runs only when `DATABASE_URL` is set.
