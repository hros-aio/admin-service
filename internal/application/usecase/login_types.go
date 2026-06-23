package usecase

// LoginInput represents the input for the login use case.
type LoginInput struct {
	Email      string
	Password   string
	RememberMe bool
	IPAddress  string
	UserAgent  string
}

// LoginOutput represents the output of the login use case.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	User         AdminUserSummary
	MFARequired  bool
	MFAToken     string
	MFAMethods   []string
}

// AdminUserSummary represents basic admin user details.
type AdminUserSummary struct {
	ID    string
	Email string
	Name  string
}
