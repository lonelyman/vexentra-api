# AI Handoff Plan: MinIO File Upload Platform (Vexentra)

Last updated: 2026-04-23
Owner intent: Build production-ready file upload/storage pipeline, starting with profile image upload, designed to scale to broader asset types later.

## 1) Context Summary (What was decided)

- Current state: profile image uses URL input; no first-class file upload pipeline yet.
- Target state: private object storage with controlled access, metadata in Postgres, and extensible architecture for future files.
- Storage decision: start with self-hosted MinIO (S3-compatible) for cost control; keep abstraction to switch provider later.
- Delivery strategy: incremental milestones (M1 -> M2 -> M3), avoid over-engineering in first release.

## 2) Non-Negotiable Principles

- Do not trust client as source of truth for file completeness.
- Business metadata lives in Postgres, binary file lives in object storage.
- Keep bucket private; access via short-lived signed URLs.
- Validation must be server-side (type, size, ownership, authorization).
- Schema changes only via goose SQL migrations; no AutoMigrate for schema evolution.
- APIs must be idempotent where retries are expected (`/complete`, worker jobs).

## 3) Scope by Milestone

## M1 (must ship first)

Goal: End-to-end profile image upload works in production.

Deliverables:
- Add MinIO service to `docker-compose.yml` with persistent volume.
- Add storage config/env in API.
- Add DB tables:
  - `files`
  - `upload_sessions`
- Add API endpoints:
  - `POST /api/v1/uploads/presign`
  - `POST /api/v1/uploads/complete`
  - `GET /api/v1/files/:id/url`
  - `DELETE /api/v1/files/:id`
- Update profile flow on web:
  - replace avatar URL input with upload UI
  - preview/progress/error states
  - save `profile_file_id` relation

Out of scope in M1:
- no Tus resumable upload yet
- no CDN edge/sign-cookie yet
- no heavy async processing pipeline yet

## M2 (quality hardening)

Goal: Improve reliability, cleanup, and image quality pipeline.

Deliverables:
- Worker for async processing (thumbnail/normalize/webp if needed).
- Add `processing_status` lifecycle.
- Temp object cleanup:
  - MinIO lifecycle policy for `uploads/tmp/*` (24h)
  - reconcile cron job to clean expired sessions/orphans.

## M3 (scale features)

Goal: Handle high traffic and large assets.

Deliverables:
- Optional imgproxy + cache strategy.
- Optional Tus.io for large/resumable uploads.
- Optional CDN (Cloudflare/CloudFront) with signed access strategy.

## 4) Data Model Contract (initial)

### `files` table

Required columns:
- `id` (uuid)
- `owner_type` (text)
- `owner_id` (uuid)
- `category` (text) e.g. `profile_image`, `portfolio_asset`, `document`
- `object_key` (text unique)
- `original_filename` (text)
- `mime_type` (text)
- `size_bytes` (bigint)
- `sha256` (text)
- `etag` (text nullable)
- `visibility` (text) default `private`
- `processing_status` (text) default `pending` values: `pending|ready|failed`
- `processing_error` (text nullable)
- `metadata` (jsonb default `{}`)
- `created_by` (uuid)
- `created_at`, `updated_at`, `deleted_at`

Indexes:
- `(owner_type, owner_id, category)`
- `(created_at)`
- partial index for active (`deleted_at IS NULL`)

### `upload_sessions` table

Required columns:
- `id` (uuid)
- `user_id` (uuid)
- `intent` (text) e.g. `profile_image`
- `temp_object_key` (text unique)
- `expected_mime` (text)
- `expected_max_size` (bigint)
- `status` (text) values `pending|completed|expired|cancelled`
- `expires_at` (timestamptz)
- `completed_at` (timestamptz nullable)
- `created_at`, `updated_at`, `deleted_at`

Rules:
- `/complete` can only complete `pending` + non-expired session.
- session reuse should be blocked after completion.

## 5) API Behavior Contract

### `POST /uploads/presign`
Input:
- `intent`
- `filename`
- `mime_type`
- `size_bytes`

Server responsibilities:
- validate user permission for intent
- validate allowlist mime and max size per intent
- create `upload_session`
- generate presigned PUT/POST for MinIO temp key
- return upload URL + required headers/fields + session_id

### `POST /uploads/complete`
Input:
- `upload_session_id`

Server responsibilities:
- fetch session + validate ownership + status
- HEAD object from MinIO temp key
- validate actual size/mime constraints
- copy object temp -> permanent key
- create `files` record
- mark session completed atomically
- update owner relation (for profile image flow)
- return file metadata

Idempotency:
- repeated `complete` on same completed session returns existing file result (not duplicate)

### `GET /files/:id/url`
- auth required
- verify requester has read permission
- return short-lived presigned GET URL

### `DELETE /files/:id`
- soft delete metadata + remove object (or tombstone strategy)
- owner/admin authorization required

## 6) Security Checklist

- Bucket is private (no anonymous read/list).
- Server-side MIME allowlist + magic byte sniff where applicable.
- Max size per category enforced server-side.
- Object key must never trust raw filename path.
- Presigned URL TTL short (5-15 minutes).
- Rate limit presign/complete endpoints.
- Audit log: who uploaded/deleted and when.

## 7) Performance/Cost Guardrails

- Start without CDN/imgproxy until observed need.
- Use small/medium derivative generation only after M1 stable.
- Monitor:
  - upload success rate
  - complete failure rate
  - orphan object count
  - MinIO disk usage growth

## 8) Known Decisions / Debates Resolved

- Event-driven notifications are not source-of-truth for M1.
  - Use client `/complete` as main path.
  - MinIO events may be added later for reconciliation.
- Do not rely on ETag as universal checksum.
  - store and trust own `sha256` hash policy.
- Avoid duplicate boolean `is_processed` initially.
  - use `processing_status` only.

## 9) Implementation Order (for next AI)

1. Add env + storage config structs and validation.
2. Add MinIO client wrapper behind `StorageService` interface.
3. Add migrations for `files`, `upload_sessions`, and `persons.profile_file_id` (or equivalent relation).
4. Add repository + service layer for upload sessions/files.
5. Add transport handlers/routes for 4 endpoints.
6. Add web uploader UI in workspace profile.
7. Add integration tests for presign+complete happy path and failures.
8. Add README runbook for local Docker + migration + smoke test.

## 10) Definition of Done (M1)

- User can upload profile image through web UI.
- File stored in MinIO private bucket.
- App displays image via signed URL retrieval.
- Ownership and authorization enforced.
- Invalid type/size blocked.
- Retry of `/complete` does not duplicate records.
- Docs updated and reproducible local runbook provided.

## 11) Handoff Prompt Template (copy to next AI)

"Read `docs/AI_HANDOFF_MINIO_PLAN_2026-04-23.md` first. Implement M1 only (no M2/M3 scope creep). Keep schema via goose migrations, add API + web flow for profile image upload using MinIO presign/complete. Preserve existing auth/session behavior and add tests for add/edit upload paths."

