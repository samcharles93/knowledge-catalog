---
type: Architecture
title: Microservices System Topology
description: Overview of user, payment, and inventory services.
tags: [architecture, topology, microservices]
timestamp: 2026-07-22T00:00:00Z
---

# Microservices System Topology

This system comprises three decoupled microservices:
1. **[User Service](/services/user-service.md)**: Manages authentication, identity, and access tokens.
2. **Payment Service**: Processes transactions and webhooks.
3. **Inventory Service**: Manages catalog items and stock levels.

## Key Principles
- All endpoints must adhere to [Security Rules](/rules/security-rules.md).
- PostgreSQL is the primary relational store (see [Users Table](/database/users-table.md)).
