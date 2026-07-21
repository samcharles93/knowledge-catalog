---
type: Table
title: Users Database Table
description: Primary storage for registered user profiles and password hashes.
resource: postgresql://db/users
tags: [database, table, postgres]
timestamp: 2026-07-22T00:00:00Z
---

# Users Database Table

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | UUID | PRIMARY KEY | Unique user identifier. |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | User login email. |
| `password_hash` | TEXT | NOT NULL | Argon2id password hash. |
| `created_at` | TIMESTAMPTZ | DEFAULT NOW() | Record creation timestamp. |

Used by [User Service](/services/user-service.md).
