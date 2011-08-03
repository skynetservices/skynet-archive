# Usage

* SkyNet consists of processes called Agents.
* An Agent which provides at least one Service is a Provider.
* An Agent which uses at least on Service is a Consumer.
* An Agent may offer several Servers on various ports,
  each of which may provide Services based on one or more Models.
* A Model is a class which describes the Operations (methods)
  available from any associated Server.
* An Agent which has no SkyNet Servers (a pure Consumer) is an Initiator.

`Agent::Start()` must be called by all Agents. It is non-blocking,
though it will start any Servers registered to an Agent.

`Agent::Wait()` may be called after `Start()` to wait on all Servers.

See **examples/GetWidgets**.
