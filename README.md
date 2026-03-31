# meet-mrt

Find fair and practical MRT meetup stations between two starting points.

## Stack

- Backend: Go, Redis, Google Distance Matrix API
- Frontend: React + Vite

## Project Structure

```text
backend/
	main.go
	handlers.go
	stations.go
	redislimiter.go
	stations.json

frontend/
	src/
	package.json
```

## Prerequisites

- Go 1.25+
- Node.js 18+
- npm
- Redis (local or remote)

## Environment Variables

Create or update `backend/.env`:

```env
GOOGLE_MAPS_API_KEY=your_api_key_here
REDIS_URL=localhost:6379
REDIS_PASSWORD=
USE_MOCK=true
PORT=8080
```

Notes:

- Set `USE_MOCK=true` for local testing without paid API calls.
- Set `USE_MOCK=false` when you want real Google travel times.

## Run Locally

### 1. Start Redis

Docker option:

```bash
docker run --name meetmrt-redis -p 6379:6379 -d redis:7
```

### 2. Start Backend

```bash
cd backend
go mod tidy
go run .
```

Backend default URL:

- `http://localhost:8080`

### 3. Start Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend default URL:

- `http://localhost:5173`

The frontend is configured to call backend at `http://localhost:8080`.

## API Endpoints

### `GET /api/health`

Returns service health.

### `GET /api/remaining`

Returns remaining monthly searches from Redis rate limiter.

### `GET /api/stations`

Returns all MRT stations loaded from `backend/stations.json`.

### `POST /api/find-meetup`

Request body:

```json
{
	"stationA": "Dhoby Ghaut",
	"stationB": "Jurong East"
}
```

Response includes top candidates sorted by score (lower is better).

Each candidate includes:

- `timeFromA`
- `timeFromB`
- `minTime`
- `maxTime`
- `difference`
- `score`

## Scoring Logic

Current weighted score:

```text
score = minTime*0.5 + difference*0.3 + maxTime*0.2
```

- `minTime` has larger weight than `difference`.
- Lower score means a better meetup station.

## Development Checks

Backend compile check:

```bash
cd backend
go build ./...
```

Backend tests:

```bash
cd backend
go test ./...
```

Frontend lint:

```bash
cd frontend
npm run lint
```

Frontend production build:

```bash
cd frontend
npm run build
```

## Troubleshooting

- Frontend command typo: use `npm run dev` (not `npm rund ev`).
- If backend cannot connect to Redis, verify `REDIS_URL` and Redis is running.
- If station lookup fails, ensure names match entries in `backend/stations.json`.
