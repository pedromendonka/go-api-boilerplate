# Adapters

This directory contains adapters that bridge different domains while maintaining clean separation of concerns.

## Purpose

Adapters implement interfaces defined in one domain using implementations from another domain. This pattern:

- **Prevents circular dependencies** between domains
- **Maintains single responsibility** - domains don't need to know about each other
- **Enables testing** - adapters can be easily mocked
- **Follows Dependency Inversion Principle** - high-level modules don't depend on low-level modules

## Current Adapters

### `user_auth_adapter.go`

Implements `auth.UserService` interface using `user.Repository`.

**Interface** (defined in `auth/service/service.go`):
```text
type UserService interface {
    GetUserForAuth(ctx context.Context, email string) (*AuthUser, error)
    CheckPassword(passwordHash, password string) bool
}
```

**Flow**:
```
auth/service → UserService interface → UserAuthAdapter → user/repository
```

## Adding New Adapters

When a domain needs data from another domain:

1. Define an interface in the **calling domain** with only the methods it needs
2. Create an adapter in this directory that implements the interface
3. The adapter can use repositories or services from the **called domain**
4. Wire the adapter in `main.go`

This keeps domains decoupled and follows idiomatic Go patterns.

