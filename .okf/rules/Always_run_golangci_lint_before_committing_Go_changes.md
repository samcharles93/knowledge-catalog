---
tags:
    - go
    - ci
timestamp: "2026-07-21T18:18:12Z"
title: Always run golangci-lint before committing Go changes
type: Rule
---

Run `golangci-lint run ./...` (or `task check`) before committing any Go change in this repo, so lint issues surface before they hit CI.
