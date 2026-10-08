# 0002: Per-home isolation with Postgres row level security

- **Status:** Accepted
- **Date:** 2026-10-06

## Context

`twin_db` is shared by all homes: every structural table (`areas`, `devices`,
`entities`, `twin_relations`, `twin_state`) carries a `home_id` column, and
home membership (`identity_db.home_members`) is decided by the auth
middleware, not by twin-core. Which user may access which home is the auth
service's concern; twin-core only has to guarantee that one home's rows never
leak into another home's operations.

A schema-per-home layout (one Postgres schema per home, generated and
search-path-routed by twin-core) was prototyped on this branch and rejected:
it makes the application generate schema identifiers and run per-tenant DDL,
which is exactly the dynamic-SQL plumbing the group wants to avoid.

## Decision

Keep the shared tables and let Postgres enforce isolation:

- Every home table has a `home_id` column (`entities` and `twin_state` gained
  one), an index on it, `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL
  SECURITY` (so the policy also binds the table owner), and one permissive
  policy `home_isolation`:
  `USING (home_id = current_setting('app.home_id', true)::uuid)`.
- twin-core opens every home transaction with `WithHome` (`internal/db`):
  `BEGIN; SELECT set_config('app.home_id', $1, true)`. The setting is
  transaction-local, so it cannot leak to the next pool user, and an unset
  `app.home_id` makes the policy hide every row and reject every write.
- Routing happens before the secured tables, through two small registry
  tables without RLS: `device_registry` maps `(gateway_id, external_id)` to
  `(home_id, device_id)` for readings, and `node_registry` maps a bare node
  id (area, device, entity, relation) to its home for the id-addressed REST
  endpoints. They hold routing data only.
- `homes` itself carries no RLS: it is the catalog, and which user may see
  which home is decided by the auth service.

## Consequences

**Positive**

- One migration set, one connection pool, one backup and one migration path;
  adding a home is an `INSERT INTO homes`, nothing else.
- Isolation survives application bugs: a query that forgets its home context
  sees zero rows instead of leaking another home's data.
- The repo layer keeps its shape: `WithHome` is one small helper, and the
  cross-home `/state` list merges per-home transactions.

**Negative**

- Policies add a per-row check to every statement; the `home_id` indexes
  keep this cheap at prototype scale.
- Bare-id lookups need a registry hop before the secured tables, and the
  registries are two more tables to keep in sync on create and delete (they
  cascade on `homes` and are cleaned up in the delete paths).
- `PATCH /devices/{id}` can no longer move a device between homes: the
  policy pins `home_id`, and the service rejects the attempt with `400`.

## Alternatives considered

- **Schema per home** (prototyped on this branch, reverted): per-home DDL,
  generated schema names and per-transaction search paths in Go. Rejected
  as dynamic-SQL plumbing with no isolation gain over RLS.
- **Separate database per home:** the strongest isolation, but databases
  would multiply with the number of homes and migrations and backups would
  fan out per home.
- **Application-level `WHERE home_id = ...` filters only:** the state before
  this change. One forgotten filter is a silent cross-home leak; RLS turns
  that class of bug into zero rows.