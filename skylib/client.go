package skylib

import (
  "github.com/erikstmartin/msgpack-rpc/go/rpc"
  "net"
  "log"
  "os"
  "strconv"
  "reflect"
)

type Client struct {
	DoozerConn *DoozerConnection
  Config *ClientConfig
	Log *log.Logger `json:"-"`
}

func (c *Client) doozer() *DoozerConnection {
	if c.DoozerConn == nil {
		c.DoozerConn = &DoozerConnection {
			Config:  c.Config.DoozerConfig,
		}

		c.DoozerConn.Connect()
	}

	return c.DoozerConn
}

func NewClient(config *ClientConfig) *Client {
	if config.Log == nil {
		config.Log = log.New(os.Stderr, "", log.LstdFlags)
	}

  client := &Client {
    Config: config,
    DoozerConn: &DoozerConnection {
      Config:  config.DoozerConfig,
      Log: config.Log,
    },
    Log: config.Log,
  }

  client.DoozerConn.Connect()

	return client
}


// TODO: For now this will return a single connection to this service, each time it's called
// refactor this so that it returns back a connection pool
// (should this return back the same pool un subsequent calls? or a new pool?)
func (c *Client) GetService(name string, version string, region string, host string) (*ServiceClient){
  var conn net.Conn
  var err error

  query := &Query{
		DoozerConn: c.DoozerConn,
		Service:    name,
		Version:    version,
		Host:       host,
		Region:     region,
	}

	results := query.FindInstances()
  ok := false

  // Connect to the first instance we find available
	for _, instance := range *results {
		//fmt.Println(instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port) + " - " + instance.Config.Name + " (" + instance.Config.Version + ")")
    conn, err = net.Dial("tcp", instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port))

    if err == nil {
      ok = true
      break
    }
	}

  if !ok {
    c.Log.Panic("Unable to find available instance for Service: " + name)
  }

  client := rpc.NewSession(conn, true)

  return &ServiceClient {
    conn: client,
  }
}

type ServiceClient struct {
  conn *rpc.Session
}

func (c *ServiceClient) Send(funcName string, arguments ...interface{}) (reflect.Value, error) {
  // TODO: do logging here, average response time, number of calls etc
  // TODO: timeout logic
  
  return c.conn.SendV(funcName, arguments)
}

