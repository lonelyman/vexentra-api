# Vexentra Business Rules

Last updated: 2026-04-24

## 1) Project Lifecycle

- Status set: `draft`, `planned`, `bidding`, `active`, `on_hold`, `closed`.
- `closed` is terminal and must include closure reason/date.
- Reopen is done by normal update away from `closed` (clears closure markers).

## 2) Client Requirement

- Transition to `active` / `on_hold` requires `client_person_id`.

## 3) Lead Requirement

- Transition to `active` requires an active lead in `project_members`.
- Lead can be set at creation time via `initial_lead_person_id` or later via transfer.

## 4) Financial Lock

- When project is `closed`, transaction and financial write operations are blocked.
- Staff can still read all data.

## 5) Role Model

- Project governance write access is granted to:
  - staff (`admin` / `manager`)
  - current `lead`
  - active `coordinator`
- Normal `member` is read-only.

## 6) Member Role Master Data

- Role definitions live in `project_role_master`.
- Assignments live in `project_member_role_assignments`.
- A member can have multiple roles and optional primary role.

## 7) Audit Semantics

- `created_by_user_id` means who recorded data, not who owns decision power.

## 8) Finance Visibility Policy

- Each project controls finance visibility with:
  - `contract_finance_visibility`
  - `expense_finance_visibility`
- Allowed values:
  - `all_members`
  - `lead_coordinator_staff`
  - `staff_only`
- Contract section (`financial-plan`) is blocked by `contract_finance_visibility`.
- Expense section behavior when denied by `expense_finance_visibility`:
  - list/export excludes expense rows
  - summary returns `expense = 0` and `net = income`
  - direct read/update/delete on an expense transaction returns `403`

## 8) Finance Visibility Per Project

- Project has 2 visibility policies:
  - `contract_finance_visibility`
  - `expense_finance_visibility`
- Allowed values:
  - `all_members`
  - `lead_coordinator_staff`
  - `staff_only`
- Contract endpoints use contract visibility (e.g. financial plan).
- Transaction endpoints use expense visibility.
