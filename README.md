# Chirpy

API REST tipo Twitter, construída em Go puro (`net/http`, sem frameworks) como parte do curso *Learn HTTP Servers in Go* do boot.dev. Persistência em PostgreSQL via `sqlc`, migrações com `goose`, autenticação JWT + refresh tokens, webhooks de terceiros.

## Stack

- Go 1.26
- PostgreSQL (`lib/pq`)
- `sqlc` para queries type-safe
- `goose` para migrações
- `golang-jwt/jwt/v5` para JWT
- `argon2id` para hashing de passwords
- `joho/godotenv` para variáveis de ambiente

## Setup

1. Cria a base de dados:
   ```bash
   createdb chirpy
   ```

2. Cria `.env` na raiz do projecto:
   ```env
   DB_URL=postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable
   PLATFORM=dev
   JWT=<secret-aleatorio-para-assinar-jwt>
   POLKA_KEY=<chave-do-webhook-polka>
   ```

3. Corre as migrações:
   ```bash
   cd sql/schema
   goose postgres "$DB_URL" up
   ```

4. Gera o código `sqlc` (se alterares queries):
   ```bash
   sqlc generate
   ```

5. Corre o servidor:
   ```bash
   go run .
   ```

Servidor disponível em `http://localhost:8080`.

## Endpoints

### Sistema

| Método | Rota | Descrição |
|---|---|---|
| GET | `/api/healthz` | Health check |
| GET | `/admin/metrics` | Número de hits ao fileserver |
| POST | `/admin/reset` | Apaga todos os users (só em `PLATFORM=dev`) |

### Users

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/users` | — | Cria utilizador |
| PUT | `/api/users` | Bearer JWT | Actualiza email/password |
| POST | `/api/login` | — | Login, devolve access token + refresh token |
| POST | `/api/refresh` | Bearer refresh token | Gera novo access token |
| POST | `/api/revoke` | Bearer refresh token | Revoga refresh token |

### Chirps

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/chirps` | Bearer JWT | Cria chirp (máx. 140 caracteres) |
| GET | `/api/chirps` | — | Lista chirps |
| GET | `/api/chirps/{chirpID}` | — | Obtém um chirp |
| DELETE | `/api/chirps/{chirpID}` | Bearer JWT | Apaga chirp (só o autor) |

#### Query params em `GET /api/chirps`

- `author_id` (UUID, opcional) — filtra chirps por autor
- `sort` (`asc` | `desc`, opcional, default `asc`) — ordena por `created_at`

Exemplos:
```
GET /api/chirps
GET /api/chirps?sort=desc
GET /api/chirps?author_id=<uuid>&sort=desc
```

### Webhooks

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/api/polka/webhooks` | ApiKey header | Activa Chirpy Red para um user (`event: "user.upgraded"`) |

## Autenticação

- Access token: JWT assinado com HS256, expira em 1h, `Subject` = user ID.
- Refresh token: string hex de 32 bytes, válido 60 dias, guardado em `refresh_tokens`.
- Header esperado: `Authorization: Bearer <token>` (ou `Authorization: ApiKey <key>` para o webhook Polka).

## Testes

```bash
go test ./...
./teste.sh
```