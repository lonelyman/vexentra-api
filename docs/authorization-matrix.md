# Vexentra Authorization Matrix

Last updated: 2026-04-24

## Roles in Project Context

- `admin/manager` (staff): full access across all projects.
- `lead`: member row with `is_lead=true`.
- `coordinator`: member with active role assignment `code=coordinator`.
- `member`: active project member without lead/coordinator privilege.

## Effective Policy (Current)

| Module / Action | admin/manager | lead | coordinator | member |
|---|---|---|---|---|
| View project detail/list (if in project scope) | ✅ | ✅ | ✅ | ✅ |
| Update project info/status (non-close) | ✅ | ✅ | ✅ | ❌ |
| Close project | ✅ | ✅ | ✅ | ❌ |
| Delete project | ✅ | ✅ | ✅ | ❌ |
| Manage members (add/remove/update roles) | ✅ | ✅ | ✅ | ❌ |
| Transfer lead | ✅ | ✅ | ✅ | ❌ |
| Financial plan (upsert) | ✅ | ✅ | ✅ | ❌ |
| Financial plan (read) | ✅ | depends on project `contract_finance_visibility` | depends on project `contract_finance_visibility` | depends on project `contract_finance_visibility` |
| Transactions (create/update/delete) | ✅ | ✅ | ✅ | ❌ |
| Transactions (read expense) | ✅ | depends on project `expense_finance_visibility` | depends on project `expense_finance_visibility` | depends on project `expense_finance_visibility` |
| Tasks - list/get | ✅ | ✅ | ✅ | ✅ |
| Tasks - create/update/delete | ✅ | ✅ | ✅ | ❌ |

## Notes

- `member` is intentionally read-only except visibility.
- `created_by_user_id` is audit-only; it does not grant governance permissions.
- Project writes are controlled by service-layer checks, not UI only.
- If expense visibility denies access, expense rows are filtered from list/export, summary hides expense amount, and direct fetch of expense transaction returns 403.
- Finance reads are additionally gated by per-project visibility policies:
  `contract_finance_visibility` and `expense_finance_visibility`.
