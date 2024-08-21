package services_transfer

type RegisterInfo struct {
	Email    string
	Password string
	FName    string
	LName    string
}

type LoginInfo struct {
	Email    string
	Password string
}

type CheckAuthInfo struct {
	AccessToken string
}

type TokenResult struct {
	AccessToken string
}
