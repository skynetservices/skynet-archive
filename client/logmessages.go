package client

type FailedConnection struct {
	Err error
}

func (fc FailedConnection) String() string {
	return fc.Err.Error()
}
