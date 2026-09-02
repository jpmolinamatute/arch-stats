---
name: backend-go-coding
description: Standards, idioms, and best practices for writing Go code in the backend following Effective Go
---

# Go Coding Standards & Best Practices

All Go code written in `backend/` must adhere to idiomatic Go as defined by
[Effective Go](https://go.dev/doc/effective_go) and the project architecture standards.

## Mandatory Execution Workflow

Whenever creating or modifying Go code, follow this sequence:

1. **Write idiomatic Go** adhering to the conventions below.
2. **Run linting and formatting** using the `backend-linting` skill:

   ```bash
   cd backend && golangci-lint run ./...
   ```

3. **Run tests** using the `backend-tests` skill:

   ```bash
   cd backend && go test ./... -v
   ```

Do not consider Go code complete until both linting and testing pass with zero errors.

## Effective Go Conventions

### 1. Formatting & Style

- Never argue about formatting: all code must be formatted using `gofumpt` (enforced by `golangci-lint`).
- Use tabs for indentation, spaces for alignment.
- Keep line lengths reasonable, but do not break lines artificially.

### 2. Naming

- **Package names**: Short, clear, lowercase, single-word names (e.g., `config`, `service`, `model`,
  `repository`, `handler`). Do not use underscores or mixedCaps.
- **Getters**: Omit the `Get` prefix. For field `owner`, name the method `Owner()`, not `GetOwner()`.
  Use `SetOwner(...)` for setters.
- **Interface names**: One-method interfaces end in `-er` (e.g., `Reader`, `Writer`, `Closer`). Keep
  interfaces small and define them where they are consumed.
- **MixedCaps**: Use camelCase or PascalCase (`MixedCaps`), never snake_case for Go identifiers.
  Acronyms stay uppercase (`URL`, `HTTP`, `ID`, `JWT`, `DSN`).

### 3. Control Flow

- Avoid unnecessary indentation: handle errors and guard clauses early with fast return.

  ```go
  // Good:
  if err != nil {
      return nil, err
  }
  // proceed with happy path
  ```

- Use short variable declarations in `if` statements when scoping variables:

  ```go
  if err := cfg.Validate(); err != nil {
      return err
  }
  ```

- Use type switches to inspect concrete types from interface values.

### 4. Functions & Returns

- Return multiple values instead of out-parameters or synthetic wrapper structs.
- The standard error return is always the last return value (`(Result, error)`).
- Use `defer` immediately after resource allocation to guarantee cleanup (e.g., rows closing, unlock
  ):

  ```go
  rows, err := pool.Query(ctx, query, args...)
  if err != nil {
      return nil, err
  }
  defer rows.Close()
  ```

### 5. Data & Allocation

- Use composite literals (`&Config{DevMode: true}`) instead of `new()`.
- Use `make` only for slices, maps, and channels where capacity or length is needed.
- Provide constructor functions (`New...`) when struct initialization requires validation, defaults,
  or dependency injection.

### 6. Methods & Interfaces

- Choose pointer receivers when the method modifies the struct or the struct is large; use value
  receivers for small immutable value objects.
- Keep receivers consistent across a type's method set.
- Satisfy interfaces implicitly, do not declare explicit interface implementations.
- Accept interfaces, return concrete structs (e.g., accept `repository.ArcherRepository` interface
  in service constructors, return `*Service`).

### 7. Error Handling

- Errors are values: inspect and handle them explicitly.
- Use `fmt.Errorf("...: %w", err)` to wrap lower-level errors with context.
- Use sentinel errors defined in `internal/apperror` (e.g., `apperror.ErrNotFound`,
  `apperror.ErrUnauthorized`).
- Check wrapped errors using `errors.Is(err, apperror.ErrNotFound)` or `errors.As`.
- **Never panic** in library, repository, service, or handler code. Reserve `panic` strictly for
  truly unrecoverable program initialization failures.

### 8. Concurrency

- "Do not communicate by sharing memory; instead, share memory by communicating."
- Always propagate `context.Context` as the first argument in blocking, I/O, or database calls.
- Always ensure goroutines have a well-defined termination condition to prevent goroutine leaks.
- Use `sync.WaitGroup` or `errgroup.Group` for coordinated concurrency.
