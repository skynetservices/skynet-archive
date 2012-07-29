# Build / Test skylib
go build && go test

# Build / Test sky
cd sky && go build && go test
cd ..

# Build / Test examples
cd examples/client && go build && go test
cd ../../

cd examples/service && go build && go test
cd ../../
