package skynet

type RegisterRequest struct {
}

type RegisterResponse struct {
}

type UnregisterRequest struct {
}

type UnregisterResponse struct {
}

type StopRequest struct {
	WaitForClients bool
}

type StopResponse struct {
}

type ServiceRPCIn struct {
	Method      string
	RequestInfo *RequestInfo
	In          []byte
}

type ServiceRPCOut struct {
	Out       []byte
	ErrString string
}
