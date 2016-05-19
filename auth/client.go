package auth

import (
	"github.com/bketelsen/skynet"
	"github.com/bketelsen/skynet/client"
)

type AuthClient struct {
	*client.ServiceClient
}

func (ac AuthClient) Authenticate(in AuthenticateRequest) (out AuthenticateResponse, err error) {
	err = ac.Send(nil, "Authenticate", in, &out)
	return
}

func GetAuthClientForRegion(c *client.Client, region string) (ac AuthClient) {
	registered := true
	q := &skynet.Query{
		DoozerConn: c.DoozerConn,
		Service:    "Authenticator",
		Region:     region,
		Registered: &registered,
	}
	ac = AuthClient{
		ServiceClient: c.GetServiceFromQuery(q),
	}
	return
}
