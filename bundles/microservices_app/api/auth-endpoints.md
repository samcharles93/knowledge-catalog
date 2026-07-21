---
type: API
title: Authentication API Endpoints
description: REST endpoints for user authentication.
resource: /api/v1/auth
tags: [api, auth, rest]
timestamp: 2026-07-22T00:00:00Z
---

# Authentication API Endpoints

## `POST /api/v1/auth/login`
Authenticates user credentials and returns JWT bearer token.

### Request Payload
```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```

### Response Payload
```json
{
  "token": "eyJhbGciOiJIUzI1Ni...",
  "expires_in": 3600
}
```

Related: [User Service](/services/user-service.md).
