# Login UseCase Contract

## Method: `Execute`

Executes the login flow.

### Signature
```go
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error)
```

### Input
```go
type LoginInput struct {
    Email    string
    Password string
}
```

### Output
```go
type LoginOutput struct {
    AccessToken  string
    RefreshToken string
    User         AdminUserSummary
}

type AdminUserSummary struct {
    ID    string
    Email string
    Name  string
}
```

### Errors
- `ErrInvalidCredentials`: Returned for wrong email or wrong password.
- `ErrUserLocked`: Returned if account is locked.
- `ErrUserInactive`: Returned if account is not active.
- `ErrInternal`: Unexpected errors.
