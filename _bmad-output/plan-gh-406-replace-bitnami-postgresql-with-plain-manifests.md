---
title: 'Replace Bitnami PostgreSQL Helm chart with plain manifests'
type: 'chore'
ticket: ''
created: '2026-10-22'
status: 'done'
baseline_revision: '8137ec8d35426638c2834aae282cf9233de7a6ae'
route: 'oneshot'
route_source: 'auto'
risk: 'low'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The `deploy-postgresql` Makefile target pulls the Bitnami PostgreSQL Helm chart — the only Bitnami dependency in the project. This is a moving target that can break CI without code changes and is inconsistent with how all other integration services (RabbitMQ, LDAP, Keycloak, Nomad) are deployed via plain manifests.

**Approach:** Replace the Bitnami Helm chart with a minimal set of plain Kubernetes manifests (ConfigMap, Deployment, Service) under `integration/postgresql/`, using the official `postgres` image. Update the Makefile target to use `kubectl apply` instead of `helm`. Delete the now-unused `integration/postgresql-values.yaml`. The Service must remain named `my-postgresql-database` on port 5432 in namespace `test-vault-config-operator` so all existing integration test connection URLs are unchanged.

</frozen-after-approval>

## Implementation Notes

Oneshot: mechanical replacement following the established pattern in `integration/rabbitmq/`, `integration/nomad/`, and `integration/ldap/`. The Bitnami chart effectively produces a Deployment, Service, and init script ConfigMap — those three resources are trivially replicated with plain manifests. No test code changes needed; the Service name, port, credentials, and init SQL stay identical.

**Files created:**
- `integration/postgresql/configmap.yaml` — init SQL (`CREATE USER helloworld ...`) mounted into `/docker-entrypoint-initdb.d`
- `integration/postgresql/deployment.yaml` — single `postgres:16` pod with env vars for password/database, readiness probe via `pg_isready`
- `integration/postgresql/service.yaml` — ClusterIP service named `my-postgresql-database` on port 5432

**Files modified:**
- `Makefile` — `deploy-postgresql` target: removed `helm` prerequisite and Bitnami chart commands; replaced with `kubectl apply -f ./integration/postgresql` pattern matching other services; updated pod label selector from `app.kubernetes.io/instance=postgresql` to `app=postgresql`

**Files deleted:**
- `integration/postgresql-values.yaml` — Bitnami chart values, no longer needed

## Dev Agent Record

- **Model Used:** Claude Opus 4.6
- **Summary of Changes:** Replaced the Bitnami PostgreSQL Helm chart (`bitnami/postgresql`) with plain Kubernetes manifests (ConfigMap, Deployment, Service) under `integration/postgresql/`. Updated `deploy-postgresql` Makefile target to use `kubectl apply` pattern. Deleted `integration/postgresql-values.yaml`.
- **Outcome:** All gates passed — unit tests, integration tests (including all 3 database secret engine tests), and helm chart tests.
- **Notes:** Commit `b5341a8`. The `helmchart-test` gate initially failed due to a stale Kind cluster with cert-manager already installed from a prior run; after tearing down the stale cluster it passed cleanly.

## Plan Change Log

## Review Triage Log

## Verification

**Commands:**
- `GOTOOLCHAIN=auto make integration 2>&1` -- expected: all integration tests pass, including database secret engine config/role/static-role tests
