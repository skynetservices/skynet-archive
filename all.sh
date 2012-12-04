function build {
  d=`pwd`

  cd $1
  go build && go test
  cd $d
}

# Build / Test skylib
build .

# Build Client
build client

# Build Service
build service

# Build Daemon
build daemon

# Build Pools
build pools

# Test helpers
build skytest

# Build RPC
build rpc/bsonrpc

# Build / Test sky
build cmd/sky

# Build / Test dashboard
build cmd/dashboard

# Build / Test skydaemon
build cmd/skydaemon

# Build / Test examples
build examples/client

build examples/service

build examples/tutorial/client

build examples/tutorial/service

build examples/testing/fibonacci/fibclient

build examples/testing/fibonacci/fibservice

build examples/testing/sleeper/sleepclient

build examples/testing/sleeper/sleepservice

build examples/testing/vagranttests
