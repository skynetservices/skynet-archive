ENV VARS

Here is a list of environmental variables that are inspected by skylib.

SKYNET_DZHOST=localhost:8046
	Where skylib will look for a doozer instance. This list is colon-separated; multiple doozerd instance can be specified.

SKYNET_DZNSHOST=localhost:8046
	The host of a DZNS instance.

SKYNET_DZDISCOVER=true
	Automatically discover new doozerd instances to connect to.

SKYNET_BIND_IP=127.0.0.1
  IP Address for skynet services to bind to

SKYNET_MIN_PORT=9000
  The start of port range for skynet services to use

SKYNET_MAX_PORT=9999
  The end of port range for skynet services to use

SKYNET_REGION=unknown
	The service's self-reported region.

SKYNET_MGOSERVER=localhost:27017
  The address of mongodb for logging

SKYNET_MGODB=log
  The name of the logging mongo database
