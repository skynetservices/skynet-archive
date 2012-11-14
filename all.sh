# Build / Test skylib
go build && go test

# Build Client
cd client && go build && go test
cd ..

# Build Service
cd service && go build && go test
cd ..

# Build Daemon
cd daemon && go build && go test
cd ..

# Build Pools
cd pools && go build && go test
cd ..

# Build RPC
cd rpc/bsonrpc && go build && go test
cd ../../

# Build / Test sky
cd cmd/sky && go build && go test
cd ..

# Build / Test dashboard
cd dashboard && go build && go test
cd ..

# Build / Test skydaemon
cd skydaemon && go build && go test
cd ../../

# Build / Test examples
cd examples/client && go build && go test
cd ../

cd service && go build && go test
cd ../

cd tutorial/client && go build && go test
cd ../

cd service && go build && go test
cd ../../

cd testing/fibonacci/fibclient && go build && go test
cd ../

cd fibservice && go build && go test
cd ../../

cd sleeper/sleepclient && go build && go test
cd ../

cd sleepservice && go build && go test
cd ../../

cd vagranttests && go build && go test
cd ../../../
