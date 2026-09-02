---
name: backend-scripts
description: How to run Go backend scripts and standalone utilities using go run
---

# Running Go Backend Scripts

To run standalone Go utilities or main programs in the `backend/` module:

```bash
cd backend
go run ./cmd/<utility_name> [args...]
```

For single-file scripts or migration tools:

```bash
cd backend
go run ./path/to/script.go
```
