package skynet

import (
	"labix.org/v2/mgo/bson"
)

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
	ClientID    string
	Method      string
	RequestInfo *RequestInfo
	In          bson.Binary
}

type ServiceRPCOut struct {
	Out       bson.Binary
	ErrString string
}
