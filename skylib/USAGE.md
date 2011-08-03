# Usage

* SkyNet consists of processes called Nodes.
* A Node which provides at least one Service is an Agent.
* An Agent may offer several Servers on various ports,
  each of which is registered via a Provision.
* A Provision is a class which describes the Services (methods)
  available from any associated Server.
* A Node which has no SkyNet Servers is an Initiator.

`Node.Start()` must be called by all Nodes. It is non-blocking,
though it will start any Servers registered to an Agent.

`Node.Wait()` may be called after `Start()` to wait on all Servers.

See **examples/GetWidgets**.
