---
type: Service
title: User Service
description: Authentication and account management service.
resource: src/services/user
tags: [service, backend, auth]
timestamp: 2026-07-22T00:00:00Z
---

# User Service

Handles user registration, login, JWT issuance, and profile updates.

## Interfaces & Contracts
- Implements [Auth Endpoints](/api/auth-endpoints.md).
- Stores user data in [Users Table](/database/users-table.md).
- Enforces [Security Rules](/rules/security-rules.md).
