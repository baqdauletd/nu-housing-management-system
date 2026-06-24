# NU Housing Management System

Backend service for housing applications and document storage.

Fronted service is the following repo - [Frontend](https://github.com/ajtmaganbetova/nu-housing-management-system-client)

## Running

`docker compose up --build` starts:
- `backend` on `http://localhost:8080`
- `minio` on `http://localhost:9000`
- MinIO console on `http://localhost:9001`


## Setup

1. Copy the example env file:

```bash
cp .env.example .env
```

2. Fill in the real values in `.env`. - *Required Environment Variable* section below

3. Start the services:

```bash
docker compose up --build
```

## Required Environment Variables

Create a root `.env` file based on `.env.example`.

Required:
- `POSTGRES_URL`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `JWT_SECRET`

Usually kept as-is for local Docker:
- `MINIO_BUCKET`
- `MINIO_PUBLIC_ENDPOINT`
- `MINIO_USE_SSL`
- `FRONTEND_ORIGINS`
- `GOOGLE_ALLOWED_DOMAIN`
  comma-separated list supported, for example `nu.edu.kz,alumni.nu.edu.kz`

Optional unless your auth flow needs it:
- `GOOGLE_CLIENT_ID`
- `OPENAI_API_KEY`
- `OPENAI_MODEL`

## Local Docker Notes

- Inside Docker, the backend connects to MinIO using `minio:9000`.
- For browser-visible links or local access, `MINIO_PUBLIC_ENDPOINT` defaults to `localhost:9000`.
- Documents are uploaded through the backend endpoint, so frontend users do not talk to MinIO directly.
- Current backend code is aligned to the live Supabase `documents` schema used by the team database.

## Common Commands

Start:

```bash
docker compose up --build
```

Stop:

```bash
docker compose down
```

Stop and remove MinIO data volume:

```bash
docker compose down -v
```

## Troubleshooting

If the backend fails on startup:
- check that `.env` exists in the repo root
- check that `POSTGRES_URL` is reachable from your machine
- check that `JWT_SECRET` is set
- check that the MinIO credentials in `.env` are not empty

If documents do not open:
- make sure backend is reachable at `http://localhost:8080`
- make sure MinIO is running
- make sure the uploaded file exists in the configured bucket

If Docker starts but backend cannot reach the database:
- verify the Supabase or Postgres host allows your teammate's network
- verify the URL and credentials in `POSTGRES_URL`

## Database Migration

For an existing database, apply the document-analysis migration before starting the updated backend:

```bash
psql "$POSTGRES_URL" -f infrastructure/db/2026-04-18_document_analysis_migration.sql
```
