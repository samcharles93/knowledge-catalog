---
type: Rule
title: Security Coding Rules
description: Strict security rules for authentication and data storage.
tags: [rules, security, compliance]
timestamp: 2026-07-22T00:00:00Z
---

# Security Coding Rules

1. **Passwords**: Never store passwords in plain text. Use Argon2id with salt.
2. **Tokens**: JWT tokens must expire within 1 hour.
3. **DB Queries**: Never use raw string concatenation in SQL queries. Always use parameterized queries or ORM models.
