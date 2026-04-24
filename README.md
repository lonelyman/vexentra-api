# Vexentra API

Go REST API built with Clean Architecture, designed for production from day one.

---

## Tech Stack

| Layer      | Technology                                         |
| ---------- | -------------------------------------------------- |
| Framework  | Fiber v3                                           |
| Database   | PostgreSQL 18 + GORM (query only)                  |
| Migrations | goose (SQL, timestamp-versioned)                   |
| Cache      | Redis                                              |
| Auth       | JWT (HS256) — Access + Refresh Token               |
| Logger     | slog (JSON production / PrettyHandler development) |
| Validation | go-playground/validator v10                        |
| Container  | Docker + Docker Compose                            |
| **Time**   | **`pkg/wela` — Bangkok timezone + พ.ศ. utility**   |

> **Schema is managed by SQL migrations only.** GORM `AutoMigrate` has been removed —
> structural changes must go through `database/migrations/*.sql` (goose). GORM is used
> strictly as a query builder / ORM.

---

## Coding Conventions

### Time & Date — ใช้ `pkg/wela` เสมอ

**ห้ามใช้ `time.Now()` โดยตรง** ให้ใช้ `wela` แทนทุกกรณี:

```go
import "vexentra-api/pkg/wela"

// ✅ ถูกต้อง
now := wela.NowUTC()                    // เก็บลง DB
expiresAt := wela.NowUTC().Add(1 * time.Hour)
wela.NowUTC().After(someTime)

// ❌ ห้ามใช้
now := time.Now()
now := time.Now().UTC()
```

**ห้ามใช้ layout string โดยตรง** ให้ใช้ฟังก์ชันของ `wela` แทน:

```go
// ✅ ถูกต้อง
wela.FormatRFC3339(t)          // แทน t.Format("2006-01-02T15:04:05Z07:00")
wela.FormatISODate(t)          // แทน t.Format("2006-01-02")
wela.ParseRFC3339Any(raw)      // แทน time.Parse(time.RFC3339, raw)

// ❌ ห้ามใช้
t.Format("2006-01-02")
t.Format("2006-01-02T15:04:05Z07:00")
time.Parse(time.RFC3339, raw)
```

ดู API ทั้งหมดได้ที่ [`pkg/wela/README.md`](pkg/wela/README.md)

---

## Quick Start

```bash
git clone git@github.com:lonelyman/vexentra-api.git && cd vexentra-api
docker compose up --build
# API available at http://localhost:8000
```

---

## Environment Variables

```env
# App
API_ENV=development
API_PORT=3000
API_TIMEZONE=Asia/Bangkok
API_CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
APP_SHOWCASE_PERSON_ID=          # optional: Person UUID for public showcase endpoint

# PostgreSQL
POSTGRES_PRIMARY_HOST=vexentra-pgsql
POSTGRES_PRIMARY_PORT=5432
POSTGRES_PRIMARY_USER=vexentra_user
POSTGRES_PRIMARY_PASSWORD=vexentra_pass
POSTGRES_PRIMARY_NAME=vexentra_db
POSTGRES_PRIMARY_SSL_MODE=disable

# Redis
REDIS_HOST=vexentra-redis
REDIS_PORT=6379
REDIS_PASSWORD=vexentra_redis_pass
REDIS_DB=0

# JWT
JWT_ACCESS_SECRET=change-me-in-production
JWT_ACCESS_EXPIRY=60m
JWT_REFRESH_SECRET=change-me-refresh-in-production
JWT_REFRESH_EXPIRY=336h
JWT_ISSUER=vexentra-api
```

---

## Project Structure

```
cmd/api/
database/
  migrations/                   # goose SQL migrations (source of truth for schema)
  backups/                      # pg_dump snapshots (pre-migration safety)
internal/
  bootstrap/                    # DI wiring (no AutoMigrate — goose-only)
  config/                       # Env loader (Fail-Fast)
  modules/
    person/                     # Person entity + repo interface  ← identity หลัก
    user/                       # User + Profile entities + repo interfaces
      usersvc/                  # UserService + ProfileService
    socialplatform/             # SocialPlatform master data
      platformsvc/
  adapters/database/postgres/
    pgperson/                   # persons table
    pguser/                     # users, profiles, skills, experiences, portfolio, social_links
    pgsocialplatform/           # social_platforms table
  transport/http/
    auth/                       # Auth handlers
    user/                       # User + Profile handlers
    socialplatform/             # Social platform handlers
    health/                     # /health/* probes
    middlewares/                # AuthMiddleware, StructuredLogger
    presenter/                  # RenderItem / RenderList / RenderError
pkg/
  auth/                         # JWT AuthService, Claims (AccessClaims + RefreshClaims)
  custom_errors/
  logger/
```

---

## Database Migrations

