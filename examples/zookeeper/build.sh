root=/Users/erik/go/src/github.com/skynetservices/skynet/examples
CGO_CFLAGS="-I$root/tools/starter-kit-osx/zookeeper/include" CGO_LDFLAGS="$root/tools/starter-kit-osx/zookeeper/lib/libzookeeper_mt.a" go build
