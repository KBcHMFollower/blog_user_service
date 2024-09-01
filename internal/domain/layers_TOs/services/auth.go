package services_transfer

type RegisterInfo struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	FName    string `validate:"required,alpha"`
	LName    string `validate:"required,alpha"`
}

type LoginInfo struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

type CheckAuthInfo struct {
	AccessToken string
}

type TokenResult struct {
	AccessToken string
}
