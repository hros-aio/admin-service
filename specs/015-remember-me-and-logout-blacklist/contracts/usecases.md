# Use Case Contracts: Remember Me and Logout Blacklist

This document defines the application layer usecase signatures updated or introduced in TSK-AUTH-015.

## LoginUseCase

### Execute Signature
```go
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error)
```

### Input Struct
```go
type LoginInput struct {
	Email      string
	Password   string
	RememberMe bool // Added
	IPAddress  string
	UserAgent  string
}
```

### Output Struct
```go
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	User         AdminUserSummary
}
```

---

## LogoutUseCase

### Execute Signature
```go
func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error
```

### Input Struct
```go
type LogoutInput struct {
	RefreshToken string
	AccessToken  string // Added
}
```
