# Usage

* SkyNet consists of processes called Agents.
* Threads within Agents can be Providers and/or Consumers.
* An Agent may offer several Servers on various ports,
  each of which may provide Services based on one or more Service Signatures.
* A Signature is a class which describes the Operations (methods)
  available from any Server which provides that Service.
* An Agent which has no SkyNet Servers (a pure Consumer) is an Initiator.

`Agent::Start()` must be called by all Agents. It is non-blocking,
though it will start any Servers registered to an Agent.

`Agent::Wait()` may be called after `Start()` to wait on all Servers.

See **examples/GetWidgets**.
