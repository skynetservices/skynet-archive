package auth

type AuthenticateRequest struct {
	Username string
	Password string
}

type AuthenticateResponse struct {
	Ok bool
}