Schema lives under `database/migrations/` and is applied with
[goose](https://github.com/pressly/goose). Filenames use timestamp format
`YYYYMMDDHHMMSS_name.sql` (goose rejects version `0`).

Applied migrations (as of 2026-04-23):

| Version        | Name               | Purpose                                                                                                     |
| -------------- | ------------------ | ----------------------------------------------------------------------------------------------------------- |
| 20260422000001 | baseline           | Snapshot of GORM-managed schema (idempotent `CREATE IF NOT EXISTS`)                                         |
| 20260422000002 | schema_hardening   | deleted_at on all tables, case-insensitive unique indexes, FK fixes, phantom-person backfill                |
| 20260422000003 | project_management | projects, project_members, transaction_categories, project_transactions (+ enums, seeds, CHECK constraints) |
| 20260422000004 | persons_creator_nullable | relax FK on `persons.created_by_user_id` for safer bootstrap compatibility                            |
| 20260422000005 | tasks              | per-project tasks (`todo/in_progress/done/cancelled`) with assignee/due_date                                |
| 20260423000006 | project_status_master | master table `project_statuses` for status dropdown/dashboard logic                                      |
| 20260423000007 | project_financial_plan | contract amount + retention + installment schedule tables                                                |
| 20260423000008 | add_penalty_expense_category | add expense category `penalty` (`ค่าปรับ`) to `transaction_categories`                            |

**Workflow:**

```bash
# status
goose -dir database/migrations postgres "$DSN" status

# apply
goose -dir database/migrations postgres "$DSN" up

# new migration
goose -dir database/migrations create <name> sql
```

หรือใช้ Makefile (แนะนำสำหรับทีม):

```bash
# run from vexentra-api/
make migrate-status
make migrate-up
make migrate-down
make migrate-version
make migrate-create name=add_project_indexes
```

Always back up before applying to non-empty databases — see `database/backups/`.

---

## Data Model (สำคัญ)

ระบบแยก **Person** ออกจาก **User** เพื่อรองรับการเพิ่มบุคคลในโครงการโดยที่เขาไม่ต้องมี account

```
persons                          ← identity หลัก (ชื่อ, bio, skill, portfolio ผูกที่นี่)
  id, name, email (nullable)
  linked_user_id (nullable FK → users.id)  ← claim ได้ถ้าสมัครทีหลัง
  created_by_user_id FK → users.id

users                            ← auth account เฉพาะคนที่ login ได้
  id, person_id FK → persons.id
  email, username, status

profiles, skills, experiences,   ← ผูกกับ person_id ทั้งหมด
portfolio_items, social_links
```

**ตอน Register:** ระบบสร้าง `Person` → สร้าง `User` → link กัน → ออก JWT ที่มีทั้ง `user_id` + `person_id`

### Project Management (migrations 003, 006, 007, 008 — มี HTTP routes แล้ว)

```
projects                         ← โครงการ (single-tenant, code = PREFIX-YYYY-NNNN)
  id, project_code, name, status, closure_reason, closed_at,
  client_person_id FK → persons.id, ...

project_members                  ← สมาชิกในโครงการ (flat + is_lead, at-most-one lead)
  project_id, person_id, is_lead, added_by_user_id, created_at, deleted_at
  (joined = created_at; left = deleted_at — ไม่มี role/joined_at/left_at แยก)

transaction_categories           ← master data (12 seeds: 4 income + 8 expense)
  code, name, type (income|expense), icon_key, is_system
  (+ migration 008 เพิ่มหมวดรายจ่าย `penalty` = ค่าปรับ)

project_transactions             ← รายรับ/รายจ่ายต่อ project
  project_id, category_id, amount NUMERIC(15,2), occurred_at, ...

project_financial_plans          ← แผนการเงินระดับโครงการ (1:1 ต่อ project)
  project_id, contract_amount, retention_amount, planned_delivery_date, payment_note

project_payment_installments     ← ตารางงวดรับเงิน
  id, project_id, sort_order, title, amount, planned_delivery_date, planned_receive_date, note
```

Key guarantees (CHECK constraints):

- `project_code` ต้องตรง `^[A-Z]+-[0-9]{4}-[0-9]{4}$`
- ถ้า `status = 'closed'` ต้องมีทั้ง `closed_at` + `closure_reason`
- ถ้า `status ∈ (active, on_hold, closed)` ต้องมี `client_person_id`
- `retention_amount` ต้องไม่ติดลบและต้องไม่เกิน `contract_amount`
- `project_payment_installments.amount` ต้องมากกว่า 0

**User role** ได้ถูกย้ายจาก `user` → `member` พร้อม CHECK constraint (`admin | manager | member`).

---

## API Endpoints

Base URL: `http://localhost:8000`

> `Bearer` = ต้องส่ง `Authorization: Bearer <access_token>` header

---

### Health

| Method | Path            | Auth | Description                       |
| ------ | --------------- | ---- | --------------------------------- |
| GET    | `/health/`      | —    | Overall status                    |
| GET    | `/health/live`  | —    | Liveness probe (process alive)    |
| GET    | `/health/ready` | —    | Readiness probe (DB + Redis ping) |

---

### Auth

| Method | Path                           | Auth   | Description                             |
| ------ | ------------------------------ | ------ | --------------------------------------- |
| POST   | `/api/v1/users/register`       | —      | สมัครสมาชิก + รับ token pair ทันที      |
| POST   | `/api/v1/auth/login`           | —      | Login ด้วย email + password             |
| POST   | `/api/v1/auth/refresh`         | —      | ต่ออายุ access token ด้วย refresh token |
| GET    | `/api/v1/auth/verify-email`    | —      | ยืนยันอีเมล ด้วย `?token=<token>`       |
| POST   | `/api/v1/auth/forgot-password` | —      | ขอ reset password token (ส่งทาง email)  |
| POST   | `/api/v1/auth/reset-password`  | —      | Reset password ด้วย token               |
| POST   | `/api/v1/auth/logout`          | Bearer | ออกจากระบบ                              |
| POST   | `/api/v1/auth/resend-verify`   | Bearer | ขอส่งอีเมลยืนยันอีกครั้ง                |

<details>
<summary>ตัวอย่าง Request Body</summary>

**POST /api/v1/users/register**

```json
{
   "email": "nipon@example.com",
   "password": "MyPassword123!",
   "re_password": "MyPassword123!"
}
```

**POST /api/v1/auth/login**

```json
{ "email": "nipon@example.com", "password": "MyPassword123!" }
```

**POST /api/v1/auth/refresh**

```json
{ "refresh_token": "<refresh_token>" }
```

**GET /api/v1/auth/verify-email**

```
GET /api/v1/auth/verify-email?token=abc123def456...
```

**POST /api/v1/auth/forgot-password**

```json
{ "email": "nipon@example.com" }
```

**POST /api/v1/auth/reset-password**

```json
{ "token": "abc123def456...", "new_password": "NewPassword456!" }
```

</details>

---

### Me (ข้อมูลตัวเอง)

> ทุก route ทำงานบน **person_id** ที่อยู่ใน JWT token (ไม่ใช่ user_id)

| Method | Path                                  | Auth   | Description                                 |
| ------ | ------------------------------------- | ------ | ------------------------------------------- |
| GET    | `/api/v1/me`                          | Bearer | ดูข้อมูล user + person ตัวเอง               |
| PUT    | `/api/v1/me/password`                 | Bearer | เปลี่ยนรหัสผ่าน                             |
| PUT    | `/api/v1/me/profile`                  | Bearer | อัปเดต profile (display name, bio ฯลฯ)      |
| GET    | `/api/v1/me/social-links`             | Bearer | ดูรายการ social links                       |
| PUT    | `/api/v1/me/social-links/:platformID` | Bearer | เพิ่ม/อัปเดต social link (1 ต่อ 1 platform) |
| DELETE | `/api/v1/me/social-links/:linkID`     | Bearer | ลบ social link                              |
| POST   | `/api/v1/me/skills`                   | Bearer | เพิ่ม skill                                 |
| DELETE | `/api/v1/me/skills/:skillID`          | Bearer | ลบ skill                                    |
| POST   | `/api/v1/me/experiences`              | Bearer | เพิ่ม experience                            |
| PUT    | `/api/v1/me/experiences/:expID`       | Bearer | อัปเดต experience                           |
| DELETE | `/api/v1/me/experiences/:expID`       | Bearer | ลบ experience                               |
| POST   | `/api/v1/me/portfolio`                | Bearer | เพิ่ม portfolio item                        |
| PUT    | `/api/v1/me/portfolio/:itemID`        | Bearer | อัปเดต portfolio item                       |
| DELETE | `/api/v1/me/portfolio/:itemID`        | Bearer | ลบ portfolio item                           |

<details>
<summary>ตัวอย่าง Request Body</summary>

**PUT /api/v1/me/password**

```json
{ "current_password": "MyPassword123!", "new_password": "NewPassword456!" }
```

**PUT /api/v1/me/profile**

```json
{
   "display_name": "Nipon K.",
   "headline": "Backend Engineer & IoT Maker",
   "bio": "หลงใหล Go, clean architecture และระบบที่ scale ได้จริง",
   "location": "Bangkok, TH",
   "avatar_url": "https://cdn.example.com/avatar.jpg"
}
```

**PUT /api/v1/me/social-links/:platformID**

```
PUT /api/v1/me/social-links/01957a12-...  (UUID จาก GET /api/v1/social-platforms)
```

```json
{ "url": "https://github.com/lonelyman", "sort_order": 1 }
```

**POST /api/v1/me/skills**

```json
{ "name": "Go", "category": "backend", "proficiency": 5, "sort_order": 1 }
```

> `category`: `backend` | `frontend` | `devops` | `other` — `proficiency`: 1–5

**POST /api/v1/me/experiences**

```json
{
   "company": "Vexentra Studio",
   "position": "Founder & Backend Engineer",
   "location": "Bangkok, TH",
   "description": "ออกแบบและพัฒนา backend ให้กับ SaaS portfolio platform",
   "started_at": "2021-01-01T00:00:00Z",
   "is_current": true
}
```

**POST /api/v1/me/portfolio**

```json
{
   "title": "SmartFarm POC",
   "summary": "ระบบ IoT อัตโนมัติสำหรับฟาร์มไฮโดรโปนิก",
   "description": "ควบคุม ESP32-S3 พร้อม WebUI Dashboard Real-time",
   "content_markdown": "## Overview\n...",
   "cover_image_url": "https://cdn.example.com/smartfarm.jpg",
   "demo_url": "https://demo.example.com/smartfarm",
   "source_url": "https://github.com/lonelyman/smartfarm",
   "status": "published",
   "featured": true,
   "sort_order": 1,
   "tags": ["IoT", "Go", "ESP32"]
}
```

> `status`: `draft` | `published`

</details>

---

### Users (Admin)

| Method | Path            | Auth   | Description                       |
| ------ | --------------- | ------ | --------------------------------- |
| GET    | `/api/v1/users` | Bearer | รายการ users (offset หรือ cursor) |

<details>
<summary>ตัวอย่าง Query Params</summary>

```
GET /api/v1/users?page=1&limit=10
GET /api/v1/users?cursor=01957a12-xxxx&limit=20
```

</details>

---

### Public Profile

> `:id` = **person_id** (UUID จาก `persons` table ไม่ใช่ user_id)

| Method | Path                        | Auth   | Description                               |
| ------ | --------------------------- | ------ | ----------------------------------------- |
| GET    | `/api/v1/showcase`          | —      | Full profile ของ showcase person (public) |
| GET    | `/api/v1/users/:id/profile` | Bearer | Full profile ของ person คนใดก็ได้         |

<details>
<summary>ตัวอย่าง</summary>

```
GET /api/v1/users/01957a12-xxxx-xxxx-xxxx-xxxxxxxxxxxx/profile
```

> `:id` คือ person_id ที่อยู่ใน JWT หรือได้จาก list users

</details>

---

### Social Platforms (Master Data)

| Method | Path                           | Auth   | Description                      |
| ------ | ------------------------------ | ------ | -------------------------------- |
| GET    | `/api/v1/social-platforms`     | —      | รายการ platform ทั้งหมด (public) |
| POST   | `/api/v1/social-platforms`     | Bearer | เพิ่ม platform ใหม่ (admin)      |
| PUT    | `/api/v1/social-platforms/:id` | Bearer | อัปเดต platform (admin)          |
| DELETE | `/api/v1/social-platforms/:id` | Bearer | ลบ platform (admin)              |

<details>
<summary>ตัวอย่าง Request Body</summary>

**POST /api/v1/social-platforms**

```json
{
   "key": "github",
   "name": "GitHub",
   "icon_url": "https://cdn.example.com/icons/github.svg",
   "sort_order": 1,
   "is_active": true
}
```

**PUT /api/v1/social-platforms/:id**

```json
{
   "name": "GitHub",
   "icon_url": "https://cdn.example.com/icons/github-new.svg",
   "sort_order": 1,
   "is_active": true
}
```

</details>

---

## Architecture

```
Transport (HTTP) → Service (Business Logic) → Repository Interface → Adapter (DB)
```

- **Person ≠ User** — Person คือ identity (profile/portfolio ผูกที่นี่), User คือ auth account
- JWT AccessToken มีทั้ง `user_id` (sub) และ `person_id` — handlers ดึง personID จาก claims โดยตรง
- **Domain entity** ไม่มี JSON tag — transport layer จัดการ serialization เพียงที่เดียว
- **Repository interface** อยู่ใน domain module — dependency inversion
- **UUID v7** — time-sorted UUID สำหรับทุก primary key
- **Soft delete** — ทุกตารางมี `deleted_at`; unique indexes ใช้ partial `WHERE deleted_at IS NULL`
- **Schema = SQL migrations** — GORM ไม่มีสิทธิ์แก้โครงสร้างตาราง

---

## Development Workflow

1. สร้าง feature branch จาก `dev`
2. แก้ไข + ทดสอบด้วย `docker compose up --build`
3. Push และเปิด PR เข้า `dev`
4. Merge `dev` → `main` สำหรับ production

---

## License

Private — owned by Vexentra.

---
