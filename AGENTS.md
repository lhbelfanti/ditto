# AI Agents Protocol

Welcome to `ditto`. If you are an AI agent, follow these rules:

## 1. Local Context

This is the shared Go infrastructure library for the Nebula ecosystem's backend services
(database access, HTTP routing/response, migrations, logging, etc.). Generic, reusable backend
behavior belongs here, not duplicated inside a consuming `nebula-*` service.

## 2. English Only

All text, code, documentation, and commit messages MUST be in English.

## 3. Code Conventions

Before every commit, apply this ecosystem's Go conventions from
[`lhbelfanti/skills`](https://github.com/lhbelfanti/skills), `code-conventions/go/`. Do not
restate or copy those rules here — read them from that repo directly, so this file never drifts
out of sync with them.
