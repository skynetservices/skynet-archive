ENV VARS

Here is a list of environmental variables that are inspected by skylib.

DZHOST=localhost:8046
	Where skylib will look for a doozer instance. This list is colon-separated; multiple doozerd instance can be specified.

DZNSHOST=localhost:8046
	The host of a DZNS instance.

DZDISCOVER=true
	Automatically discover new doozerd instances to connect to.

SKYNET_LISTEN=:9999
	The address and port to listen to for the main RPC calls.

SKYNET_ADMIN=:9998
	The address and port to listen to for admin RPC calls.

SKYNET_REGION=unknown
	The service's self-reported region.
