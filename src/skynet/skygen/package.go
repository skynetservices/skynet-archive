package main 

const packageTemplate = `package <%PackageName%>

type <%ServiceName%>Request struct {
	YourInputValue string
}

type <%ServiceName%>Response struct {
	YourOutputValue string
	Errors               []string
}
`