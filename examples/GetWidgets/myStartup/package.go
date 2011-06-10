package myStartup

type GetUserDataRequest struct {
	YourInputValue string
}

type GetUserDataResponse struct {
	YourOutputValue string
	Errors               []string
}
