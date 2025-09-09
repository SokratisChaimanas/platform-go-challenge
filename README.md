# Favourites API (Go) — Solution Overview & Runbook

## What this service offers

A small, production‑minded HTTP API to **manage user favourites** over *assets* (charts, insights, audiences).
Users can add/remove assets to their favourites and list them with pagination; assets can have their
descriptions edited.

**Tech stack**
- Go 1.25.0 (modules)
- Router: [chi](https://github.com/go-chi/chi)
- Persistence: Postgres 16 via [ent](https://entgo.io) ORM
- OpenAPI / Swagger UI: [swaggo](https://github.com/swaggo/swag) + `http-swagger`
- Logging: std `slog`

## Architecture at a glance

The code follows a **ports & adapters (hexagonal)** split:

```text
internal/
  app/           # Application services (use-cases, orchestration)
  domain/        # Entities and business errors (User, Asset, Favourite)
  ports/         # Interfaces (repositories) consumed by app services
  adapters/
    http/chi/    # HTTP transport (handlers, router, JSON I/O)
    ent/         # Postgres persistence (ent client & repository impls)
  platform/
    config/      # Environment-driven configuration
    db/          # DB client, migrations & dev seeding
  shared/logger/ # Slog setup helpers
cmd/api/          # Program entrypoint and Swagger metadata
docs/             # Generated OpenAPI (swagger.json/yaml)
ent/              # ent schema & generated code
```

**Entities**
- `User` — ID (UUID), timestamp
- `Asset` — ID (UUID), `type` (`chart|insight|audience`), `description`, `payload` (free-form JSON), timestamp
- `Favourite` — `(user_id, asset_id)` pair + timestamp

**Key services**
- `FavouritesService` — validates user & asset, prevents duplicates, creates/removes/list favourites
- `AssetService` — edits asset descriptions with simple validation
- `UserService` — retrieves users

## Repository layout

```text
cmd/api/
  main.go      # wiring (config, db client, router, graceful shutdown)
  swg.go       # Swagger metadata (title, basePath, etc.)
internal/
  adapters/
    http/chi/handlers/  # health, user, favourites, assets handlers
    ent/                # ent-backed repository implementations
  app/                  # use-cases: user, favourites, asset
  domain/               # entities + domain errors
  ports/                # repository interfaces
  platform/
    config/             # env -> Config
    db/                 # ent client + dev seeding
  shared/logger/        # slog helpers
docs/                   # generated OpenAPI (json|yaml|go)
ent/schema/             # ent Entity schemas (user, asset, favourite)
Dockerfile, docker-compose.yml, .env
```

## How to run

### With Docker Compose (recommended)
Requirements: Docker Desktop / Engine + Compose.

```bash
# From repo root
docker compose up --build
```
- API will listen on **http://localhost:8080** (mapped from container `:8080`).
- Postgres 16 starts alongside the API (service `postgres`), using credentials from `.env`.
- On first run in **dev**, the DB is **seeded** with a few users and assets (see _Dev seed IDs_ below).

### Environment
Config comes from env vars (see `.env`):

| Var | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `dev` | Controls dev-only seeding |
| `HTTP_ADDR` | `:8080` | Listen address inside the container |
| `DB_HOST`/`DB_PORT`/`DB_NAME`/`DB_USER`/`DB_PASS` | `postgres`/`5432`/`favs`/`app`/`app` | Postgres connection |
| `DB_SSLMODE` | `disable` | Postgres SSL mode |
| `LOG_LEVEL` | `info` | slog level |

Compose additionally maps `${HTTP_PORT:-8080}:8080`, so you can override the **host** port with `HTTP_PORT=9090` etc.

## API & Swagger
- **Base URL**: `http://localhost:8080/api`
- **Swagger UI**: `http://localhost:8080/docs/` (served by `http-swagger`).
- OpenAPI is generated and also checked into `docs/swagger.json` & `docs/swagger.yaml`.

### Endpoints
- **PATCH /api/assets/{asset_id}/description** — _Edit asset description_  
  Tags: `assets`  
  Responses: 200, 400, 404, 500
  Parameters: `asset_id` (path, string, required), `payload` (body, #/definitions/handlers.AssetEditRequest, required)
- **GET /api/healthz** — _Health check_  
  Tags: `health`  
  Responses: 200
- **GET /api/users/{user_id}** — _Get user_  
  Tags: `users`  
  Responses: 200, 400, 404, 500
  Parameters: `user_id` (path, string, required)
- **GET /api/users/{user_id}/favourites** — _List favourites for a user_  
  Tags: `favourites`  
  Responses: 200, 400, 404, 500
  Parameters: `user_id` (path, string, required), `limit` (query, integer, optional), `offset` (query, integer, optional)
- **POST /api/users/{user_id}/favourites** — _Add favourite_  
  Tags: `favourites`  
  Responses: 201, 400, 404, 409, 500
  Parameters: `user_id` (path, string, required), `payload` (body, #/definitions/handlers.FavouriteAddRequest, required)
- **DELETE /api/users/{user_id}/favourites/{asset_id}** — _Remove favourite_  
  Tags: `favourites`  
  Responses: 204, 400, 404, 500
  Parameters: `user_id` (path, string, required), `asset_id` (path, string, required)

### Quick cURL examples
```bash
# Health
curl -s http://localhost:8080/api/healthz

# Get a user
curl -s http://localhost:8080/api/users/11111111-1111-1111-1111-111111111111

# List favourites (paginated)
curl -s 'http://localhost:8080/api/users/11111111-1111-1111-1111-111111111111/favourites?limit=10&offset=0'

# Add favourite
curl -s -X POST http://localhost:8080/api/users/11111111-1111-1111-1111-111111111111/favourites   -H 'Content-Type: application/json'   -d '{"asset_id":"aaaaaaa1-0000-0000-0000-000000000001"}'

# Remove favourite
curl -s -X DELETE http://localhost:8080/api/users/11111111-1111-1111-1111-111111111111/favourites/aaaaaaa1-0000-0000-0000-000000000001 -i

# Edit asset description
curl -s -X PATCH http://localhost:8080/api/assets/aaaaaaa1-0000-0000-0000-000000000001/description   -H 'Content-Type: application/json'   -d '{"description":"New description from Swagger"}'
```

## Dev seed IDs (handy for testing)
When `APP_ENV=dev` (or `SEED=1`), the DB is pre-populated with a few users and assets.

**Users**
- `11111111-1111-1111-1111-111111111111`
- `22222222-2222-2222-2222-222222222222`
- `33333333-3333-3333-3333-333333333333`

**Assets**
- `aaaaaaa1-0000-0000-0000-000000000001`
- `aaaaaaa1-0000-0000-0000-000000000002`
- `bbbbbbb2-0000-0000-0000-000000000001`
- `bbbbbbb2-0000-0000-0000-000000000002`
- `ccccccc3-0000-0000-0000-000000000001`
- `ccccccc3-0000-0000-0000-000000000002`

## Notes & decisions

- Routes are mounted under `/api`, with Swagger UI exposed at `/docs/*`.
- Standard chi middleware in use: `RequestID`, `RealIP`, `Logger`, `Recoverer`, and a request `Timeout(30s)`.
- All responses are JSON with consistent error shapes: `{ "error": "..." }`.
- Pagination for listing favourites uses `limit` (defaults to 20, max 50) and `offset`.
- Duplicate favourite inserts respond with **409 Conflict**.
- ent applies schema migrations on startup; dev seeding runs once when the DB is empty.
